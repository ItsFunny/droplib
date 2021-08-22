package ed25519

//-------------------------------------


const (
	PrivKeyName = "tendermint/PrivKeyEd25519"
	PubKeyName  = "tendermint/PubKeyEd25519"
	// PubKeySize is is the size, in bytes, of public keys as used in this package.
	PubKeySize = 32
	// PrivateKeySize is the size, in bytes, of private keys as used in this package.
	PrivateKeySize = 64
	// Size of an Edwards25519 signature. Namely the size of a compressed
	// Edwards25519 point, and a field element. Both of which are 32 bytes.
	SignatureSize = 64
	// SeedSize is the size, in bytes, of private key seeds. These are the
	// private key representations used by RFC 8032.
	SeedSize = 32

	KeyType = "ed25519"
)
func init() {
	// tmjson.RegisterType(PubKey{}, PubKeyName)
	// tmjson.RegisterType(PrivKey{}, PrivKeyName)
}
