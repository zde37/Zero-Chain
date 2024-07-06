package helpers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

func (s *Signature) String() string {
	return fmt.Sprintf("%064x%064x", s.R, s.S)
}

func StringToBigIntTuple(s string) (big.Int, big.Int) {
	bx, err := hex.DecodeString(s[:64])
	if err != nil {
		log.Printf("ecdsa: failed to decode string: %v", err)
		return big.Int{}, big.Int{}
	}
	by, err := hex.DecodeString(s[64:])
	if err != nil {
		log.Printf("ecdsa: failed to decode string: %v", err)
		return big.Int{}, big.Int{}
	}

	var bix, biy big.Int
	_ = bix.SetBytes(bx)
	_ = biy.SetBytes(by)

	return bix, biy
}

func SignatureFromString(s string) *Signature {
	x, y := StringToBigIntTuple(s)
	return &Signature{R: &x, S: &y}
}

func PublicKeyFromString(s string) *ecdsa.PublicKey {
	x, y := StringToBigIntTuple(s)
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: &x, Y: &y}
}

func PrivateKeyFromString(s string, publicKey *ecdsa.PublicKey) *ecdsa.PrivateKey {
	b, err := hex.DecodeString(s[:])
	if err != nil {
		log.Printf("ecdsa: failed to decode string: %v", err)
		return nil
	}
	var bi big.Int
	_ = bi.SetBytes(b)
	return &ecdsa.PrivateKey{PublicKey: *publicKey, D: &bi}
}
