package main

import (
	"github.com/spf13/cobra"

	"github.com/saturninoabril/dashboard-server/store"
)

func cmdStore(command *cobra.Command) (*store.SqlStore, error) {
	database, err := command.Flags().GetString("database")
	if err != nil {
		return nil, err
	}

	tablePrefix, _ := command.Flags().GetString("table-prefix")
	if err != nil {
		return nil, err
	}

	newStore, err := store.New(database, tablePrefix, logger)
	if err != nil {
		return nil, err
	}

	return newStore, nil
}
