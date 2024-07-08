package transaction

import (
	"crypto/sha256"
	"encoding/json"
	"log"
	"time"
)

type Transaction struct {
	SenderBlockChainAddress    string
	RecipientBlockChainAddress string
	Value                      float32
	Hash                       [32]byte
	TimeStamp                  string
}

func New(senderBlockChainAddress string, recipientBlockChainAddress string, value float32) *Transaction {
	t := new(Transaction)
	t.SenderBlockChainAddress = senderBlockChainAddress
	t.RecipientBlockChainAddress = recipientBlockChainAddress
	t.Value = value
	t.TimeStamp = time.Now().String()
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

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.SenderBlockChainAddress,
		Recipient: t.RecipientBlockChainAddress,
		Value:     t.Value,
	})
}
