package email

import (
	"context"
	"crypto/tls"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jaytaylor/html2text"
	"github.com/pkg/errors"
	gomail "gopkg.in/mail.v2"
)

const (
	connSecurityStartTTLS   = "STARTTLS"
	connSecurityTLS         = "TLS"
	sendEmailDefaultRetries = 3
	sendEmailRetryDuration  = 2 * time.Second
)

// Config contains SMTP settings for sending emails.
type Config struct {
	ReplyToName              string
	ReplyToAddress           string
	BCCAddresses             []string
	SMTPUsername             string
	SMTPPassword             string
	SMTPServer               string
	SMTPPort                 string
	SMTPConnectionSecurity   string
	SMTPServerTimeout        int
	SMTPSkipCertVerification bool
}

type Attachment struct {
	Name     string
	MimeType string
	Data     io.Reader
}

type mailData struct {
	mimeTo        string
	smtpTo        string
	from          mail.Address
	replyTo       mail.Address
	bcc           []mail.Address
	subject       string
	htmlBody      string
	attachments   []*Attachment
	embeddedFiles map[string]io.Reader
	mimeHeaders   map[string]string
}

// smtpClient is implemented by an smtp.Client.
// See https://golang.org/pkg/net/smtp/#Client.
type smtpClient interface {
	Mail(string) error
	Rcpt(string) error
	Data() (io.WriteCloser, error)
}

func encodeRFC2047Word(s string) string {
	return mime.BEncoding.Encode("utf-8", s)
}

// SMTPConnectionInfo contains connection information for a SMTP server.
type SMTPConnectionInfo struct {
	SMTPUsername         string
	SMTPPassword         string
	SMTPServerName       string
	SMTPServerHost       string
	SMTPPort             string
	SMTPServerTimeout    int
	SkipCertVerification bool
	ConnectionSecurity   string
	Auth                 bool
}

type authChooser struct {
	smtp.Auth
	connectionInfo *SMTPConnectionInfo
}

func (a *authChooser) Start(server *smtp.ServerInfo) (string, []byte, error) {
	smtpAddress := a.connectionInfo.SMTPServerName + ":" + a.connectionInfo.SMTPPort
	a.Auth = newLoginAuth(a.connectionInfo.SMTPUsername, a.connectionInfo.SMTPPassword, smtpAddress)
	for _, method := range server.Auth {
		if method == "PLAIN" {
			a.Auth = smtp.PlainAuth("", a.connectionInfo.SMTPUsername, a.connectionInfo.SMTPPassword, a.connectionInfo.SMTPServerName+":"+a.connectionInfo.SMTPPort)
			break
		}
	}
	return a.Auth.Start(server)
}

type loginAuth struct {
	username, password, host string
}

func newLoginAuth(username, password, host string) smtp.Auth {
	return &loginAuth{username, password, host}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		return "", nil, errors.New("unencrypted connection")
	}

	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
	}

	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unknown fromServer")
		}
	}
	return nil, nil
}

// ConnectToSMTPServerAdvanced provides advanced SMTP server connection handling.
func ConnectToSMTPServerAdvanced(connectionInfo *SMTPConnectionInfo) (net.Conn, error) {
	var conn net.Conn
	var err error

	smtpAddress := connectionInfo.SMTPServerHost + ":" + connectionInfo.SMTPPort
	dialer := &net.Dialer{
		Timeout: time.Duration(connectionInfo.SMTPServerTimeout) * time.Second,
	}

	if connectionInfo.ConnectionSecurity == connSecurityTLS {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: connectionInfo.SkipCertVerification,
			ServerName:         connectionInfo.SMTPServerName,
		}

		conn, err = tls.DialWithDialer(dialer, "tcp", smtpAddress, tlsconfig)
		if err != nil {
			return nil, errors.Wrap(err, "utils.mail.server.connect_smtp.open_tls.app_error")
		}
	} else {
		conn, err = dialer.Dial("tcp", smtpAddress)
		if err != nil {
			return nil, errors.Wrap(err, "utils.mail.connect_smtp.open.app_error")
		}
	}

	return conn, nil
}

