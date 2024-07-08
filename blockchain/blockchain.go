package blockchain

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/zde37/Zero-Chain/helpers"
	"github.com/zde37/Zero-Chain/protobuf/protogen"
	"github.com/zde37/Zero-Chain/transaction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	MINING_DIFFICULTY = 4
	MINING_SENDER     = "Zero-Chain"
	MINING_REWARD     = 2.0
	MINING_TIMER_SEC  = 200

	BLOCKCHAIN_PORT_RANGE_START       = 7000
	BLOCKCHAIN_PORT_RANGE_END         = 7003
	NEIGHBOR_IP_RANGE_START           = 0
	NEIGHBOR_IP_RANGE_END             = 1
	BLOCKCHAIN_NEIGHBOR_SYNC_TIME_SEC = 10
)

type BlockChain struct {
	Chain             []*Block
	MemPool           []*transaction.Transaction
	BlockChainAddress string
	Port              uint16
	mut               sync.Mutex
	wgConsensus       *sync.WaitGroup
	wgMining          *sync.WaitGroup
	transactionChan   chan bool
	wgBlock           *sync.WaitGroup

	neighbors    []string
	mutNeighbors sync.Mutex
}

func New(blockchainAddress string, port uint16) *BlockChain {
	bc := new(BlockChain)
	bc.BlockChainAddress = blockchainAddress
	bc.Port = port
	bc.wgConsensus = new(sync.WaitGroup)
	bc.wgBlock = new(sync.WaitGroup)
	bc.transactionChan = make(chan bool)
	bc.wgMining = new(sync.WaitGroup)
	bc.genesisBlock() 
	return bc
}

func (bc *BlockChain) Run() {
	bc.StartSyncNeighbors()
	bc.ResolveConflicts()
	bc.StartMining()
}

func(bc *BlockChain) genesisBlock()  {
	block := new(Block) 
	block.Hash = block.GenerateHash()
	block.Index = 0
	block.PreviousHash = [32]byte{}
	block.TimeStamp = time.Now().String() 
	bc.Chain = append(bc.Chain, block)
}

func (bc *BlockChain) CreateBlock(nonce, previousIndex int, previousHash [32]byte) {
	block := NewBlock(nonce, previousIndex, previousHash, bc.MemPool)
	bc.Chain = append(bc.Chain, block)
	bc.MemPool = []*transaction.Transaction{} // clear memory pool on current blockchain node

	bc.wgBlock.Add(len(bc.neighbors))
	ctx := context.Background()
	for _, n := range bc.neighbors { // clear memory pool on other blockchain nodes
		go func() {
			conn, err := grpc.NewClient(
				n,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				log.Printf("create-block: failed to create grpc client on %s node: %v", n, err)
				return
			}
			defer conn.Close()

			client := protogen.NewBlockChainServiceClient(conn)
			resp, err := client.DeleteTransaction(ctx, &protogen.Empty{})
			if err != nil {
				log.Printf("create-block: failed to clear mempool on %s node: %v", n, err)
				return
			}
			bc.wgBlock.Done()
			log.Printf("create-block: %s", resp.GetStatus())
		}()
	}
	bc.wgBlock.Wait()
}

func (bc *BlockChain) SetNeighbors() {
	bc.neighbors = helpers.FindNeighbors(
		"127.0.0.1", bc.Port, NEIGHBOR_IP_RANGE_START, NEIGHBOR_IP_RANGE_END,
		BLOCKCHAIN_PORT_RANGE_START, BLOCKCHAIN_PORT_RANGE_END)

	log.Printf("neighbors: %v", bc.neighbors)
}

func (bc *BlockChain) SyncNeighbors() {
	bc.mutNeighbors.Lock()
	defer bc.mutNeighbors.Unlock()
	bc.SetNeighbors()
}

