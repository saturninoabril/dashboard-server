package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/saturninoabril/dashboard-server/api"
	"github.com/saturninoabril/dashboard-server/app"
	"github.com/saturninoabril/dashboard-server/model"
	"github.com/saturninoabril/dashboard-server/store"
)

var instanceID string

func init() {
	viper.SetEnvPrefix("DASHBOARD")
	viper.AutomaticEnv()

	instanceID = model.NewID()

	serverCmd.PersistentFlags().String("listen", ":8085", "The interface and port on which to listen on the API.")
	serverCmd.PersistentFlags().Bool("debug", false, "Whether to output debug logs.")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the dashboard server.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		debug, _ := command.Flags().GetBool("debug")
		if debug {
			logger.SetLevel(logrus.DebugLevel)
		}

		logger := logger.WithField("instance", instanceID)

		var config app.Config
		config.SiteURL, _ = command.Flags().GetString("siteurl")
		config.APIURL, _ = command.Flags().GetString("apiurl")
		if config.APIURL == "" {
			config.APIURL = config.SiteURL
		}
		config.Email.SMTPUsername, _ = command.Flags().GetString("smtp-username")
		config.Email.SMTPPassword, _ = command.Flags().GetString("smtp-password")
		config.Email.SMTPServer, _ = command.Flags().GetString("smtp-server")
		config.Email.SMTPPort, _ = command.Flags().GetString("smtp-port")
		config.Email.SMTPServerTimeout, _ = command.Flags().GetInt("smtp-servertimeout")

		dev, _ := command.Flags().GetBool("dev")
		if dev {
			app.SetDevConfig(&config)
			logger.Debug("Using dev configuration")
		}

		database, _ := command.Flags().GetString("database")
		tablePrefix, _ := command.Flags().GetString("table-prefix")

		store, err := store.New(database, tablePrefix, logger)
		if err != nil {
			return err
		}

		userService := app.NewUserService(logger, store)

		app := app.NewApp(logger, store, config, userService)
		err = app.ReloadHTMLTemplates()
		if err != nil {
			logger.WithError(err).Warn("Unable to load HTML templates")
		}

		listen, _ := command.Flags().GetString("listen")

		publicRouter := mux.NewRouter()

		api.Register(publicRouter, &api.Context{
			Logger: logger,
			App:    app,
		})

		startServer := func(router *mux.Router, listen string) *http.Server {
			srv := &http.Server{
				Addr:           listen,
				Handler:        router,
				ReadTimeout:    180 * time.Second,
				WriteTimeout:   180 * time.Second,
				IdleTimeout:    time.Second * 180,
				MaxHeaderBytes: 1 << 20,
				ErrorLog:       log.New(&logrusWriter{logger}, "", 0),
			}

			go func() {
				logger.WithField("addr", srv.Addr).Info("Listening")
				err := srv.ListenAndServe()
				if err != nil && err != http.ErrServerClosed {
					logger.WithError(err).Error("Failed to listen and serve")
				}
			}()
			return srv
		}

		publicServer := startServer(publicRouter, listen)

		c := make(chan os.Signal, 1)
		// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
		// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
		signal.Notify(c, os.Interrupt)

		// Block until we receive our signal.
		<-c
		logger.Info("Shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		publicServer.Shutdown(ctx)

		return nil
	},
}