// ConnectToSMTPServer connects to an SMTP server.
func ConnectToSMTPServer(config *Config) (net.Conn, error) {
	return ConnectToSMTPServerAdvanced(
		&SMTPConnectionInfo{
			ConnectionSecurity:   config.SMTPConnectionSecurity,
			SkipCertVerification: config.SMTPSkipCertVerification,
			SMTPServerName:       config.SMTPServer,
			SMTPServerHost:       config.SMTPServer,
			SMTPPort:             config.SMTPPort,
			SMTPServerTimeout:    config.SMTPServerTimeout,
		},
	)
}

// NewSMTPClientAdvanced provides an SMTP client.
func NewSMTPClientAdvanced(ctx context.Context, conn net.Conn, hostname string, connectionInfo *SMTPConnectionInfo) (*smtp.Client, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var c *smtp.Client
	ec := make(chan error)
	go func() {
		var err error
		c, err = smtp.NewClient(conn, connectionInfo.SMTPServerName+":"+connectionInfo.SMTPPort)
		if err != nil {
			ec <- err
			return
		}
		cancel()
	}()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		if err != nil && err.Error() != "context canceled" {
			return nil, errors.Wrap(err, "utils.mail.connect_smtp.open_tls.app_error")
		}
	case err := <-ec:
		return nil, errors.Wrap(err, "utils.mail.connect_smtp.open_tls.app_error")
	}

	if hostname != "" {
		err := c.Hello(hostname)
		if err != nil {
			return nil, errors.Wrap(err, "utils.mail.connect_smtp.helo.app_error")
		}
	}

	if connectionInfo.ConnectionSecurity == connSecurityStartTTLS {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: connectionInfo.SkipCertVerification,
			ServerName:         connectionInfo.SMTPServerName,
		}
		c.StartTLS(tlsconfig)
	}

	if connectionInfo.Auth {
		if err := c.Auth(&authChooser{connectionInfo: connectionInfo}); err != nil {
			return nil, errors.Wrap(err, "utils.mail.new_client.auth.app_error")
		}
	}
	return c, nil
}

// NewSMTPClient returns a new SMTP client.
func NewSMTPClient(ctx context.Context, conn net.Conn, config *Config) (*smtp.Client, error) {
	return NewSMTPClientAdvanced(
		ctx,
		conn,
		"dashboard",
		&SMTPConnectionInfo{
			ConnectionSecurity:   config.SMTPConnectionSecurity,
			SkipCertVerification: config.SMTPSkipCertVerification,
			SMTPServerName:       config.SMTPServer,
			SMTPServerHost:       config.SMTPServer,
			SMTPPort:             config.SMTPPort,
			SMTPServerTimeout:    config.SMTPServerTimeout,
			Auth:                 config.SMTPUsername != "",
			SMTPUsername:         config.SMTPUsername,
			SMTPPassword:         config.SMTPPassword,
		},
	)
}

// TestConnection tests the connection to the SMTP server.
func TestConnection(config *Config) error {
	conn, err := ConnectToSMTPServer(config)
	if err != nil {
		return errors.Wrap(err, "unable to connect to SMTP server, check SMTP server settings")
	}
	defer conn.Close()

	sec := config.SMTPServerTimeout

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(sec)*time.Second)
	defer cancel()

	c, err := NewSMTPClient(ctx, conn, config)
	if err != nil {
		return errors.Wrap(err, "unable to connect to SMTP server, check SMTP server settings")
	}
	c.Close()
	c.Quit()

	return nil
}

// SendMailWithEmbeddedFilesUsingConfig sends an email with file attachments.
func SendMailWithEmbeddedFilesUsingConfig(to, subject, htmlBody string, sendBcc bool, embeddedFiles map[string]io.Reader, attachments []*Attachment, config *Config) error {
	fromMail := mail.Address{Name: config.ReplyToName, Address: config.ReplyToAddress}
	replyTo := mail.Address{Name: config.ReplyToName, Address: config.ReplyToAddress}

	mailData := mailData{
		mimeTo:        to,
		smtpTo:        to,
		from:          fromMail,
		replyTo:       replyTo,
		subject:       subject,
		htmlBody:      htmlBody,
		embeddedFiles: embeddedFiles,
		attachments:   attachments,
	}

	if sendBcc {
		for _, address := range config.BCCAddresses {
			mailData.bcc = append(mailData.bcc, mail.Address{Address: address})
		}
	}

	return sendMailUsingConfigAdvanced(mailData, config, sendEmailDefaultRetries)
}

