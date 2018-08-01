package migrate

import (
	"errors"
	"flag"
	"io"

	"github.com/ki4jnq/forge/config"
	"github.com/ki4jnq/forge/lib/db"
	"github.com/ki4jnq/forge/migration"
)

var (
	conf = &Config{
		DBConn: db.DBConn{
			DBUser:  "postgres",
			DBHost:  "127.0.0.1",
			DBPort:  "5432",
			SSLMode: "require",
		},
		Table: "schema_versions",
	}
	flags = flag.NewFlagSet("db", flag.ExitOnError)
)

type Config struct {
	db.DBConn `yaml:",inline"`
	Table     string

	Rollback bool   `yaml:"-"`
	Migrate  bool   `yaml:"-"`
	Target   string `yaml:"-"`
	Init     bool   `yaml:"-"`
	Seed     bool   `yaml:"-"`
}

func init() {
	flags.BoolVar(&conf.Rollback, "rollback", false, "Only rollback migrations.")
	flags.BoolVar(&conf.Migrate, "migrate", false, "Migrate the database to the latest version.")
	flags.BoolVar(&conf.Init, "init", false, "Generate tables for the database.")
	flags.BoolVar(&conf.Seed, "seed", false, "Generate sample table data from a SQL file.")
	flags.StringVar(&conf.Target, "to", "", "The target version to migrate/rollback to.")
	flags.StringVar(&conf.SSLMode, "ssl", "disable", "The SSL Postgres argument")

	config.Register(&config.Cmd{
		Name:      "db",
		Flags:     flags,
		SubConf:   conf,
		SubRunner: run,
	})
}

func migrateOrRollback() error {
	runner, err := migration.NewRunner(conf.DBConn)
	if err != nil {
		return err
	}
	defer runner.Cleanup()

	target, err := migration.VersionFromString(conf.Target)
	if err != nil && err != io.EOF {
		return err
	}

	var fn func(migration.Version) error
	if conf.Migrate {
		fn = runner.UpTo
	} else if conf.Rollback {
		fn = runner.BackTo
	} else {
		return errors.New("Missing required arguments.")
	}
	return fn(target)
}

func initOrSeed() error {
	var exec db.Executer
	var sqlFile string
	if conf.Init {
		sqlFile = "db/sql/structure.sql"
	} else if conf.Seed {
		sqlFile = "db/sql/seed.sql"
	}
	return exec.ExecSQLFile(sqlFile, conf.DBConn)
}

func run() error {
	if conf.Init || conf.Seed {
		return initOrSeed()
	} else {
		return migrateOrRollback()
	}
}
