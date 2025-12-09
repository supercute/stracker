package commands

import (
	"errors"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/supercute/stracker/internal/db"
)

func NewStartCmd(database *db.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start [description|index]",
		Aliases: []string{"s"},
		Short:   "Start new task or resume by index",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// st s 1 — resume по индексу
			if len(args) == 1 {
				if idx, err := strconv.Atoi(args[0]); err == nil && idx > 0 {
					entry, err := database.ResumeByIndex(idx)
					if err != nil {
						if errors.Is(err, db.ErrNoActive) || errors.Is(err, db.ErrIndexOutRange) {
							cmd.Println("No task with such index.")
							return nil
						}
						return err
					}
					cmd.Printf("Resumed #%d: %s\n", idx, entry.Description)
					return nil
				}
			}

			// обычный start с описанием
			desc := strings.Join(args, " ")
			if strings.TrimSpace(desc) == "" {
				cmd.Println("Description is required.")
				return nil
			}

			if err := database.StartEntry(desc); err != nil {
				return err
			}
			cmd.Println("Started:", desc)
			return nil
		},
	}
	return cmd
}
