package main

import (
	"flag"
	"log"

	"github.com/zde37/Zero-Chain/blockchain"
)

func main() {
	port := flag.Uint("port", 4000, "blockchain server port")
	flag.Parse()

	bc := blockchain.New("Sender-A", uint16(*port))
	log.Println(bc.Chain)
	log.Println(bc.MemPool)
	bc.AddTransaction("A", "B", 2)
	bc.AddTransaction("E", "F", 4)
	log.Println(bc.MemPool)

	bc.Mining()
	log.Println(bc.Chain)
	log.Println(bc.MemPool)
	bc.AddTransaction("Y", "Z", 1)
	log.Println(bc.MemPool)
	
	bc.Mining()
	log.Println(bc.Chain)
	log.Println(bc.MemPool)
}
