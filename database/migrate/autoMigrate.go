// Package migrate to migrate the schema for the example application
package migrate

import (
	"fmt"

	"github.com/tinkerbaj/gintemp/config"
	"github.com/tinkerbaj/gintemp/database"
	"github.com/tinkerbaj/gintemp/database/model"
)

// Load all the models
type twoFA model.TwoFA
type twoFABackup model.TwoFABackup
type tempEmail model.TempEmail
type user model.User
type post model.Post
type hobby model.Hobby

// type auth model.Auth
// type twoFA model.TwoFA
// type twoFABackup model.TwoFABackup
// type tempEmail model.TempEmail
// DropAllTables - careful! It will drop all the tables!
func DropAllTables() error {
	db := database.GetDB()

	if err := db.Migrator().DropTable(
		&hobby{},
		&post{},
		&user{},
		&tempEmail{},
		&twoFABackup{},
		&twoFA{},
	); err != nil {
		return err
	}

	fmt.Println("old tables are deleted!")
	return nil
}

// StartMigration - automatically migrate all the tables
//
// - Only create tables with missing columns and missing indexes
// - Will not change/delete any existing columns and their types
func StartMigration(configure config.Configuration) error {
	db := database.GetDB()
	configureDB := configure.Database.RDBMS
	driver := configureDB.Env.Driver

	if driver == "mysql" {
		// db.Set() --> add table suffix during auto migration
		if err := db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
			&twoFA{},
			&twoFABackup{},
			&tempEmail{},
			&user{},
			&post{},
			&hobby{},
		); err != nil {
			return err
		}

		fmt.Println("new tables are  migrated successfully!")
		return nil
	}

	if err := db.AutoMigrate(
		&twoFA{},
		&twoFABackup{},
		&tempEmail{},
		&user{},
		&post{},
		&hobby{},
	); err != nil {
		return err
	}

	fmt.Println("new tables are  migrated successfully!")
	return nil
}

// SetPkFk - manually set foreign key for MySQL and PostgreSQL
func SetPkFk() error {
	db := database.GetDB()

	if !db.Migrator().HasConstraint(&user{}, "Posts") {
		err := db.Migrator().CreateConstraint(&user{}, "Posts")
		if err != nil {
			return err
		}
	}

	return nil
}
