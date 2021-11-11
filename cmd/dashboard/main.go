package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultDatabase    = "postgres://dashboarduser:dashboardpwd@localhost:5433/dashboard_dev?sslmode=disable"
	defaultTablePrefix = "dashboard_"
)

var rootCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Dashboard server runs the backend for Test Automation Dashboard.",
	Run: func(cmd *cobra.Command, args []string) {
		_ = serverCmd.RunE(cmd, args)
	},
	// SilenceErrors allows us to explicitly log the error returned from rootCmd below.
	SilenceErrors: true,
}

func init() {
	database, ok := os.LookupEnv("DASHBOARD_DATABASE")
	if !ok {
		fmt.Printf("\nUsing default database: '%s'", defaultDatabase)
		database = defaultDatabase
	}

	tablePrefix, ok := os.LookupEnv("DASHBOARD_TABLE_PREFIX")
	if !ok {
		tablePrefix = defaultTablePrefix
	}

	rootCmd.PersistentFlags().String("database", database, "The database backing the dashboard.")
	rootCmd.PersistentFlags().String("table-prefix", tablePrefix, "Table prefix of the database.")
	rootCmd.PersistentFlags().Bool("dev", false, "Set to run in dev mode and configures basic settings if not provided.")
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(userCmd)
}

func main() {
	viper.SetEnvPrefix("DASHBOARD")
	viper.AutomaticEnv()
	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Error("command failed")
		os.Exit(1)
	}
}
