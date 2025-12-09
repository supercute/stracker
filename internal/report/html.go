package report

import (
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/supercute/stracker/internal/db"
)

const tpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Time report - {{.Period}}</title>
<style>
body { font-family: sans-serif; }
table { border-collapse: collapse; width: 100%; }
th, td { border: 1px solid #ccc; padding: 4px 8px; font-size: 14px; }
th { background: #f0f0f0; text-align: left; }
</style>
</head>
<body>
<h1>Time report - {{.Period}}</h1>
<p>From {{.From}} to {{.To}}</p>
<table>
<thead>
<tr>
<th>Start</th>
<th>End</th>
<th>Duration</th>
<th>Description</th>
</tr>
</thead>
<tbody>
{{range .Rows}}
<tr>
<td>{{.Start}}</td>
<td>{{.End}}</td>
<td>{{.Duration}}</td>
<td>{{.Description}}</td>
</tr>
{{end}}
</tbody>
</table>
<p><strong>Total: {{.Total}}</strong></p>
</body>
</html>`

type row struct {
	Start       string
	End         string
	Duration    string
	Description string
}

type data struct {
	Period string
	From   string
	To     string
	Rows   []row
	Total  string
}

func WriteHTML(period string, from, to time.Time, entries []db.Entry) (string, error) {
	now := time.Now()
	var rows []row
	var total time.Duration

	for _, e := range entries {
		end := e.End
		if end == nil {
			t := now
			end = &t
		}
		dur := end.Sub(e.Start)
		total += dur

		rows = append(rows, row{
			Start:       e.Start.Format("02.01.2006 15:04"),
			End:         end.Format("02.01.2006 15:04"),
			Duration:    dur.Truncate(time.Second).String(),
			Description: e.Description,
		})
	}

	d := data{
		Period: period,
		From:   from.Format("02.01.2006 15:04"),
		To:     to.Format("02.01.2006 15:04"),
		Rows:   rows,
		Total:  total.Truncate(time.Second).String(),
	}

	filename := filenameForPeriod(period)
	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	t := template.Must(template.New("report").Parse(tpl))
	if err := t.Execute(f, d); err != nil {
		return "", err
	}
	return filename, nil
}

func filenameForPeriod(period string) string {
	switch period {
	case "day":
		return filepath.Join(".", "time-report-day.html")
	case "week":
		return filepath.Join(".", "time-report-week.html")
	case "month":
		return filepath.Join(".", "time-report-month.html")
	default:
		return filepath.Join(".", "time-report.html")
	}
}
