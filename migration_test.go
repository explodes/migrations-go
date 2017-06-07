package migrations

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type testMigrationGetter struct{}

func (testMigrationGetter) GetMigration(version int) Migration { return NewSimpleMigration("test", "select 1", "select 1") }

func TestMigrator_Postgres(t *testing.T) {
	db, err := sql.Open("postgres", "postgresql://test_mig:test_mig@localhost/test_mig?sslmode=disable")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		defer db.Close()
		if _, err := db.Exec(`DELETE FROM migrations`); err != nil {
			t.Error(err)
		}
	}()

	test_migrator_on_connection(t, db)
}

func TestMigrator_Sqlite(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	test_migrator_on_connection(t, db)
}

func test_migrator_on_connection(t *testing.T, db *sql.DB) {
	migrator := NewMigrator(db, testMigrationGetter{})

	if err := migrator.MigrateToVersion(5); err != nil {
		t.Error(err)
	} else if version, err := migrator.GetCurrentVersion(); version != 5 || err != nil {
		t.Errorf("unexpected state: version=%d err=%v", version, err)
	}
	if err := migrator.MigrateToVersion(10); err != nil {
		t.Error(err)
	} else if version, err := migrator.GetCurrentVersion(); version != 10 || err != nil {
		t.Errorf("unexpected state: version=%d err=%v", version, err)
	}

	if err := migrator.MigrateToVersion(5); err != nil {
		t.Error(err)
	} else if version, err := migrator.GetCurrentVersion(); version != 5 || err != nil {
		t.Errorf("unexpected state: version=%d err=%v", version, err)
	}

	if err := migrator.MigrateToVersion(0); err != nil {
		t.Error(err)
	} else if version, err := migrator.GetCurrentVersion(); version != 0 || err != nil {
		t.Errorf("unexpected state: version=%d err=%v", version, err)
	}

}
