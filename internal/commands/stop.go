package commands

import (
	"errors"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/supercute/stracker/internal/db"
)

func NewStopCmd(database *db.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stop [index]",
		Aliases: []string{"sp"},
		Short:   "Stop all tasks or one by index",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// без аргумента: остановить все активные
			if len(args) == 0 {
				if err := database.StopAllActive(); err != nil {
					return err
				}
				cmd.Println("Stopped all active tasks.")
				return nil
			}

			// с аргументом: st stop 1
			idx, err := strconv.Atoi(args[0])
			if err != nil || idx <= 0 {
				cmd.Println("Invalid index:", args[0])
				return nil
			}

			entry, err := database.StopByIndex(idx)
			if err != nil {
				if errors.Is(err, db.ErrNoActive) || errors.Is(err, db.ErrIndexOutRange) {
					cmd.Println("No task with such index.")
					return nil
				}
				return err
			}

			dur := time.Since(entry.Start)
			cmd.Printf("Stopped #%d: %s (duration %s)\n",
				idx,
				entry.Description,
				dur.Truncate(time.Second),
			)
			return nil
		},
	}
	return cmd
}
