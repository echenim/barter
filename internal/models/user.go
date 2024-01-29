package models

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

type User struct {
	ID         int64
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(privKey string, id int64) *User {
	pk, err := crypto.HexToECDSA(privKey)
	if err != nil {
		panic(err)
	}

	return &User{
		ID:         id,
		PrivateKey: pk,
	}
}
