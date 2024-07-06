package transaction

type Request struct {
	SenderBlockchainAddress    string
	RecipientBlockchainAddress string
	SenderPublicKey            string
	Value                      float32
	Signature                  string
}
