package commands

import (
	"errors"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/supercute/stracker/internal/db"
)

func NewPauseCmd(database *db.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pause index",
		Aliases: []string{"p"},
		Short:   "Pause active task by index",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idx, err := strconv.Atoi(args[0])
			if err != nil || idx <= 0 {
				cmd.Println("Invalid index:", args[0])
				return nil
			}

			entry, err := database.PauseByIndex(idx)
			if err != nil {
				if errors.Is(err, db.ErrNoActive) || errors.Is(err, db.ErrIndexOutRange) {
					cmd.Println("No task with such index.")
					return nil
				}
				return err
			}

			dur := time.Since(entry.Start)
			cmd.Printf("Paused #%d: %s (duration %s)\n",
				idx,
				entry.Description,
				dur.Truncate(time.Second),
			)
			return nil
		},
	}
	return cmd
}