func (bc *BlockChain) StartSyncNeighbors() {
	bc.SyncNeighbors()
	_ = time.AfterFunc(BLOCKCHAIN_NEIGHBOR_SYNC_TIME_SEC*time.Second, bc.StartSyncNeighbors)
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
		TimeStamp:    "",
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

func (bc *BlockChain) CreateTransaction(ctx context.Context, sender, recipient string, value float32, senderPublicKey *ecdsa.PublicKey, s *helpers.Signature) bool {
	isTransacted := bc.AddTransaction(sender, recipient, value, senderPublicKey, s) 
	if isTransacted {
		publicKeyStr := fmt.Sprintf("%064x%064x", senderPublicKey.X.Bytes(), senderPublicKey.Y.Bytes())
		signatureStr := s.String()
 
		for _, n := range bc.neighbors {
			go func(ch chan<- bool) {
				conn, err := grpc.NewClient(
					n,
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				) 
				if err != nil {
					log.Printf("create-transaction: failed to create grpc client on %s node: %v", n, err)
					ch <- false
					return
				}
				defer conn.Close()
 
				client := protogen.NewBlockChainServiceClient(conn)
				resp, err := client.UpdateTransaction(ctx, &protogen.TransactionRequest{
					SenderBlockchainAddress:    sender,
					RecipientBlockchainAddress: recipient,
					SenderPublicKey:            publicKeyStr,
					Signature:                  signatureStr,
					Value:                      value,
				}) 
				if err != nil {
					log.Printf("create-transaction: failed to update transaction on %s node: %v", n, err)
					ch <- false
					return
				}
 
				ch <- true
				log.Printf("create-transaction: %s", resp.GetStatus())
			}(bc.transactionChan)
		}

		isTransacted = <-bc.transactionChan 
	}

	return isTransacted
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
		if bc.CalculateWalletBalance(senderBlockChainAddress) < value { // this should be checked on the wallet server and frontend and returned to the user
			log.Println("blockchain: Insufficient funds")
			return false
		} 
		bc.MemPool = append(bc.MemPool, t)
		return true
	} 
	return false
}

func (bc *BlockChain) VerifyTransactionSignature(senderPublicKey *ecdsa.PublicKey, s *helpers.Signature, t *transaction.Transaction) bool {
	// m, err := json.Marshal(t)
	m, err := t.MarshalJSON()
	if err != nil {
		log.Printf("blockchain: failed to marshal transaction: %v", err)
		return false
	}
	hash := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublicKey, hash[:], s.R, s.S)
}

func (bc *BlockChain) StartMining() {
	t := time.NewTicker(MINING_TIMER_SEC * time.Second)
	go func(ticker *time.Ticker) {
		for range ticker.C {
			bc.Mining()
		}
	}(t) // add ticker.Stop() during graceful shutdown
	// bc.Mining()
	// _ = time.AfterFunc(MINING_TIMER_SEC*time.Second, bc.Mining)
}

