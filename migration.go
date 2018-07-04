package migrations

import (
	"database/sql"
	"errors"
)

var (
	ErrDowngradeNotSupported = errors.New("downgrade not supported")
)

// Migration is an individual migration
type Migration interface {
	// Name returns the name of the migration
	Name() string

	// Upgrade runs the upwards migration
	Upgrade(tx *sql.Tx) error

	// Downgrade performs the undoing of this migration.
	// If downgrading is not supported, it should return
	// ErrDowngradeNotSupported
	Downgrade(tx *sql.Tx) error
}

type simpleSqlMigration struct {
	name      string
	upgrade   string
	downgrade string
}

// NewSimpleMigration creates a migration that executes
// the given sql statements on upgrade and downgrade
//
// If downgrade is the empty string, downgrade is not
// supported for this migration
func NewSimpleMigration(name, upgrade, downgrade string) Migration {
	return simpleSqlMigration{
		name:      name,
		upgrade:   upgrade,
		downgrade: downgrade,
	}
}

func (m simpleSqlMigration) Name() string {
	return m.name
}

func (m simpleSqlMigration) Upgrade(tx *sql.Tx) error {
	_, err := tx.Exec(m.upgrade)
	return err
}

func (m simpleSqlMigration) Downgrade(tx *sql.Tx) error {
	if m.downgrade == "" {
		return ErrDowngradeNotSupported
	}
	_, err := tx.Exec(m.downgrade)
	return err
}
