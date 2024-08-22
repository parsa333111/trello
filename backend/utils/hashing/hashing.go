package hashing_utils

import (
	"crypto/sha256"
)

func HashUsingSha256(input string) []byte {
	hash := sha256.New()
	hash.Write([]byte(input))

	hashed_input := hash.Sum(nil)

	return hashed_input
}
