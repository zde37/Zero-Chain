package server

import (
	"context"
	"html/template"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/zde37/Zero-Chain/config"
	"github.com/zde37/Zero-Chain/protobuf/protogen"
	"github.com/zde37/Zero-Chain/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

type BlockChainServer struct {
	protogen.UnimplementedBlockChainServiceServer
	blockChainService service.BlockChainService
	grpcServer        *grpc.Server
	config            config.Config
}

type WalletServer struct {
	protogen.UnimplementedWalletServiceServer
	walletService service.WalletService
	grpcServer    *grpc.Server
	config        config.Config
}

func NewWalletServer(walletService service.WalletService, config config.Config) *WalletServer {
	return &WalletServer{
		walletService: walletService,
		config:        config,
	}
}

func NewBlockChainServer(blockChainService service.BlockChainService, config config.Config) *BlockChainServer {
	return &BlockChainServer{
		blockChainService: blockChainService,
		config:            config,
	}
}

func (bcs *BlockChainServer) RunGrpcServer() {
	grpcServer := grpc.NewServer()
	bcs.grpcServer = grpcServer

	protogen.RegisterBlockChainServiceServer(grpcServer, bcs)
	reflection.Register(grpcServer) // self-documentation for the server

	listener, err := net.Listen("tcp", bcs.config.BlockChainGrpcServerAddr)
	if err != nil {
		log.Printf("server: failed to create listener for blockchain gRPC server: %v", err)
		return
	}
	defer listener.Close()

	log.Printf("server: blockchain gRPC server started on: %s", bcs.config.BlockChainGrpcServerAddr)
	go bcs.blockChainService.Run()
	if err = grpcServer.Serve(listener); err != nil {
		log.Printf("server: failed to start blockchain gRPC server: %v", err)
		return
	}
}

func (bcs *BlockChainServer) RunGatewayServer() {
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions:   protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true},
		UnmarshalOptions: protojson.UnmarshalOptions{DiscardUnknown: true},
	})
	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := protogen.RegisterBlockChainServiceHandlerServer(ctx, grpcMux, bcs); err != nil {
		log.Printf("server: failed to register blockchain http handlers: %v", err)
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)
	mux.Handle("/hello-world", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // health route
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	}))
	mux.Handle("/explorer", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { 
		var (
			err  error
			once sync.Once
			tpl  *template.Template
		)

		once.Do(func() {
			tpl, err = template.ParseFiles("./templates/explorer.html")
		})
		if err != nil {
			log.Fatalf("failed to parse templates: %v", err)
		}
		tpl.Execute(w, "")
	}))
	mux.Handle("/transactions", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { 
		var (
			err  error
			once sync.Once
			tpl  *template.Template
		)

		once.Do(func() {
			tpl, err = template.ParseFiles("./templates/transactions.html")
		})
		if err != nil {
			log.Fatalf("failed to parse template: %v", err)
		}
		tpl.Execute(w, "")
	}))
	httpServer := &http.Server{
		Handler: mux,
		Addr:    bcs.config.BlockChainGatewayServerAddr,
	}

	log.Printf("server: blockchain gateway server started on: %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Printf("server: failed start blockchain gateway server: %v", err)
		return
	}
}

func (s *BlockChainServer) StopGrpcServer() {
	s.grpcServer.GracefulStop()
}

func (s *WalletServer) RunGrpcServer() {
	grpcServer := grpc.NewServer()
	s.grpcServer = grpcServer

	protogen.RegisterWalletServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", s.config.WalletGrpcServerAddr)
	if err != nil {
		log.Printf("server: failed to create listener for wallet gRPC server: %v", err)
		return
	}
	defer listener.Close()

	log.Printf("server: wallet gRPC server started on: %s", s.config.WalletGrpcServerAddr)
	if err = grpcServer.Serve(listener); err != nil {
		log.Printf("server: failed to start wallet gRPC server: %v", err)
		return
	}
}

func (s *WalletServer) RunGatewayServer() {
	// set json response to use snake case
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	grpcMux := runtime.NewServeMux(jsonOption)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := protogen.RegisterWalletServiceHandlerServer(ctx, grpcMux, s); err != nil {
		log.Printf("server: failed to register wallet http handlers: %v", err)
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)
	mux.Handle("/hello-world", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // health route
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	}))
	mux.Handle("/index", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { 
		var (
			err  error
			once sync.Once
			tpl  *template.Template
		)

		once.Do(func() {
			tpl, err = template.ParseFiles("./templates/index.html")
		})
		if err != nil {
			log.Fatalf("failed to parse templates: %v", err)
		}
		tpl.Execute(w, "")
	}))

	httpServer := &http.Server{
		Handler: mux,
		Addr:    s.config.WalletGatewayServerAddr,
	}

	log.Printf("server: wallet gateway server started on: %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Printf("server: failed start wallet gateway server: %v", err)
		return
	}
}

func (s *WalletServer) StopGrpcServer() {
	s.grpcServer.GracefulStop()
}
