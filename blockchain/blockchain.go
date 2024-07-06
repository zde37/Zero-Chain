package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/zde37/Zero-Chain/helpers"
	"github.com/zde37/Zero-Chain/transaction"
)

const (
	MINING_DIFFICULTY = 5
	MINING_SENDER     = "Zero-Chain"
	MINING_REWARD     = 7.0
	// MINING_TIMER_SEC  = 180 // 5 minutes
)

type BlockChain struct {
	Chain             []*Block
	MemPool           []*transaction.Transaction
	BlockChainAddress string
	Port              uint16
}

func New(blockchainAddress string, port uint16) *BlockChain {
	bc := &BlockChain{
		BlockChainAddress: blockchainAddress,
		Port:              port,
	}
	bc.CreateBlock(0, [32]byte{}) // genesis block

	return bc
}

func (bc *BlockChain) CreateBlock(nonce int, previousHash [32]byte) {
	block := NewBlock(nonce, previousHash, bc.MemPool)
	bc.Chain = append(bc.Chain, block)
	bc.MemPool = []*transaction.Transaction{} // clear memory pool on current blockchain node
}

func(bc *BlockChain) Run() {

}

func (bc *BlockChain) CopyMemPool() []*transaction.Transaction {
	transactions := make([]*transaction.Transaction, 0)
	transactions = append(transactions, bc.MemPool...)
	return transactions
}

func (bc *BlockChain) ClearMemPool() {
	bc.MemPool = bc.MemPool[:0]
}

func (bc *BlockChain) ValidProof(nonce int,
	previousHash [32]byte, transactions []*transaction.Transaction) bool {
	zeros := strings.Repeat("0", MINING_DIFFICULTY)
	tryBlock := Block{
		Nonce:        nonce,
		PreviousHash: previousHash,
		Timestamp:    0,
		Transactions: transactions,
	}
	tryHashStr := fmt.Sprintf("%x", tryBlock.GenerateHash())
	return tryHashStr[:MINING_DIFFICULTY] == zeros
}

func (bc *BlockChain) ProofOfWork() int {
	transactions := bc.CopyMemPool()
	previousHash := bc.LastBlock().Hash
	nonce := 0

	for !bc.ValidProof(nonce, previousHash, transactions) {
		nonce++
	}
	return nonce
}

func (bc *BlockChain) LastBlock() *Block {
	return bc.Chain[len(bc.Chain)-1]
}

func (bc *BlockChain) AddTransaction(senderBlockChainAddress, recipientBlockChainAddress string, value float32,
	senderPublicKey *ecdsa.PublicKey, s *helpers.Signature) bool {
	t := transaction.New(senderBlockChainAddress, recipientBlockChainAddress, value)
	if senderBlockChainAddress == MINING_SENDER { // miner
		bc.MemPool = append(bc.MemPool, t)
		return true
	}

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {
		if senderBlockChainAddress == recipientBlockChainAddress { // this should be checked on the wallet server and frontend and returned to the user
			log.Println("blockchain: you can't send money to yourself")
			return false
		}
		if bc.CalculateTotalAmount(senderBlockChainAddress) < value { // this should be checked on the wallet server and frontend and returned to the user
			log.Println("blockchain: Insufficient funds")
			return false
		}
		bc.MemPool = append(bc.MemPool, t)
		return true
	}
	return false
}

func (bc *BlockChain) VerifyTransactionSignature(senderPublicKey *ecdsa.PublicKey, s *helpers.Signature, t *transaction.Transaction) bool {
	m, err := json.Marshal(t)
	if err != nil {
		log.Printf("blockchain: failed to marshal transaction: %v", err)
		return false
	}
	hash := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublicKey, hash[:], s.R, s.S)
}

func (bc *BlockChain) Mining() {
	bc.AddTransaction(MINING_SENDER, bc.BlockChainAddress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash
	bc.CreateBlock(nonce, previousHash)
}

func (bc *BlockChain) CalculateWalletBalance(blockchainAddress string) float32 {
	var totalAmount float32
	for _, b := range bc.Chain {
		for _, t := range b.Transactions {
			value := t.Value
			if t.RecipientBlockChainAddress == blockchainAddress {
				totalAmount += value
			}

			if t.SenderBlockChainAddress == blockchainAddress {
				totalAmount -= value
			}
		}
	}
	return totalAmount
}
