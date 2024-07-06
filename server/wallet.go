package server

import (
	"context"

	"github.com/zde37/Zero-Chain/protobuf/protogen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (ws *WalletServer) CreateTransaction(ctx context.Context, req *protogen.TransactionRequest) (*protogen.StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTransaction not implemented")
}
func (ws *WalletServer) CreateWallet(ctx context.Context, req *protogen.Empty) (*protogen.CreateWalletResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWallet not implemented")
}
func (ws *WalletServer) WalletBalance(ctx context.Context, req *protogen.BalanceRequest) (*protogen.BalanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WalletBalance not implemented")
}
