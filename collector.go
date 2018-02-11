package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sergeleger/powermeter/power"
	"github.com/urfave/cli"
)

// CollectorCommand collects power usage information from the command line.
var CollectorCommand = cli.Command{
	Name:      "collector",
	Usage:     "Collects power usage details from the command line.",
	ArgsUsage: "dbFile",
	Action:    collectorAction,
}

func collectorAction(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return errors.New("error: not enough arguments")
	}

	// open and create database file
	db, err := openDatabase(ctx.Args().First())
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error: could not create transaction: %v", err)
	}

	// prepare the statement
	st, err := tx.Prepare(`insert into power(MeterID, Time, Usage) values(?,?,?)`)
	if err != nil {
		return fmt.Errorf("error: could not prepare the statement: %v", err)
	}

	// Read standard input.
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		var usage power.Usage
		if err = usage.UnmarshalJSON(sc.Bytes()); err != nil {
			log.Printf("could not marshall entry: %v", err)
		}

		st.Exec(usage.MeterID, usage.Time, usage.Consumption)
	}

	tx.Commit()
	st.Close()

	return nil
}

// openDatabase creates the database and prepares the database object.
func openDatabase(dbFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("error: could not open database %q: %v", dbFile, err)
	}

	// create database objects
	if _, err := db.Exec(ddl); err != nil {
		db.Close()
		return nil, fmt.Errorf("error: could not create Power table: %v", err)
	}

	return db, nil
}

var ddl = `
create table if not exists power (
	MeterID integer,
	Time integer,
	Usage real
);

create view if not exists power_by_day as
	select
		MeterID,
		strftime("%Y-%m-%d", Time) as Time,
		max(Usage) as Usage
	from
		Power
	group by MeterID, strftime("%Y-%m-%d", Time);
`
