package main

import (
	"database/sql"
	"log"

	"github.com/explodes/migrations-go"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	migrator := migrations.NewMigrator(db, appMigrations{})
	if err := migrator.MigrateToVersion(VersionLatest); err != nil {
		log.Fatal(err)
	}

	// use db ....

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM images`).Scan(&count); err != nil {
		log.Fatal(err)
	}
	log.Printf("There are %d images in the database", count)
}

const (
	VersionInitial        = 1
	VersionImages         = 2
	VersionLatest         = VersionImages
	DowngradeNotSupported = ``
	UpgradeInitial        = `
CREATE TABLE users (
  id       INTEGER PRIMARY KEY,
  username VARCHAR(64),
  email    VARCHAR(128)
);

CREATE TABLE post (
  id      INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL,
  content TEXT NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users (id)
);
	`
	UpgradeImages = `
CREATE TABLE images (
  id  INTEGER PRIMARY KEY,
  url TEXT
)
`
	DowgradeImages = `DROP TABLE images`
)

type appMigrations struct{}

func (appMigrations) GetMigration(version int) migrations.Migration {
	switch version {
	case VersionInitial:
		// empty string, downgrade not supported
		return migrations.NewSimpleMigration("initial", UpgradeInitial, DowngradeNotSupported)
	case VersionImages:
		return migrations.NewSimpleMigration("create_images_table", UpgradeImages, DowgradeImages)
	}

	return nil
}
