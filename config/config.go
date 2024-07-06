package config

type Config struct {
	WalletGrpcServerAddr        string
	WalletGatewayServerAddr     string
	BlockChainGrpcServerAddr    string
	BlockChainGatewayServerAddr string
}

func LoadConfig(
	walletGrpcServerAddr,
	walletGatewayServerAddr,
	blockChainGrpcServerAddr,
	blockChainGatewayServerAddr string) (Config, error) {
	return Config{
		WalletGrpcServerAddr:        walletGrpcServerAddr,
		WalletGatewayServerAddr:     walletGatewayServerAddr,
		BlockChainGrpcServerAddr:    blockChainGrpcServerAddr,
		BlockChainGatewayServerAddr: blockChainGatewayServerAddr,
	}, nil
}
