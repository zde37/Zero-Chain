# High-Performance Blockchain Network

## Overview
This project is a high-performance blockchain network developed using Golang. The network supports multiple nodes, uses a secure Proof-of-Work (PoW) consensus mechanism, ECC for key generation, and SHA-256 for block hashing.

## Features
- Supports multiple nodes for enhanced scalability and performance.
- Secure PoW consensus mechanism for block validation.
- Utilizes ECC for secure key generation and SHA-256 for block hashing.
- Users can create wallets, securely generate public/private key pairs and participate in mining activities.
- The network can facilitate Y transactions per second, optimized for real-time usability.
- Seamless interaction with the blockchain network through an integrated blockchain explorer for real-time visualization of the blockchain ledger.
- GRPC and Protobuf for efficient internal service communication and secure gateway for external communication.


## Architecture

### Core Components

1. **Blockchain Service**
   - Handles core blockchain operations
   - Implements Proof-of-Work (PoW) consensus
   - Manages block creation and validation
   - Handles peer-to-peer network synchronization
   - Ports:
     - gRPC Server: 7000 (default)
     - Gateway/HTTP Server: 7070 (default)

2. **Wallet Service**
   - Manages wallet operations
   - Handles key pair generation using ECC
   - Manages transaction creation and signing
   - Ports:
     - gRPC Server: 5000 (default)
     - Gateway/HTTP Server: 5050 (default)

### Technical Features
- **Consensus**: Proof-of-Work (PoW) mechanism
- **Cryptography**: 
  - ECC for key generation
  - SHA-256 for block hashing
- **Network Protocol**: 
  - gRPC for internal service communication
  - REST API gateway for external access
- **Data Storage**: Persistent blockchain storage
- **User Interface**: Web-based blockchain explorer and transaction viewer

## API Endpoints

### Blockchain Service
- **Gateway Server** (default: 7070)
  - `/explorer` - Blockchain explorer UI
  - `/transactions` - Transaction viewer UI
  - `/hello-world` - Health check endpoint
  - `/` - REST API endpoint

### Wallet Service
- **Gateway Server** (default: 5050)
  - `/index` - Wallet management UI
  - `/hello-world` - Health check endpoint
  - `/` - REST API endpoint


## Running the Project
#### NOTE 
- `bch-grpc` port must be between 7000 and 7003
- `wal-grpc` port must be between 5000 and 5003
- `bch-gateway` port must be between 5050 and 5053
-  the terms "neighbors" and "peers" are used interchangeably in the context of the project.

### To run the project, follow these steps:

1. First, clone the repository:
```bash
git clone github.com/zde37/Zero-Chain
```

2. Navigate to the project directory:
```bash
cd Zero-Chain
```

3. install the dependencies:
```bash
go mod tidy
```

4. Spin up the first node with default settings:
```bash
go run main.go
```

5. Spin up other nodes:
```bash
go run main.go --bch-grpc=<USE_PORT_OF_CHOICE> --wal-grpc=<USE_PORT_OF_CHOICE> --wal-gateway=<USE_PORT_OF_CHOICE>
```

#### Available command-line flags:

- --bch-grpc: Blockchain gRPC server port (default: 7000)
- --bch-gateway: Blockchain HTTP/Gateway server port (default: 7070)
- --bch-host: Blockchain server host (default: 127.0.0.1)
- --wal-grpc: Wallet gRPC server port (default: 5000)
- --wal-gateway: Wallet HTTP/Gateway server port (default: 5050)

#### Once running, you can access:

- Blockchain Explorer UI: http://localhost:7070/explorer
- Transactions UI: http://localhost:7070/transactions
- Wallet UI: http://localhost:5050/index
- Health check endpoints: http://localhost:7070/hello-world and http://- localhost:5050/hello-world

#### The system will automatically:

- Start the blockchain service
- Begin syncing with neighbor nodes
- Resolve any blockchain conflicts
- Start the mining process
- Make all services available via both gRPC and REST APIs