package app

import "github.com/saturninoabril/dashboard-server/internal/email"

// Config is the config used by the dashboard server app.
type Config struct {
	// the location to which a user might point their browser
	SiteURL string

	// the location to which the API should be called if is different than SiteURL
	APIURL string

	// email server related configuration
	Email email.Config

	// developer mode
	Dev bool
}

// NewConfig returns a new empty config.
func NewConfig() Config {
	return Config{}
}

// SetDevConfig create config for local dev mode.
func SetDevConfig(config *Config) {
	config.Dev = true

	if len(config.SiteURL) == 0 {
		config.SiteURL = "http://localhost:3000"
	}
	if len(config.Email.SMTPServer) == 0 {
		config.Email.SMTPServer = "localhost"
	}
	if len(config.Email.SMTPPort) == 0 {
		config.Email.SMTPPort = "12025"
	}
	if config.Email.SMTPServerTimeout == 0 {
		config.Email.SMTPServerTimeout = 10
	}
}
