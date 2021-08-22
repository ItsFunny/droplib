package batch

import (
	crypto2 "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/crypto"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
)

// CreateBatchVerifier checks if a key type implements the batch verifier interface.
// Currently only ed25519 & sr25519 supports batch verification.
func CreateBatchVerifier(pk cryptolibs.IPublicKey) (crypto2.BatchVerifier, bool) {
	switch pk.Type() {
	// case ed25519.KeyType:
	// 	return ed25519.NewBatchVerifier(), true
	// case sr25519.KeyType:
	// 	return sr25519.NewBatchVerifier(), true
	default:
		panic("123")
	}

	// case where the key does not support batch verification
	return nil, false
}