// SendMailUsingConfig sends an email with the provided config.
func SendMailUsingConfig(to, subject, htmlBody string, sendBcc bool, attachments []*Attachment, config *Config) error {
	return SendMailWithEmbeddedFilesUsingConfig(to, subject, htmlBody, sendBcc, nil, attachments, config)
}

// sendMailUsingConfigAdvanced allows for sending an email with attachments and
// differing MIME/SMTP recipients.
func sendMailUsingConfigAdvanced(mail mailData, config *Config, retries uint64) error {
	if len(config.SMTPServer) == 0 {
		return errors.New("no SMTP server configured")
	}

	conn, err := ConnectToSMTPServer(config)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.SMTPServerTimeout)*time.Second)
	defer cancel()

	c, err := NewSMTPClient(ctx, conn, config)
	if err != nil {
		return err
	}
	defer c.Quit()
	defer c.Close()

	sendEmailOperation := func() error {
		return SendMail(c, mail, time.Now())
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = sendEmailRetryDuration
	backoffRetries := backoff.WithMaxRetries(bo, retries)
	if err := backoff.Retry(sendEmailOperation, backoffRetries); err != nil {
		return err
	}

	return nil
}

// SendMail sends an email.
func SendMail(c smtpClient, mail mailData, date time.Time) error {
	htmlMessage := "\r\n<html><body>" + mail.htmlBody + "</body></html>"

	txtBody, err := html2text.FromString(mail.htmlBody)
	if err != nil {
		return errors.Wrap(err, "failed to convert email body to html text")
	}

	headers := map[string][]string{
		"From":                      {mail.from.String()},
		"To":                        {mail.mimeTo},
		"Subject":                   {encodeRFC2047Word(mail.subject)},
		"Content-Transfer-Encoding": {"8bit"},
		"Auto-Submitted":            {"auto-generated"},
		"Precedence":                {"bulk"},
	}

	if len(mail.replyTo.Address) > 0 {
		headers["Reply-To"] = []string{mail.replyTo.String()}
	}

	for k, v := range mail.mimeHeaders {
		headers[k] = []string{encodeRFC2047Word(v)}
	}

	m := gomail.NewMessage(gomail.SetCharset("UTF-8"))
	m.SetHeaders(headers)
	m.SetDateHeader("Date", date)
	m.SetBody("text/plain", txtBody)
	m.AddAlternative("text/html", htmlMessage)

	for name, reader := range mail.embeddedFiles {
		m.EmbedReader(name, reader)
	}

	for _, attachment := range mail.attachments {
		m.AttachReader(attachment.Name, attachment.Data)
	}

	if err = c.Mail(mail.from.Address); err != nil {
		return errors.Wrapf(err, "failed to add from email address %s", mail.from.Address)
	}
	if err = c.Rcpt(mail.smtpTo); err != nil {
		return errors.Wrapf(err, "failed to add to email address %s", mail.smtpTo)
	}
	for _, bcc := range mail.bcc {
		if err = c.Rcpt(bcc.Address); err != nil {
			return errors.Wrapf(err, "failed to add bcc address %s", bcc.Address)
		}
	}

	w, err := c.Data()
	if err != nil {
		return errors.Wrap(err, "utils.mail.send_mail.msg_data.app_error")
	}

	if _, err = m.WriteTo(w); err != nil {
		return errors.Wrap(err, "utils.mail.send_mail.msg.app_error")
	}
	if err = w.Close(); err != nil {
		return errors.Wrap(err, "utils.mail.send_mail.close.app_error")
	}

	return nil
}
