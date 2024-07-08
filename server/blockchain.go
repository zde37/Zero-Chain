package server

import (
	"context"
	"fmt"

	"github.com/zde37/Zero-Chain/blockchain"
	"github.com/zde37/Zero-Chain/protobuf/protogen"
	"github.com/zde37/Zero-Chain/transaction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (bcs *BlockChainServer) ListTransactions(ctx context.Context, req *protogen.Empty) (*protogen.ListTransactionsResponse, error) {
	transactions, length := bcs.blockChainService.ListTransactions()

	return &protogen.ListTransactionsResponse{
		Transactions: bcs.convertTransactions(transactions),
		Length:       int64(length),
	}, nil
}

func (bcs *BlockChainServer) GetBlockChain(ctx context.Context, req *protogen.Empty) (*protogen.GetBlockChainResponse, error) {
	blockchain := bcs.blockChainService.GetBlockChain()

	return &protogen.GetBlockChainResponse{
		BlockChain: bcs.convertBlockChain(blockchain),
	}, nil
}

func (bcs *BlockChainServer) WalletBalance(ctx context.Context, req *protogen.BalanceRequest) (*protogen.BalanceResponse, error) {
	if req.GetBlockchainAddress() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "blockchain address is required")
	}
	balance := bcs.blockChainService.GetWalletBalance(req.GetBlockchainAddress())

	return &protogen.BalanceResponse{
		Balance: balance,
	}, nil
}

func (bcs *BlockChainServer) CreateTransaction(ctx context.Context, req *protogen.TransactionRequest) (*protogen.StatusResponse, error) {
	if !bcs.validateTransaction(req) {
		return nil, status.Errorf(codes.InvalidArgument, "failed due to missing fields")
	}

	if err := bcs.blockChainService.CreateTransaction(ctx, transaction.Request{
		SenderBlockchainAddress:    req.GetSenderBlockchainAddress(),
		RecipientBlockchainAddress: req.GetRecipientBlockchainAddress(),
		SenderPublicKey:            req.GetSenderPublicKey(),
		Value:                      req.GetValue(),
		Signature:                  req.GetSignature(),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &protogen.StatusResponse{
		Status: "Success",
	}, nil
}

func (bcs *BlockChainServer) UpdateTransaction(ctx context.Context, req *protogen.TransactionRequest) (*protogen.StatusResponse, error) {

	if !bcs.validateTransaction(req) {
		return nil, status.Errorf(codes.InvalidArgument, "failed due to missing fields")
	}

	if err := bcs.blockChainService.UpdateTransaction(transaction.Request{
		SenderBlockchainAddress:    req.GetSenderBlockchainAddress(),
		RecipientBlockchainAddress: req.GetRecipientBlockchainAddress(),
		SenderPublicKey:            req.GetSenderPublicKey(),
		Value:                      req.GetValue(),
		Signature:                  req.GetSignature(),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &protogen.StatusResponse{
		Status: "Success",
	}, nil
}

func (bcs *BlockChainServer) DeleteTransaction(ctx context.Context, req *protogen.Empty) (*protogen.StatusResponse, error) {
	if err := bcs.blockChainService.DeleteTransactions(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete transactions")
	}
	return &protogen.StatusResponse{
		Status: "Success",
	}, nil
}

func (bcs *BlockChainServer) Consensus(ctx context.Context, req *protogen.Empty) (*protogen.StatusResponse, error) {
	if err := bcs.blockChainService.Consensus(); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &protogen.StatusResponse{
		Status: "Success",
	}, nil
}

func (bcs *BlockChainServer) convertBlockChain(bc []*blockchain.Block) []*protogen.Block {
	blockchain := make([]*protogen.Block, 0)
	for _, b := range bc {
		blockchain = append(blockchain, &protogen.Block{
			Nonce:        int64(b.Nonce),
			Index:        int64(b.Index),
			PreviousHash: fmt.Sprintf("%x", b.PreviousHash),
			Timestamp:    b.TimeStamp,
			Hash:         fmt.Sprintf("%x", b.Hash),
			Transactions: bcs.convertTransactions(b.Transactions),
		})
	}
	return blockchain
}

func (bcs *BlockChainServer) convertTransactions(tx []*transaction.Transaction) []*protogen.Transaction {
	transactions := make([]*protogen.Transaction, 0)
	for _, t := range tx {
		transactions = append(transactions, &protogen.Transaction{
			SenderBlockchainAddress:    t.SenderBlockChainAddress,
			RecipientBlockchainAddress: t.RecipientBlockChainAddress,
			Value:                      t.Value,
			Hash:                       fmt.Sprintf("%x", t.Hash),
			Timestamp:                  t.TimeStamp,
		})
	}
	return transactions
}

func (bcs *BlockChainServer) validateTransaction(tr *protogen.TransactionRequest) bool {
	if tr.GetSignature() == "" ||
		tr.GetSenderPublicKey() == "" ||
		tr.GetSenderBlockchainAddress() == "" ||
		tr.GetRecipientBlockchainAddress() == "" ||
		tr.GetValue() == 0 {
		return false
	}
	return true
}
