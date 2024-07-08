package main

import (
	"flag"
	"fmt"
	"log"

	_ "github.com/joho/godotenv/autoload"
	"github.com/zde37/Zero-Chain/config"
	"github.com/zde37/Zero-Chain/server"
	"github.com/zde37/Zero-Chain/service"
)

func main() {
	blockchainGRPCPort := flag.Uint("bch-grpc", 7000, "blockchain grpc server port") // the range is from 7000-7003. Adjust according to your needs
	blockchainGatewayPort := flag.Uint("bch-gateway", 7070, "blockchain gateway server port")
	host := flag.String("bch-host", "127.0.0.1", "blockchain server host")
	walletGRPCPort := flag.Uint("wal-grpc", 5000, "wallet grpc server port")
	walletGatewayPort := flag.Uint("wal-gateway", 5050, "wallet gateway server port")
	flag.Parse()

	config := config.LoadConfig(fmt.Sprintf("0.0.0.0:%d", *walletGRPCPort), fmt.Sprintf("0.0.0.0:%d", *walletGatewayPort),
		fmt.Sprintf("%s:%d", *host, *blockchainGRPCPort), fmt.Sprintf("%s:%d", *host, *blockchainGatewayPort))

	blockchainService := service.NewBlockChainServiceImpl(uint16(*blockchainGRPCPort))
	walletService, err := service.NewWalletServiceImpl(uint16(*walletGRPCPort), fmt.Sprintf("%s:%d", *host, *blockchainGRPCPort))
	if err != nil {
		log.Fatalf("failed to create wallet service: %v", err)
	}

	bcGRPCServer := server.NewBlockChainServer(blockchainService, config)
	walletGRPCServer := server.NewWalletServer(walletService, config)

	go bcGRPCServer.RunGatewayServer()
	go bcGRPCServer.RunGrpcServer()

	// go walletGRPCServer.RunGrpcServer()
	walletGRPCServer.RunGatewayServer()
}
