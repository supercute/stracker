package commands

import (
	"errors"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/supercute/stracker/internal/db"
)

func NewDeleteCmd(database *db.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete index",
		Aliases: []string{"d"},
		Short:   "Delete task by index",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idx, err := strconv.Atoi(args[0])
			if err != nil || idx <= 0 {
				cmd.Println("Invalid index:", args[0])
				return nil
			}

			if err := database.DeleteByIndex(idx); err != nil {
				if errors.Is(err, db.ErrNoActive) || errors.Is(err, db.ErrIndexOutRange) {
					cmd.Println("No task with such index.")
					return nil
				}
				return err
			}

			cmd.Printf("Deleted task #%d.\n", idx)
			return nil
		},
	}
	return cmd
}
