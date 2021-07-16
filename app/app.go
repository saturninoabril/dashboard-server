package app

import (
	"html/template"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/saturninoabril/dashboard-server/store"
	"github.com/saturninoabril/dashboard-server/utils"
	"github.com/sirupsen/logrus"
)

// App contains all the business logic as methods.
type App struct {
	config        Config
	store         store.Store
	user          UserService
	htmlTemplates *template.Template
	logger        logrus.FieldLogger
}

// NewApp creates a new instance of App
func NewApp(logger logrus.FieldLogger, store store.Store, config Config, userService UserService) *App {
	return &App{
		config: config,
		store:  store,
		user:   userService,
		logger: logger,
	}
}

// Clone creates a shallow copy of app
func (a *App) Clone() *App {
	return &App{
		config:        a.Config(),
		store:         a.Store(),
		user:          a.User(),
		htmlTemplates: a.HTMLTemplates(),
		logger:        a.Logger(),
	}
}

func (a *App) Logger() logrus.FieldLogger {
	return a.logger
}

// Config is an accessor for the app config.
func (a *App) Config() Config {
	return a.config
}

// Store is an accessor for the app store.
func (a *App) Store() store.Store {
	return a.store
}

// User is an accessor for the user service.
func (a *App) User() UserService {
	return a.user
}

// ReloadHTMLTemplates refreshes the in-memory HTML templates.
func (a *App) ReloadHTMLTemplates() error {
	templatesDir, ok := utils.GetTemplateDirectory()
	if !ok {
		return errors.New("unable to find templates directory")
	}
	t, err := template.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return errors.Wrap(err, "unable to reload templates")
	}
	a.htmlTemplates = t

	return nil
}

// HTMLTemplates is an accessor for the app html templates.
func (a *App) HTMLTemplates() *template.Template {
	return a.htmlTemplates
}
