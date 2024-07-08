package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"log"
	"time"

	"github.com/zde37/Zero-Chain/transaction"
)

type Block struct {
	Hash         [32]byte                    
	Nonce        int                         
	Index        int                         
	TimeStamp    string                      
	PreviousHash [32]byte                    
	Transactions []*transaction.Transaction  
}

func NewBlock(nonce, previousIndex int, previousHash [32]byte, transactions []*transaction.Transaction) *Block {
	b := new(Block)
	b.Nonce = nonce
	b.Index = previousIndex + 1
	b.PreviousHash = previousHash
	b.Transactions = transactions
	b.TimeStamp = time.Now().String()

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
