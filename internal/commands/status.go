package commands

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/supercute/stracker/internal/db"
)

func NewStatusCmd(database *db.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   "Show tasks for indexing",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := database.IndexedEntries()
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				cmd.Println("No tasks.")
				return nil
			}

			now := time.Now()
			cmd.Println("Tasks:")
			for i, e := range entries {
				idx := i + 1
				end := now
				state := "active"
				if e.End != nil {
					end = *e.End
					state = "paused"
				}
				dur := end.Sub(e.Start)
				cmd.Printf("  #%d  [%s] %s  (%s)  %s\n",
					idx,
					state,
					e.Description,
					dur.Truncate(time.Second),
					e.Start.Format("2006-01-02 15:04:05"),
				)
			}
			return nil
		},
	}
	return cmd
}
