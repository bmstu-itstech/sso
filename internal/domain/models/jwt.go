package models

type TokenInfo struct {
	Uid     int64
	AppId   int32
	IsAdmin bool
}

type TokenModel struct {
	Uid     int64
	AppId   int32
	Secret  string
	IsAdmin bool
}
