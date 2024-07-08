package service

import (
	"context"
	"fmt" 
	"strconv"
	"strings"

	"github.com/zde37/Zero-Chain/blockchain"
	"github.com/zde37/Zero-Chain/helpers"
	"github.com/zde37/Zero-Chain/protobuf/protogen"
	"github.com/zde37/Zero-Chain/transaction"
	"github.com/zde37/Zero-Chain/wallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	DB           map[string]*blockchain.BlockChain = make(map[string]*blockchain.BlockChain) // in-memory database
	minersWallet map[uint16]*wallet.Wallet         = make(map[uint16]*wallet.Wallet)
)

type WalletServiceImpl struct {
	port    uint16
	gateway string
	conn    *grpc.ClientConn
	client  protogen.BlockChainServiceClient
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

	w.client = protogen.NewBlockChainServiceClient(w.conn)
	return w, nil
}

func NewBlockChainServiceImpl(port uint16) BlockChainService {
	return &BlockChainServiceImpl{port: port}
}

func (w *WalletServiceImpl) CreateTransaction(ctx context.Context, tr wallet.TransactionRequest) error {
	publicKey := helpers.PublicKeyFromString(tr.SenderPublicKey)
	privateKey := helpers.PrivateKeyFromString(tr.SenderPrivateKey, publicKey)
	if tr.SenderBlockchainAddress == tr.RecipientBlockchainAddress {
		return fmt.Errorf("ERR: c'mon man, you can't send z-coin to yourself")
	}
	senderBalance, err := w.GetWalletBalance(ctx, tr.SenderBlockchainAddress)
	if err != nil {
		return fmt.Errorf("ERR: failed to fetch wallet balance: %v", err)
	}
	if senderBalance < tr.Value {
		return fmt.Errorf("ERR: insufficient funds for this transaction")
	}

	transaction := transaction.NewMetaData(privateKey, publicKey, tr.SenderBlockchainAddress, tr.RecipientBlockchainAddress, tr.Value)
	signature := transaction.GenerateSignature()
	signatureStr := signature.String()

	resp, err := w.client.CreateTransaction(ctx, &protogen.TransactionRequest{
		SenderBlockchainAddress:    tr.SenderBlockchainAddress,
		RecipientBlockchainAddress: tr.RecipientBlockchainAddress,
		SenderPublicKey:            tr.SenderPublicKey,
		Value:                      tr.Value,
		Signature:                  signatureStr,
	})
	if err != nil || resp.GetStatus() != "Success" {
		return fmt.Errorf("ERR: failed to create transaction: %v", err)
	}
	return nil
}

func (w *WalletServiceImpl) CreateWallet() (*wallet.Wallet, error) {
	val := strings.Split(w.gateway, ":")

	port, err := strconv.ParseUint(val[1], 10, 16)
	if err != nil {
		return nil, fmt.Errorf("create-wallet: failed to parse port number%v", err)
	}
	return getWallet(uint16(port)), nil
}

func (w *WalletServiceImpl) GetWalletBalance(ctx context.Context, blockchainAddress string) (float32, error) {
	resp, err := w.client.WalletBalance(ctx, &protogen.BalanceRequest{
		BlockchainAddress: blockchainAddress,
	})
	if err != nil {
		return 0, fmt.Errorf("ERR: failed to get wallet balance: %v", err)
	}
	return resp.GetBalance(), nil
}

func getWallet(port uint16) *wallet.Wallet {
	w, ok := minersWallet[port]
	if !ok {
		w = wallet.New()
		minersWallet[port] = w
	}
	return w
}

func (b *BlockChainServiceImpl) getBlockchain() *blockchain.BlockChain {
	bc, ok := DB["blockchain"] // check if blockchain already exists
	if !ok {
		minersWallet := getWallet(b.port)
		bc = blockchain.New(minersWallet.BlockchainAddress, b.port)
		DB["blockchain"] = bc
	}
	return bc
}

func (b *BlockChainServiceImpl) Run() {
	b.getBlockchain().Run()
}

func (b *BlockChainServiceImpl) CreateTransaction(ctx context.Context, t transaction.Request) error {
	publicKey := helpers.PublicKeyFromString(t.SenderPublicKey)
	signature := helpers.SignatureFromString(t.Signature)
	bc := b.getBlockchain()
	isCreated := bc.CreateTransaction(ctx, t.SenderBlockchainAddress, t.RecipientBlockchainAddress, t.Value, publicKey, signature)
 
	if !isCreated {
		return fmt.Errorf("ERR: failed to create transaction")
	}
 	return nil
}

func (b *BlockChainServiceImpl) UpdateTransaction(t transaction.Request) error {
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
