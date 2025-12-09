package db

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	sql *sql.DB
}

type Entry struct {
	ID          int64
	Start       time.Time
	End         *time.Time
	Description string
}

func Open() (*DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".stracker")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dir, "st.db")

	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := migrate(sqlDB); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return &DB{sql: sqlDB}, nil
}

func (d *DB) Close() error {
	return d.sql.Close()
}

func migrate(s *sql.DB) error {
	_, err := s.Exec(`
		CREATE TABLE IF NOT EXISTS entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			start TIMESTAMP NOT NULL,
			end   TIMESTAMP,
			description TEXT NOT NULL
		);
	`)
	return err
}

var (
	ErrNoActive      = errors.New("no active entry")
	ErrIndexOutRange = errors.New("no entry with such index")
)

// StartEntry просто создаёт новую активную задачу.
func (d *DB) StartEntry(desc string) error {
	_, err := d.sql.Exec(
		`INSERT INTO entries(start, description) VALUES(?, ?)`,
		time.Now(),
		desc,
	)
	return err
}

// StopAllActive — старый stop без индексов (если хочешь сохранить).
func (d *DB) StopAllActive() error {
	_, err := d.sql.Exec(`
		UPDATE entries
		SET end = ?
		WHERE end IS NULL
	`, time.Now())
	return err
}

func (d *DB) StopByIndex(idx int) (*Entry, error) {
	actives, err := d.ActiveEntries()
	if err != nil {
		return nil, err
	}
	if len(actives) == 0 {
		return nil, ErrNoActive
	}
	if idx < 1 || idx > len(actives) {
		return nil, ErrIndexOutRange
	}
	target := actives[idx-1]

	now := time.Now()
	_, err = d.sql.Exec(`
		UPDATE entries
		SET end = ?
		WHERE id = ?
	`, now, target.ID)
	if err != nil {
		return nil, err
	}
	target.End = &now
	return &target, nil
}

// ActiveEntries возвращает все активные по start ASC.
func (d *DB) ActiveEntries() ([]Entry, error) {
	rows, err := d.sql.Query(`
		SELECT id, start, end, description
		FROM entries
		WHERE end IS NULL
		ORDER BY start ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Entry
	for rows.Next() {
		var e Entry
		var end sql.NullTime
		if err := rows.Scan(&e.ID, &e.Start, &end, &e.Description); err != nil {
			return nil, err
		}
		if end.Valid {
			e.End = &end.Time
		}
		res = append(res, e)
	}
	return res, rows.Err()
}

func (d *DB) PauseByIndex(idx int) (*Entry, error) {
	list, err := d.IndexedEntries()
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrNoActive
	}
	if idx < 1 || idx > len(list) {
		return nil, ErrIndexOutRange
	}
	target := list[idx-1]
	if target.End != nil {
		// уже пауза/завершена
		return nil, ErrIndexOutRange
	}

	now := time.Now()
	_, err = d.sql.Exec(`
		UPDATE entries
		SET end = ?
		WHERE id = ?
	`, now, target.ID)
	if err != nil {
		return nil, err
	}
	target.End = &now
	return &target, nil
}

func (d *DB) ResumeByIndex(idx int) (*Entry, error) {
	list, err := d.IndexedEntries()
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrNoActive
	}
	if idx < 1 || idx > len(list) {
		return nil, ErrIndexOutRange
	}
	src := list[idx-1]

	now := time.Now()
	res, err := d.sql.Exec(
		`INSERT INTO entries(start, description) VALUES(?, ?)`,
		now,
		src.Description,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Entry{
		ID:          id,
		Start:       now,
		End:         nil,
		Description: src.Description,
	}, nil
}

func (d *DB) ListBetween(from, to time.Time) ([]Entry, error) {
	rows, err := d.sql.Query(`
		SELECT id, start, end, description
		FROM entries
		WHERE start >= ? AND start < ?
		ORDER BY start ASC
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Entry
	for rows.Next() {
		var e Entry
		var end sql.NullTime
		if err := rows.Scan(&e.ID, &e.Start, &end, &e.Description); err != nil {
			return nil, err
		}
		if end.Valid {
			e.End = &end.Time
		}
		res = append(res, e)
	}
	return res, rows.Err()
}

func (d *DB) IndexedEntries() ([]Entry, error) {
	rows, err := d.sql.Query(`
		SELECT id, start, end, description
		FROM entries
		ORDER BY start ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Entry
	for rows.Next() {
		var e Entry
		var end sql.NullTime
		if err := rows.Scan(&e.ID, &e.Start, &end, &e.Description); err != nil {
			return nil, err
		}
		if end.Valid {
			e.End = &end.Time
		}
		res = append(res, e)
	}
	return res, rows.Err()
}

func (d *DB) DeleteByIndex(idx int) error {
	list, err := d.IndexedEntries()
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return ErrNoActive
	}
	if idx < 1 || idx > len(list) {
		return ErrIndexOutRange
	}
	target := list[idx-1]

	_, err = d.sql.Exec(`DELETE FROM entries WHERE id = ?`, target.ID)
	return err
}
