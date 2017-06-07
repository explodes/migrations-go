package migrations

import (
	"database/sql"
	"fmt"
	"time"
)

// MigrationGetter gets migrations by version number
type MigrationGetter interface {
	// GetMigrations gets a migration by version number
	GetMigration(version int) Migration
}

// Migrator is used to upgrade and downgrade database versions.
//
// Version 0 is considered the "clean" slate version, and version 1 is the initial version.
type Migrator struct {
	db     *sql.DB
	getter MigrationGetter
}

// NewMigrator builds a new migrator to be used to run database upgrades
// and downgrades using the given database connection using the given dialect
func NewMigrator(db *sql.DB, getter MigrationGetter) *Migrator {
	return &Migrator{
		db:     db,
		getter: getter,
	}
}

// MigrateToVersion will migrate the database up or down to get the specified version
func (m *Migrator) MigrateToVersion(version int) error {
	current, err := m.GetCurrentVersion()
	if err != nil {
		return err
	}

	// nothing to migrate!
	if current == version {
		return nil
	}

	// upgrade or downgrade
	if version < current {
		return m.DowngradeDatabase(current, version)
	} else {
		return m.UpgradeDatabase(current, version)
	}
}

func (m *Migrator) GetCurrentVersion() (version int, err error) {
	version = 0

	// start our transaction
	tx, err := m.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// create our table if it doesn't exist already
	if _, err = tx.Exec(`CREATE TABLE IF NOT EXISTS migrations (version INTEGER NOT NULL UNIQUE, name TEXT NOT NULL, date_ran TIMESTAMP WITH TIME ZONE NOT NULL)`); err != nil {
		return
	}

	var max sql.NullInt64
	if err = tx.QueryRow(`SELECT MAX(version) FROM migrations`).Scan(&max); err != nil {
		// if some driver considers MAX(foo) of an empty table as no row,
		// then we should account for that
		if err == sql.ErrNoRows {
			version = 0
			err = nil
			return
		}
		return
	}
	if max.Valid {
		version = int(max.Int64)
	}
	return
}

func (m *Migrator) recordUpgrade(version int, migration Migration) error {
	_, err := m.db.Exec(`INSERT INTO migrations (version, name, date_ran) VALUES ($1,$2,$3)`, version+1, migration.Name(), time.Now())
	return err
}

func (m *Migrator) UpgradeDatabase(from, to int) error {
	for version := from; version < to; version ++ {
		migration := m.getter.GetMigration(version + 1)
		if migration == nil {
			return fmt.Errorf("unexpected nil upgrade migration for version %d", version)
		}
		if err := migration.Upgrade(m.db); err != nil {
			return err
		}
		if err := m.recordUpgrade(version, migration); err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) recordDowngrade(version int) error {
	_, err := m.db.Exec(`DELETE FROM migrations WHERE version = $1`, version+1)
	return err
}

func (m *Migrator) DowngradeDatabase(from, to int) error {
	for version := from - 1; version >= to; version -- {
		migration := m.getter.GetMigration(version + 1)
		if migration == nil {
			return fmt.Errorf("unexpected nil downgrade migration for version %d", version)
		}
		if err := migration.Downgrade(m.db); err != nil {
			return err
		}
		if err := m.recordDowngrade(version); err != nil {
			return err
		}
	}
	return nil
}
