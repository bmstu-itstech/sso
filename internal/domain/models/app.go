package models

type App struct {
	Id     int32  `db:"id"`
	Name   string `db:"name"`
	Secret string `db:"secret_key"`
}

type AppJWT struct {
	Uid   int64
	Login string
	Email string
	AppId int32
}
