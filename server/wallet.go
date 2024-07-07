package server

import (
	"context"

	"github.com/zde37/Zero-Chain/protobuf/protogen"
	"github.com/zde37/Zero-Chain/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (ws *WalletServer) CreateTransaction(ctx context.Context, req *protogen.WalletTransactionRequest) (*protogen.StatusResponse, error) {
	if !ws.validateTransactionRequest(req) {
		return nil, status.Errorf(codes.InvalidArgument, "failed due to missing fields")
	}

	if err := ws.walletService.CreateTransaction(ctx, wallet.TransactionRequest{
		SenderPrivateKey:           req.GetSenderPrivateKey(),
		SenderBlockchainAddress:    req.GetSenderBlockchainAddress(),
		RecipientBlockchainAddress: req.GetRecipientBlockchainAddress(),
		SenderPublicKey:            req.GetSenderPublicKey(),
		Value:                      req.GetValue(),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &protogen.StatusResponse{
		Status: "Success",
	}, nil
}

func (ws *WalletServer) CreateWallet(ctx context.Context, req *protogen.Empty) (*protogen.CreateWalletResponse, error) {
	wallet := ws.walletService.CreateWallet()

	return &protogen.CreateWalletResponse{
		PrivateKey:        wallet.PrivateKeyStr(),
		PublicKey:         wallet.PublicKeyStr(),
		BlockchainAddress: wallet.BlockchainAddress,
	}, nil
}

func (ws *WalletServer) WalletBalance(ctx context.Context, req *protogen.BalanceRequest) (*protogen.BalanceResponse, error) {
	if req.GetBlockchainAddress() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "blockchain address is required")
	}

	balance, err := ws.walletService.GetWalletBalance(ctx, req.GetBlockchainAddress())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &protogen.BalanceResponse{
		Balance: balance,
	}, nil
}

func (ws *WalletServer) validateTransactionRequest(req *protogen.WalletTransactionRequest) bool {
	if req.GetSenderPrivateKey() == "" ||
		req.GetSenderPublicKey() == "" ||
		req.GetSenderBlockchainAddress() == "" ||
		req.GetRecipientBlockchainAddress() == "" ||
		req.GetValue() == 0 {
		return false
	}
	return true
}
