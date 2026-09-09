package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand/v2"
)

const (
	Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func RandomInt(max, min int) int {
	return rand.IntN(max-min) + min
}

func RandomInt64(max, min int64) int64 {
	return rand.Int64N(max-min) + min
}

func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = Alphabet[RandomInt(len(Alphabet)-1, 0)]
	}
	return string(b)
}

func RandomSha1() string {
	str := RandomString(40)
	hash := sha256.New()
	hash.Write([]byte(str))
	hexStr := hex.EncodeToString(hash.Sum(nil))
	return hexStr
}
