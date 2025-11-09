package models

type App struct {
	Id     int32  `db:"id"`
	Name   string `db:"name"`
	Secret string `db:"secret_key"`
}
