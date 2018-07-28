package main

import (
	"github.com/go-pg/migrations"
	//"github.com/go-pg/pg"
)

func init() {
	//db = pg.DB{}
	migrations.Register(func (db migrations.DB) error {
		err := db.CreateTable(&GameSettings{}, nil)
		return err
	}, func(db migrations.DB) error {
		err := db.DropTable(&GameSettings{}, nil)
		return err
	})
}
