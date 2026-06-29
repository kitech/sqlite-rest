package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

const dbNameDefault = "default"

type DatabaseRegistry struct {
	dbs map[string]*sqlx.DB
}

func NewDatabaseRegistry(dbs map[string]*sqlx.DB) *DatabaseRegistry {
	return &DatabaseRegistry{dbs: dbs}
}

func (r *DatabaseRegistry) Get(name string) (*sqlx.DB, bool) {
	db, ok := r.dbs[name]
	return db, ok
}

func (r *DatabaseRegistry) Default() *sqlx.DB {
	for _, db := range r.dbs {
		return db
	}
	return nil
}

func (r *DatabaseRegistry) Names() []string {
	names := make([]string, 0, len(r.dbs))
	for name := range r.dbs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *DatabaseRegistry) HasMultiple() bool {
	return len(r.dbs) > 1
}

func (r *DatabaseRegistry) Close() error {
	var firstErr error
	for name, db := range r.dbs {
		if err := db.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close %q: %w", name, err)
		}
	}
	return firstErr
}

func openDatabases(cmd *cobra.Command) (*DatabaseRegistry, error) {
	dsns, err := cmd.Flags().GetStringSlice(cliFlagDBDSN)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", cliFlagDBDSN, err)
	}

	dbDir, err := cmd.Flags().GetString(cliFlagDBDir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", cliFlagDBDir, err)
	}

	if len(dsns) == 0 && dbDir == "" {
		return nil, fmt.Errorf("specify at least one --%s or --%s", cliFlagDBDSN, cliFlagDBDir)
	}

	dbs := make(map[string]*sqlx.DB)

	for _, dsn := range dsns {
		name, path := parseDSN(dsn)
		db, err := openDatabaseFromDSN(name, path)
		if err != nil {
			return nil, fmt.Errorf("open dsn %q: %w", dsn, err)
		}
		if _, exists := dbs[name]; exists {
			db.Close()
			return nil, fmt.Errorf("duplicate database name %q", name)
		}
		dbs[name] = db
	}

	if dbDir != "" {
		dirDBs, err := openDatabasesFromDir(dbDir)
		if err != nil {
			return nil, fmt.Errorf("open dir %q: %w", dbDir, err)
		}
		for name, db := range dirDBs {
			if _, exists := dbs[name]; exists {
				db.Close()
				return nil, fmt.Errorf("duplicate database name %q from directory", name)
			}
			dbs[name] = db
		}
	}

	return NewDatabaseRegistry(dbs), nil
}

func parseDSN(dsn string) (name, path string) {
	if i := strings.IndexByte(dsn, '='); i > 0 {
		name = dsn[:i]
		path = dsn[i+1:]
	} else {
		name = stemFromPath(dsn)
		if name == "" {
			name = dbNameDefault
		}
		path = dsn
	}
	return name, path
}

func stemFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func openDatabaseFromDSN(name, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func openDatabasesFromDir(dir string) (map[string]*sqlx.DB, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	dbs := make(map[string]*sqlx.DB)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".db" && ext != ".sqlite" && ext != ".sqlite3" {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		p := filepath.Join(dir, entry.Name())
		db, err := openDatabaseFromDSN(name, p)
		if err != nil {
			return nil, fmt.Errorf("open %q: %w", p, err)
		}
		if _, exists := dbs[name]; exists {
			db.Close()
			return nil, fmt.Errorf("duplicate database name %q in directory", name)
		}
		dbs[name] = db
	}

	if len(dbs) == 0 {
		return nil, fmt.Errorf("no database files found in %q", dir)
	}

	return dbs, nil
}
