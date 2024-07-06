package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	PrivateKey        *ecdsa.PrivateKey
	PublicKey         *ecdsa.PublicKey
	BlockchainAddress string
}

type TransactionRequest struct {
	SenderPrivateKey           string
	SenderBlockchainAddress    string
	RecipientBlockchainAddress string
	SenderPublicKey            string
	Value                      float32
}

func New() *Wallet {
	w := new(Wallet)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Printf("wallet: failed to generate private key: %v", err)
		return nil
	}
	w.PrivateKey = privateKey
	w.PublicKey = &w.PrivateKey.PublicKey

	h2 := sha256.New()
	h2.Write(w.PublicKey.X.Bytes())
	h2.Write(w.PublicKey.Y.Bytes())
	digest2 := h2.Sum(nil)

	h3 := ripemd160.New()
	h3.Write(digest2)
	digest3 := h3.Sum(nil)

	vd4 := make([]byte, 21)
	vd4[0] = 0x00
	copy(vd4[1:], digest3[:])

	h5 := sha256.New()
	h5.Write(vd4)
	digest5 := h5.Sum(nil)

	h6 := sha256.New()
	h6.Write(digest5)
	digest6 := h6.Sum(nil)

	checkSum := digest6[:4]

	dc8 := make([]byte, 25)
	copy(dc8[:21], vd4[:])
	copy(dc8[21:], checkSum[:])

	address := base58.Encode(dc8)
	w.BlockchainAddress = address

	return w
}

func (w *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x", w.PrivateKey.D.Bytes())
}

func (w *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%064x%064x", w.PublicKey.X.Bytes(), w.PublicKey.Y.Bytes())
}
