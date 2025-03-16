package random

import (
	"math/rand"
	"time"
)

func GetRandomString(length int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	alph := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		b[i] = alph[rnd.Intn(len(alph))]
	}

	return string(b)
}
