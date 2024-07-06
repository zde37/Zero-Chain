package server

import (
	"context"

	"github.com/zde37/Zero-Chain/protobuf/protogen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (bcs *BlockChainServer) ListTransactions(ctx context.Context, req *protogen.Empty) (*protogen.ListTransactionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTransactions not implemented")
}
func (bcs *BlockChainServer) GetBlockChain(ctx context.Context, req *protogen.Empty) (*protogen.GetBlockChainResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlockChain not implemented")
}
func (bcs *BlockChainServer) WalletBalance(ctx context.Context, req *protogen.BalanceRequest) (*protogen.BalanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WalletBalance not implemented")
}
func (bcs *BlockChainServer) CreateTransaction(ctx context.Context, req *protogen.TransactionRequest) (*protogen.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTransaction not implemented")
}
func (bcs *BlockChainServer) UpdateTransaction(ctx context.Context, req *protogen.TransactionRequest) (*protogen.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTransaction not implemented")
}
func (bcs *BlockChainServer) DeleteTransaction(ctx context.Context, req *protogen.Empty) (*protogen.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteTransaction not implemented")
}
func (bcs *BlockChainServer) Consensus(ctx context.Context, req *protogen.Empty) (*protogen.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Consensus not implemented")
}
