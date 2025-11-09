package lib

import (
	"math/rand"
)

func RandoInt64() int64 {
	num := rand.Int63()
	return num
}
