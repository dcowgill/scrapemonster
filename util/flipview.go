package util

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"
)

var (
	// Matches the format of our generated tables.
	tableNameRegexp = regexp.MustCompile(`^(\w+)_(\d{14})$`)
)

type ViewFlipper struct {
	ViewName  string
	CreateSQL string
	UpsertSQL string
}

func (v *ViewFlipper) Create(db *sql.DB, args ...interface{}) error {
	// Ensure the user provided a SQL statement.
	if v.CreateSQL == "" {
		return errors.New("CreateSQL is empty")
	}
	// Create the new table.
	tableName := v.generateTableName()
	_, err := db.Exec(fmt.Sprintf(v.CreateSQL, tableName), args...)
	if err != nil {
		return err
	}
	// Recreate the view to point to the new table.
	_, err = db.Exec(fmt.Sprintf("create or replace view %s as select * from %s",
		v.ViewName, tableName))
	if err != nil {
		return err
	}
	// Drop obsolete tables if any exist.
	tableNames, err := v.findGeneratedTables(db)
	if err != nil {
		return err
	}
	for _, s := range tableNames {
		if s != tableName {
			_, err := db.Exec("drop table " + s)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *ViewFlipper) Upsert(db *sql.DB, args ...interface{}) error {
	// Ensure the user provided a SQL statement.
	if v.UpsertSQL == "" {
		return errors.New("UpsertSQL is empty")
	}
	// Determine the name of the active table.
	tableName, err := v.activeTableName(db)
	if err != nil {
		return err
	}
	// Replace the table name in the SQL statement.
	sql := fmt.Sprintf(v.UpsertSQL, tableName)
	_, err = db.Exec(sql, args...)
	return err
}

func (v *ViewFlipper) generateTableName() string {
	// Format base date: Mon Jan 2 15:04:05 -0700 MST 2006
	return fmt.Sprintf("%s_%s", v.ViewName, time.Now().Format("20060102150405"))
}

func (v *ViewFlipper) findGeneratedTables(db *sql.DB) (names []string, err error) {
	// Find tables which begin with our view name.
	var rows *sql.Rows
	rows, err = db.Query(fmt.Sprintf("show tables like '%s%%'", v.ViewName))
	if err != nil {
		return
	}
	defer rows.Close()
	// For each table name:
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return
		}
		// The table must match the pattern of our generated names.
		matches := tableNameRegexp.FindStringSubmatch(name)
		if matches != nil {
			if matches[1] != v.ViewName {
				panic(fmt.Sprintf("%v != %v in %v", matches[1], v.ViewName, name))
			}
			names = append(names, name)
		}
	}
	return
}

func (v *ViewFlipper) activeTableName(db *sql.DB) (string, error) {
	// Get the set of tables matching the view name.
	tableNames, err := v.findGeneratedTables(db)
	if err != nil {
		return "", err
	}
	// We expect to find at least one table.
	if len(tableNames) == 0 {
		return "", errors.New("no table found")
	}
	// Return the lexicographically most recent table name.
	sort.Strings(tableNames)
	return tableNames[len(tableNames)-1], nil
}
