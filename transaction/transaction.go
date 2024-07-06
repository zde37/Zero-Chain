package transaction

import (
	"crypto/sha256"
	"encoding/json"
	"log"
	"time"
)

type Transaction struct {
	SenderBlockChainAddress    string   `json:"sender_blockchain_address"`
	RecipientBlockChainAddress string   `json:"recipient_blockchain_address"`
	Value                      float32  `json:"value"`
	Hash                       [32]byte `json:"hash"`
	TimeStamp                  int64    `json:"timestamp"`
	Status                     string   `json:"status"`
}

func New(senderBlockChainAddress string, recipientBlockChainAddress string, value float32) *Transaction {
	t := &Transaction{
		SenderBlockChainAddress:    senderBlockChainAddress,
		RecipientBlockChainAddress: recipientBlockChainAddress,
		Value:                      value,
		TimeStamp:                  time.Now().UnixNano(),
		Status:                     "Pending",
	}
	t.Hash = t.TxHash()
	return t
}

func (t *Transaction) TxHash() [32]byte {
	m, err := json.Marshal(t)
	if err != nil {
		log.Printf("transaction: failed to generate hash: %v", err)
		return sha256.Sum256([]byte("Invalid transaction hash"))
	}
	return sha256.Sum256([]byte(m))
}
