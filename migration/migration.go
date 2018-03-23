package migration

import (
	"fmt"
	_ "github.com/lib/pq"
	"github.com/ki4jnq/forge/lib/db"
	"path/filepath"
	"sort"
)

type Migration struct {
	Version
	db.Executer
}

// ListMigrations returns a sorted list of all migrations in the db/migrations
// directory.
func ListMigrations() ([]*Migration, error) {
	filenames, err := filepath.Glob("db/sql/version-*.sql")
	if err != nil {
		return []*Migration{}, err
	}

	sort.Slice(filenames, func(i, j int) bool {
		return filenames[i] < filenames[j]
	})

	migrations := make([]*Migration, 0, len(filenames))
	for _, filename := range filenames {
		migration, err := NewFromFilename(filename)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

func NewFromFilename(filename string) (*Migration, error) {
	v := Version{}
	if err := v.Scanf(filename, "db/sql/version-%d.%d.%d.sql"); err != nil {
		return nil, err
	}
	return NewFromVersion(v), nil
}

func NewFromVersion(version Version) *Migration {
	return &Migration{
		Version:  version,
		Executer: db.Executer{},
	}
}

// Up runs the migration.
func (m *Migration) Up(conn db.DBConn) error {
	if m.Version.IsZero() {
		return nil
	}

	filename := fmt.Sprintf("db/sql/version-%s.sql", m.String())

	if err := m.ExecSQLFile(filename, conn); err != nil {
		return err
	}
	return nil
}

// Down runs the "rollback" file for this migration.
func (m *Migration) Down(conn db.DBConn) error {
	if m.Version.IsZero() {
		return nil
	}

	filename := fmt.Sprintf("db/sql/rollback-%s.sql", m.String())

	if err := m.ExecSQLFile(filename, conn); err != nil {
		return err
	}
	return nil
}
