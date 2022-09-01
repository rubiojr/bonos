package db

import "github.com/rubiojr/kv"

type DB struct {
	Handler kv.Database
}

func New(dbPath string) (*DB, error) {
	db, err := kv.New("sqlite", dbPath)
	return &DB{Handler: db}, err
}
