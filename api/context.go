package api

import (
	"encoding/json"
	"net/http"

	"github.com/saturninoabril/dashboard-server/app"
	"github.com/saturninoabril/dashboard-server/model"
	"github.com/sirupsen/logrus"
)

// Context provides the API with all necessary data and interfaces for responding to requests.
type Context struct {
	RequestID string
	Logger    logrus.FieldLogger
	App       *app.App
	Session   *model.Session
}

// Clone creates a shallow copy of context, allowing clones to apply per-request changes.
func (c *Context) Clone() *Context {
	return &Context{
		Logger: c.Logger,
		App:    c.App.Clone(),
	}
}

func (c *Context) writeAndLogErrorWithFields(w http.ResponseWriter, err error, logFields logrus.Fields) {
	logger := c.Logger
	if logFields != nil {
		logger = logger.WithFields(logFields)
	}

	logger.Error(err.Error())

	b, _ := json.Marshal(&model.APIError{Message: err.Error()})
	w.Write(b)
}

func (c *Context) writeAndLogError(w http.ResponseWriter, err error) {
	c.writeAndLogErrorWithFields(w, err, nil)
}
