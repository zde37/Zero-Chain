package service

import (
	"context"

	"github.com/zde37/Zero-Chain/blockchain"
	"github.com/zde37/Zero-Chain/transaction"
	"github.com/zde37/Zero-Chain/wallet"
)

type WalletService interface {
	CreateTransaction(ctx context.Context, tr wallet.TransactionRequest) error
	CreateWallet() (*wallet.Wallet, error)
	GetWalletBalance(ctx context.Context, blockchainAddress string) (float32, error)
}

type BlockChainService interface {
	CreateTransaction(ctx context.Context, t transaction.Request) error
	UpdateTransaction(t transaction.Request) error
	ListTransactions() ([]*transaction.Transaction, int)
	DeleteTransactions() error
	Consensus() error
	Run()  
	GetBlockChain() []*blockchain.Block
	GetWalletBalance(blockchainAddress string) float32
}
