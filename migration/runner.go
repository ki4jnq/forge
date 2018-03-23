package migration

import (
	"database/sql"
	"github.com/ki4jnq/forge/lib/db"
)

// TODO: Make the tablename dynamic.
const (
	findQuery   = "SELECT version FROM schema_versions ORDER BY version DESC LIMIT 1;"
	insertQuery = "INSERT INTO schema_versions (version) VALUES ($1);"
	deleteQuery = "DELETE FROM schema_versions WHERE version=$1;"
)

type Runner struct {
	Current    Version
	migrations []*Migration

	conn       *sql.DB
	connParams db.DBConn
}

func NewRunner(c db.DBConn) (*Runner, error) {
	db, err := sql.Open("postgres", c.DbUrl())
	if err != nil {
		return nil, err
	}

	r := &Runner{
		conn:       db,
		connParams: c,
	}

	r.migrations, err = ListMigrations()
	if err != nil {
		return nil, err
	}

	var rawVersion string
	err = db.QueryRow(findQuery).Scan(&rawVersion)
	if err == sql.ErrNoRows {
		r.Current = Version{0, 0, 0}
	} else if err != nil {
		return nil, err
	} else if err = r.Current.Scan(rawVersion); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Runner) Cleanup() {
	r.conn.Close()
}

// UpTo migrates the database up to the target version.
func (r *Runner) UpTo(target Version) error {
	migrations, err := ListMigrations()
	if err != nil {
		return err
	}

	for _, m := range migrations {
		// Skip the migration if it's outside of the desired range.
		// A Zero version is treated specially. If Zero is the target, then
		// run all pending migrations.
		if r.Current.Compare(m) >= 0 || (target.Compare(m) < 0 && !target.IsZero()) {
			continue
		}
		if err := m.Up(r.connParams); err != nil {
			return err
		}
		if err := r.insertVersion(m.Version); err != nil {
			return err
		}
	}

	return nil
}

// RollbackTo rolls the database back to the target version.
func (r *Runner) BackTo(target Version) error {
	migrations, err := ListMigrations()
	if err != nil {
		return err
	}

	for i := len(migrations) - 1; i >= 0; i-- {
		m := migrations[i]
		// Skip the rollbacks if they are outside of the desired range. A zero
		// version is treated specially. If Zero is the target, only rollback
		// the current schema_version.
		if (target.IsZero() && r.Current.Compare(m) != 0) ||
			r.Current.Compare(m) < 0 ||
			target.Compare(m) >= 0 {
			continue
		}
		if err := m.Down(r.connParams); err != nil {
			return err
		}
		if err := r.deleteVersion(m.Version); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) insertVersion(v Version) error {
	_, err := r.conn.Exec(insertQuery, v.String())
	return err
}

func (r *Runner) deleteVersion(v Version) error {
	_, err := r.conn.Exec(deleteQuery, v.String())
	return err
}