func (bc *BlockChain) Mining() {
	bc.mut.Lock()
	defer bc.mut.Unlock()

	bc.AddTransaction(MINING_SENDER, bc.BlockChainAddress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash
	previousIndex := bc.LastBlock().Index
	bc.CreateBlock(nonce, previousIndex, previousHash)

	bc.wgMining.Add(len(bc.neighbors))
	ctx := context.Background()
	for _, n := range bc.neighbors {
		go func() {
			conn, err := grpc.NewClient(
				n,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				log.Printf("mining: failed to create grpc client on %s node: %v", n, err)
				return
			}
			defer conn.Close()

			client := protogen.NewBlockChainServiceClient(conn)
			resp, err := client.Consensus(ctx, &protogen.Empty{})
			if err != nil {
				log.Printf("mining: consensus failed on %s node: %v", n, err)
				return
			}
			bc.wgMining.Done()
			log.Printf("mining: consensus %s", resp.GetStatus())
		}()
	}
	bc.wgMining.Wait()
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

func (bc *BlockChain) ValidChain(chain []*Block) bool {
	preBlock := chain[0]
	currentIndex := 1

	// genesis block will always be valid
	for currentIndex < len(chain) {
		b := chain[currentIndex]

		if b.Index != (preBlock.Index + 1) {
			log.Printf("invalid index: index 1=>%d index 2=>%d ", b.Index, preBlock.Index+1)
			return false
		}

		if b.PreviousHash != preBlock.Hash {
			log.Printf("invalid hash: hash 1=>%x hash2=>%x ", b.PreviousHash, preBlock.Hash)
			return false
		}

		if !bc.ValidProof(b.Nonce, b.PreviousHash, b.Transactions) {
			return false
		}
		preBlock = b
		currentIndex++
	}
	return true
}

func (bc *BlockChain) ResolveConflicts() bool {
	var longestChain []*Block = nil
	maxLength := len(bc.Chain)
	bc.wgConsensus.Add(len(bc.neighbors))

	ctx := context.Background()
	for _, n := range bc.neighbors {
		go func() {
			conn, err := grpc.NewClient(
				n,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				log.Printf("resolve-conflicts: failed to create grpc client on %s node: %v", n, err)
				return
			}
			defer conn.Close() 
			client := protogen.NewBlockChainServiceClient(conn)
			resp, err := client.GetBlockChain(ctx, &protogen.Empty{})
			if err != nil {
				log.Printf("resolve-conflicts: failed to update transaction on %s node: %v", n, err)
				return
			}
 
			chain, err := bc.convertProtoBlockChain(resp.GetBlockChain())
			if err != nil {
				log.Printf("resolve-conflicts: %v", err)
				return
			} 
			if len(chain) > maxLength && bc.ValidChain(chain) {
				maxLength = len(chain)
				longestChain = chain 
			} 
			bc.wgConsensus.Done()
		}()
	}
	bc.wgConsensus.Wait()
	if longestChain == nil {
		log.Println("resolve conflicts failed")
		return false
	}

	bc.Chain = longestChain
	log.Println("resolve conflicts success")
	return true
}

func (bc *BlockChain) convertProtoBlockChain(blocks []*protogen.Block) ([]*Block, error) {
	blockChain := make([]*Block, 0)
	for _, b := range blocks {
		var hash, previousHash [32]byte
		_, err := hex.Decode(hash[:], []byte(b.GetHash()))
		if err != nil {
			return blockChain, fmt.Errorf("blockchain: failed to convert block hash: %v", err)
		}
		_, err = hex.Decode(previousHash[:], []byte(b.GetPreviousHash()))
		if err != nil {
			return blockChain, fmt.Errorf("blockchain: failed to convert block previous hash: %v", err)
		}

		transactions, err := bc.convertProtoTransactions(b.GetTransactions())
		if err != nil {
			return blockChain, err
		}

		blockChain = append(blockChain, &Block{
			Hash:         hash,
			Nonce:        int(b.GetNonce()),
			Index:        int(b.GetIndex()),
			TimeStamp:    b.GetTimestamp(),
			PreviousHash: previousHash,
			Transactions: transactions,
		})
	}
	return blockChain, nil
}

func (bc *BlockChain) convertProtoTransactions(tx []*protogen.Transaction) ([]*transaction.Transaction, error) {
	transactions := make([]*transaction.Transaction, 0)
	for _, t := range tx {
		var hash [32]byte
		_, err := hex.Decode(hash[:], []byte(t.GetHash()))
		if err != nil {
			return transactions, fmt.Errorf("blockchain: failed to convert transaction hash: %v", err)
		}

		transactions = append(transactions, &transaction.Transaction{
			SenderBlockChainAddress:    t.GetSenderBlockchainAddress(),
			RecipientBlockChainAddress: t.GetRecipientBlockchainAddress(),
			Value:                      t.GetValue(),
			TimeStamp:                  t.GetTimestamp(),
			Hash:                       hash,
		})
	}
	return transactions, nil
}
