package main

import (
	"crypto/sha256"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
	"math/big"
)

func PBKDF2(password []byte, salt []byte) (result string) {
	// FIXME
	dkey := pbkdf2.Key(password, salt, 250000, 8, sha256.New)
	return baseStringEnc(dkey)
}

func Scrypt(password []byte, salt []byte) (result string) {
	dkey, err := scrypt.Key(password, salt, 262144 /* 2^18 */, 8, 3, 12)
	if err != nil {
		panic(err)
	}
	return baseStringEnc(dkey)
}

var table []rune = nil

func init() {
	start, end := int8('0'), int8('z')+1
	table = make([]rune, end-start)
	for i := start; i < end; i++ {
		table[i-start] = rune(i)
	}
}

func baseStringEnc(data []byte) string {
	n := big.NewInt(0)
	n.SetBytes(data)
	zero := big.NewInt(0)
	if n.Cmp(zero) == 0 {
		return string(table[0])
	}
	digits := make([]rune, 0, 12)
	base := big.NewInt(int64(len(table)))
	for n.Cmp(zero) > 0 {
		rem := big.NewInt(0)
		n.DivMod(n, base, rem)
		digits = append(digits, table[rem.Int64()])
	}
	// Reverse the digits because the most significant digit came out last in the division.
	len := len(digits)
	digitsReversed := make([]rune, len)
	for i := 0; i < len; i++ {
		digitsReversed[i] = digits[len-i-1]
	}
	return string(digitsReversed)
}
