package commands

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/supercute/stracker/internal/db"
	"github.com/supercute/stracker/internal/report"
)

func NewSumCmd(database *db.DB) *cobra.Command {
	var html bool

	cmd := &cobra.Command{
		Use:   "sum [day|week|month]",
		Short: "Show time stats",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			period := args[0]
			now := time.Now()
			from, to := periodRange(now, period)

			entries, err := database.ListBetween(from, to)
			if err != nil {
				return err
			}

			if html {
				filename, err := report.WriteHTML(period, from, to, entries)
				if err != nil {
					return err
				}
				cmd.Println("Report written to", filename)
				return nil
			}

			var total time.Duration
			for _, e := range entries {
				end := e.End
				if end == nil {
					t := now
					end = &t
				}
				dur := end.Sub(e.Start)
				total += dur
				cmd.Printf("%s - %s  (%s)  %s\n",
					e.Start.Format("02.01.2006 15:04"),
					end.Format("02.01.2006 15:04"),
					dur.Truncate(time.Second),
					e.Description,
				)
			}
			cmd.Println("Total:", total.Truncate(time.Second))
			return nil
		},
	}

	cmd.Flags().BoolVar(&html, "html", false, "export report as HTML")
	return cmd
}

func periodRange(now time.Time, period string) (time.Time, time.Time) {
	y, m, d := now.Date()
	loc := now.Location()

	switch period {
	case "day":
		start := time.Date(y, m, d, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 0, 1)
		return start, end
	case "week":
		// ISO week: понедельник
		wd := int(now.Weekday())
		if wd == 0 {
			wd = 7
		}
		start := time.Date(y, m, d, 0, 0, 0, 0, loc).AddDate(0, 0, -wd+1)
		end := start.AddDate(0, 0, 7)
		return start, end
	case "month":
		start := time.Date(y, m, 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 1, 0)
		return start, end
	default:
		// по умолчанию день
		start := time.Date(y, m, d, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 0, 1)
		return start, end
	}
}
