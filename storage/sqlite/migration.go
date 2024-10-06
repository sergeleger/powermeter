package sqlite

import (
	"fmt"
	"io/fs"
	"path"
	"sort"

	"github.com/jmoiron/sqlx"
)

// migrateSchema migrates the database schema if it is required.
func migrateSchema(db *Database) error {
	if err := db.Ping(); err != nil {
		return err
	}

	ok, err := tableExists(db, "history")
	if err != nil {
		return err
	}

	if !ok {
		return executeDDLs(db, "")
	}

	// the schema history table exist, find the last migration script.
	lastEntry, err := findLastEntry(db)
	if err != nil {
		return err
	}

	return executeDDLs(db, lastEntry)
}

// tableExists verifies the table exists in the database.
func tableExists(db *Database, tableName string) (bool, error) {
	var counter int64

	stmt := `select count(*) from sqlite_master where type='table' and name=?`
	return counter == 1, sqlx.Get(db, &counter, stmt, tableName)
}

// findLastEntry returns the last schema_history entry
func findLastEntry(db *Database) (fileName string, _ error) {
	stmt := `select file from history order by file desc limit 1`
	return fileName, sqlx.Get(db, &fileName, stmt)
}

// executeDDLs executes each of the migration scripts that have not been executed yet.
func executeDDLs(db *Database, lastEntry string) error {
	// read and sort the DDL files.
	files, _ := ddlFS.ReadDir("ddl")
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	// Apply each file that are newer than lastEntry
	var buf []byte
	var err error
	for _, file := range files {
		name := file.Name()
		if name <= lastEntry {
			continue
		}

		if buf, err = fs.ReadFile(ddlFS, path.Join("ddl", name)); err != nil {
			return fmt.Errorf("migrate: error in %q: %w", name, err)
		}

		if _, err := db.Exec(string(buf)); err != nil {
			return fmt.Errorf("migrate: error in %q: %w", name, err)
		}
	}

	return nil
}
