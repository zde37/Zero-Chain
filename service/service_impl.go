package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zde37/Zero-Chain/blockchain"
	"github.com/zde37/Zero-Chain/helpers"
	"github.com/zde37/Zero-Chain/protobuf/protogen"
	"github.com/zde37/Zero-Chain/transaction"
	"github.com/zde37/Zero-Chain/wallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var DB map[string]*blockchain.BlockChain = make(map[string]*blockchain.BlockChain) // in-memory database

type WalletServiceImpl struct {
	port    uint16
	gateway string
	conn    *grpc.ClientConn
	client  protogen.WalletServiceClient
}

type BlockChainServiceImpl struct {
	port uint16
}

func NewWalletServiceImpl(port uint16, gateway string) (WalletService, error) {
	w := &WalletServiceImpl{port: port, gateway: gateway}

	var err error
	w.conn, err = grpc.NewClient(
		w.gateway,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	// defer w.conn.Close() // call this during graceful shutdown

	w.client = protogen.NewWalletServiceClient(w.conn)
	return w, nil
}

func NewBlockChainServiceImpl(port uint16) BlockChainService {
	return &BlockChainServiceImpl{port: port}
}

func (w *WalletServiceImpl) CreateTransaction(ctx context.Context, tr wallet.TransactionRequest) error {
	if !w.validateTransactionRequest(tr) {
		return fmt.Errorf("ERR: create wallet transaction failed due to missing fields")
	}

	publicKey := helpers.PublicKeyFromString(tr.SenderPublicKey)
	privateKey := helpers.PrivateKeyFromString(tr.SenderPrivateKey, publicKey)
	value, err := strconv.ParseFloat(tr.Value, 32)
	if err != nil {
		return fmt.Errorf("ERR: failed to convert string to float")
	}
	value32 := float32(value)
	if tr.SenderBlockchainAddress == tr.RecipientBlockchainAddress {
		return fmt.Errorf("ERR: c'mon man, you can't send z-coin to yourself")
	}
	senderBalance, err := w.GetWalletBalance(ctx, tr.SenderBlockchainAddress)
	if err != nil {
		return fmt.Errorf("ERR: failed to fetch wallet balance: %v", err)
	}
	if senderBalance < value32 {
		return fmt.Errorf("ERR: insufficient funds for this transaction")
	}

	transaction := transaction.NewMetaData(privateKey, publicKey, tr.SenderBlockchainAddress, tr.RecipientBlockchainAddress, value32)
	signature := transaction.GenerateSignature()
	signatureStr := signature.String()

	resp, err := w.client.CreateTransaction(ctx, &protogen.TransactionRequest{
		SenderBlockchainAddress:    tr.SenderBlockchainAddress,
		RecipientBlockchainAddress: tr.RecipientBlockchainAddress,
		SenderPublicKey:            tr.SenderPublicKey,
		Value:                      float32(value),
		Signature:                  signatureStr,
	})
	if err != nil || resp.GetStatus() != "Success" {
		return fmt.Errorf("ERR: failed to create transaction: %v", err)
	}
	return nil
}

func (w *WalletServiceImpl) CreateWallet() wallet.Wallet {
	return *wallet.New()
}

func (w *WalletServiceImpl) GetWalletBalance(ctx context.Context, blockchainAddress string) (float32, error) {
	resp, err := w.client.WalletBalance(ctx, &protogen.BalanceRequest{
		BlockchainAddress: blockchainAddress,
	})
	if err != nil {
		return 0.0, fmt.Errorf("ERR: failed to get wallet balance: %v", err)
	}
	balance, err := strconv.ParseFloat(resp.GetBalance(), 32)
	if err != nil {
		return 0.0, fmt.Errorf("ERR: failed to convert string to float")
	}
	balance32 := float32(balance)
	return balance32, nil
}

func (w *WalletServiceImpl) validateTransactionRequest(tr wallet.TransactionRequest) bool {
	if tr.SenderPrivateKey == "" ||
		tr.SenderPublicKey == "" ||
		tr.SenderBlockchainAddress == "" ||
		tr.RecipientBlockchainAddress == "" ||
		tr.Value == "" {
		return false
	}
	return true
}

func (b *BlockChainServiceImpl) getBlockchain() *blockchain.BlockChain {
	bc, ok := DB["blockchain"] // check if blockchain already exists
	if !ok {
		minersWallet := wallet.New()
		bc = blockchain.New(minersWallet.BlockchainAddress, b.port)
		DB["blockchain"] = bc
		fmt.Printf("private_key %s\n", minersWallet.PrivateKeyStr())
		fmt.Printf("public_key %s\n", minersWallet.PublicKeyStr())
		fmt.Printf("blockchain_address %s\n", minersWallet.BlockchainAddress)
	}
	return bc
}

func (b *BlockChainServiceImpl) CreateTransaction(t transaction.Request) error {
	if !b.validateTransaction(t) {
		return fmt.Errorf("ERR: create transaction failed due to missing fields")
	}

	publicKey := helpers.PublicKeyFromString(t.SenderPublicKey)
	signature := helpers.SignatureFromString(t.Signature)
	bc := b.getBlockchain()
	isCreated := bc.CreateTransaction(t.SenderBlockchainAddress, t.RecipientBlockchainAddress, t.Value, publicKey, signature)

	if !isCreated {
		return fmt.Errorf("ERR: failed to create transaction")
	}
	return nil
}

func (b *BlockChainServiceImpl) UpdateTransaction(t transaction.Request) error {
	if !b.validateTransaction(t) {
		return fmt.Errorf("ERR: update transaction failed due to missing fields")
	}

	publicKey := helpers.PublicKeyFromString(t.SenderPublicKey)
	signature := helpers.SignatureFromString(t.Signature)
	bc := b.getBlockchain()
	isUpdated := bc.AddTransaction(t.SenderBlockchainAddress, t.RecipientBlockchainAddress, t.Value, publicKey, signature)

	if !isUpdated {
		return fmt.Errorf("ERR: failed to update transaction")
	}
	return nil
}

func (b *BlockChainServiceImpl) ListTransactions() ([]*transaction.Transaction, int) {
	bc := b.getBlockchain()
	return bc.MemPool, len(bc.MemPool)
}

func (b *BlockChainServiceImpl) DeleteTransactions() error {
	bc := b.getBlockchain()
	bc.ClearMemPool()
	return nil
}

func (b *BlockChainServiceImpl) Consensus() error {
	bc := b.getBlockchain()
	replaced := bc.ResolveConflicts()
	if !replaced {
		return fmt.Errorf("ERR: failed to resolve conflicts")
	}
	return nil
}

func (b *BlockChainServiceImpl) GetBlockChain() []*blockchain.Block {
	return b.getBlockchain().Chain
}

func (b *BlockChainServiceImpl) GetWalletBalance(blockchainAddress string) float32 {
	bc := b.getBlockchain()
	return bc.CalculateWalletBalance(blockchainAddress)
}

func (b *BlockChainServiceImpl) validateTransaction(t transaction.Request) bool {
	if &t.Signature == nil ||
		&t.SenderPublicKey == nil ||
		&t.SenderBlockchainAddress == nil ||
		&t.RecipientBlockchainAddress == nil ||
		&t.Value == nil {
		return false
	}
	return true
}
