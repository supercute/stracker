package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"github.com/supercute/stracker/internal/commands"
	"github.com/supercute/stracker/internal/db"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	database, err := db.Open()
	if err != nil {
		return err
	}
	defer database.Close()

	root := &cobra.Command{
		Use:   "st",
		Short: "Simple time tracker",
	}

	root.AddCommand(
		commands.NewStartCmd(database),
		commands.NewStopCmd(database),
		commands.NewSumCmd(database),
		commands.NewStatusCmd(database),
		commands.NewPauseCmd(database),
		commands.NewDeleteCmd(database),
	)

	return fang.Execute(context.Background(), root)
}
