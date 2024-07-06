package transaction

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"log"

	"github.com/zde37/Zero-Chain/helpers"
)

type MetaData struct {
	SenderPrivateKey           *ecdsa.PrivateKey
	SenderPublicKey            *ecdsa.PublicKey
	SenderBlockchainAddress    string
	RecipientBlockchainAddress string
	Value                      float32
}

func NewMetaData(senderPrivateKey *ecdsa.PrivateKey, senderPublicKey *ecdsa.PublicKey, senderBlockchainAddress string,
	recipientBlockchainAddress string, value float32) *MetaData {
	return &MetaData{
		SenderPrivateKey:           senderPrivateKey,
		SenderPublicKey:            senderPublicKey,
		SenderBlockchainAddress:    senderBlockchainAddress,
		RecipientBlockchainAddress: recipientBlockchainAddress,
		Value:                      value,
	}
}

func (md *MetaData) GenerateSignature() *helpers.Signature {
	m, err := json.Marshal(md)
	if err != nil {
		log.Printf("wallet: failed to marshal transaction: %v", err)
		return nil
	}
	hash := sha256.Sum256([]byte(m))
	r, s, err := ecdsa.Sign(rand.Reader, md.SenderPrivateKey, hash[:])
	if err != nil {
		log.Printf("wallet: sign hash: %v", err)
		return nil
	}
	return &helpers.Signature{
		R: r,
		S: s,
	}
}

func (md *MetaData) Validate() bool {
	if md.SenderPrivateKey == nil ||
		md.SenderPublicKey == nil ||
		md.SenderBlockchainAddress == "" ||
		md.RecipientBlockchainAddress == "" ||
		md.Value == 0.0 {
		return false
	}

	return true
}
