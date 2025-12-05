package models

type TokenInfo struct {
	Uid   int64
	AppId int32
}

type TokenModel struct {
	Uid    int64
	AppId  int32
	Secret string
}
