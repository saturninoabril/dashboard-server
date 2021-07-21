package app

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/url"

	"github.com/pkg/errors"
	"github.com/saturninoabril/dashboard-server/internal/email"
	"github.com/saturninoabril/dashboard-server/model"
)

const (
	mimeTypeTextPlain = "text/plain"
)

// SendMail sends mail.
func (a *App) SendMail(to, subject, htmlBody string, sendBcc bool) error {
	emailConfig := a.Config().Email
	return email.SendMailUsingConfig(to, subject, htmlBody, sendBcc, nil, &emailConfig)
}

// SendEmailWithAttachments would send the email including the passed attachments
func (a *App) SendMailWithAttachments(to, subject, htmlBody string, sendBcc bool, attachments []*email.Attachment) error {
	emailConfig := a.Config().Email
	return email.SendMailUsingConfig(to, subject, htmlBody, sendBcc, attachments, &emailConfig)
}

// GetHTMLTemplate returns the HTMLTemplate of a give name.
func (a *App) GetHTMLTemplate(templateName string) *HTMLTemplate {
	return &HTMLTemplate{
		Template:     a.HTMLTemplates(),
		TemplateName: templateName,
		Props:        make(map[string]interface{}),
	}
}

// SendTestEmail creates and sends a test email.
func (a *App) SendTestEmail(userEmail, siteURL string) error {
	bodyPage := a.GetHTMLTemplate("test_email_body")

	bodyPage.Props["Title"] = "Test Email"
	bodyPage.Props["Info"] = "This is a test email generated by the dashboard server."
	bodyPage.Props["Organization"] = "Automation Dashboard"

	renderedBody, err := bodyPage.Render()
	if err != nil {
		return errors.Wrap(err, "unable to render test email")
	}

	err = a.SendMail(userEmail, "Test Email", renderedBody, false)
	if err != nil {
		return errors.Wrap(err, "unable to send test email")
	}

	return nil
}

// SendVerifyEmailEmail sends a verify-email email.
func (a *App) SendVerifyEmailEmail(email, siteURL string, token *model.Token) error {
	subject := "Verify Email"

	bodyPage := a.GetHTMLTemplate("verify_email_body")
	bodyPage.SetBaseProps()
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = "Verify your email address"
	bodyPage.Props["Info"] = "Enter the code below into the browser window where you began creating your Dashboard account."
	bodyPage.Props["Footer"] = "This email address was used to create an account with the Dashboard. \nIf it was not you, you can safely ignore this email."
	bodyPage.Props["token"] = fmt.Sprintf("%s %s", token.Token[:3], token.Token[3:])

	if a.Config().Dev {
		a.logger.Debugf("Verification code for %s: %s", email, token.Token)
	}

	renderedBody, err := bodyPage.Render()
	if err != nil {
		return errors.Wrap(err, "unable to render verify email body")
	}

	err = a.SendMail(email, subject, renderedBody, false)
	if err != nil {
		return errors.Wrap(err, "unable to send verify-email email")
	}

	return nil
}

// SendPasswordResetEmail sends a password reset email.
func (a *App) SendPasswordResetEmail(email, siteURL string, token *model.Token) error {
	subject := "Password Reset"

	bodyPage := a.GetHTMLTemplate("password_reset_body")
	bodyPage.SetBaseProps()
	bodyPage.Props["SiteURL"] = siteURL
	bodyPage.Props["Title"] = "Reset Your Password"
	bodyPage.Props["Info1"] = "Click the button below to reset your password. If you didn’t request this, you can safely ignore this email."
	bodyPage.Props["ResetUrl"] = fmt.Sprintf("%s/reset-password?token=%s", siteURL, url.QueryEscape(token.Token))
	bodyPage.Props["Button"] = "Reset Password"

	renderedBody, err := bodyPage.Render()
	if err != nil {
		return errors.Wrap(err, "unable to render reset password email")
	}

	err = a.SendMail(email, subject, renderedBody, false)
	if err != nil {
		return errors.Wrap(err, "unable to send reset password email")
	}

	return nil
}

// HTMLTemplate is a wrapper for specifying and rendering a given HTML template.
type HTMLTemplate struct {
	Template     *template.Template
	TemplateName string
	Props        map[string]interface{}
}

// SetBaseProps sets common prop values for most email templates.
func (t *HTMLTemplate) SetBaseProps() {
	t.Props["Footer"] = "© 2021 Test Automation Dashboard"
}

// Render renders the HTMLTemplate to a string.
func (t *HTMLTemplate) Render() (string, error) {
	var text bytes.Buffer
	err := t.RenderToWriter(&text)
	if err != nil {
		return "", errors.Wrap(err, "unable to render template")
	}

	return text.String(), nil
}

// RenderToWriter renders the template of the given name to the provided reader.
func (t *HTMLTemplate) RenderToWriter(w io.Writer) error {
	if t.Template == nil {
		return errors.New("no template found")
	}

	err := t.Template.ExecuteTemplate(w, t.TemplateName, t)
	if err != nil {
		return errors.Wrap(err, "unable to execute template")
	}

	return nil
}
