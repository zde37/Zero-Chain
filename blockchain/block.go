package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"log"
	"time"

	"github.com/zde37/Zero-Chain/transaction"
)

type Block struct {
	Hash         [32]byte                   `json:"hash"`
	Nonce        int                        `json:"int"`
	Timestamp    int64                      `json:"timestamp"`
	PreviousHash [32]byte                   `json:"previous_hash"`
	Transactions []*transaction.Transaction `json:"transactions"`
}

func NewBlock(nonce int, previousHash [32]byte, transactions []*transaction.Transaction) *Block {
	b := &Block{
		Nonce:        nonce,
		PreviousHash: previousHash,
		Transactions: transactions,
		Timestamp:    time.Now().UnixNano(),
	}
	b.Hash = b.GenerateHash()

	return b
}

func (b *Block) GenerateHash() [32]byte {
	m, err := json.Marshal(b)
	if err != nil {
		log.Printf("block: failed to generate hash: %v", err)
		return sha256.Sum256([]byte("Invalid block hash"))
	}
	return sha256.Sum256([]byte(m))
}

// // MarshalJSON is created so the value of Hash and PreviousHash will be printed as string not [32]byte
// func (b *Block) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(struct {
// 		Hash         string         `json:"hash"`
// 		Nonce        int            `json:"int"`
// 		Timestamp    int64          `json:"timestamp"`
// 		PreviousHash string         `json:"previous_hash"`
// 		Transactions []*Transaction `json:"transactions"`
// 	}{
// 		Timestamp:    b.Timestamp,
// 		Nonce:        b.Nonce,
// 		PreviousHash: fmt.Sprintf("%x", b.PreviousHash),
// 		Hash:         fmt.Sprintf("%x", b.Hash),
// 		Transactions: b.Transactions,
// 	})
// }
