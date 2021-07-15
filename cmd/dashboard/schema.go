package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	viper.SetEnvPrefix("DASHBOARD")
	viper.AutomaticEnv()

	schemaCmd.AddCommand(schemaMigrateCmd)
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Generate a schema used by the dashboard server.",
}

var schemaMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the schema to the latest supported version.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		_, err := cmdStore(command)
		if err != nil {
			return err
		}

		return nil
	},
}
