package sqlite

import (
	"context"
	"io/fs"
	"path/filepath"
	"sort"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
)

// initDDL inspects the history to determine the last ddl created
func (s *Service) initDDL() error {
	conn := s.db.Get(context.Background())
	defer s.db.Put(conn)

	// Verify if the database exist.
	stmt := conn.Prep(`SELECT name FROM sqlite_master WHERE type='table' AND name=$name`)
	stmt.SetText("$name", "history")
	defer stmt.Finalize()

	switch hasRow, err := stmt.Step(); {
	case err != nil:
		return err

	case !hasRow:
		return s.createDDL(conn, "")

	default:
		name, err := s.lastEntry(conn)
		if err != nil {
			return err
		}
		return s.createDDL(conn, name)
	}
}

// createDDL creates the required DDL statements
func (s *Service) createDDL(conn *sqlite.Conn, last string) error {
	// read and sort the DDL files.
	files, _ := ddlFS.ReadDir("ddl")
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	// Apply each file that are newer.
	var buf []byte
	var err error
	for _, file := range files {
		name := file.Name()
		if name <= last {
			continue
		}

		if buf, err = fs.ReadFile(ddlFS, filepath.Join("ddl", name)); err != nil {
			return err
		}

		if err = sqlitex.ExecScript(conn, string(buf)); err != nil {
			return err
		}
	}

	return nil
}

// lastEntry locates the last DDL executed on the this database.
func (s *Service) lastEntry(conn *sqlite.Conn) (string, error) {
	q := `select file
		from history
	where
		rowid = (select max(rowid) from history)`
	stmt := conn.Prep(q)
	defer stmt.Finalize()
	return sqlitex.ResultText(stmt)
}
