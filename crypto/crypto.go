package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"

	"github.com/diode_go_client/crypto/secp256k1"
	"github.com/diode_go_client/crypto/sha3"
)

const ecPrivKeyVersion = 1

var (
	secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

var errInvalidPubkey = errors.New("invalid secp256k1 public key")

// ECPEMPrivateKey openssl ec pem private key
// see (maybe): https://tls.mbed.org/kb/cryptography/asn1-key-structures-in-der-and-pem
type ECPEMPrivateKey struct {
	E1      *big.Int
	KeyInfo struct {
		KeyType   asn1.ObjectIdentifier
		CurveName asn1.ObjectIdentifier
	}
	KeyData []byte
}

// ECPublicKey ec public key format
// see (maybe): https://tls.mbed.org/kb/cryptography/asn1-key-structures-in-der-and-pem
type ECPublicKey struct {
	Algorithm struct {
		Algorithm  asn1.ObjectIdentifier
		Parameters []byte `asn1:"optional"`
	}
	PublicKey asn1.BitString
}

// ECPrivateKey reflects an ASN.1 Elliptic Curve Private Key Structure.
// References:
//   RFC 5915
//   SEC1 - http://www.secg.org/sec1-v2.pdf
// Per RFC 5915 the NamedCurveOID is marked as ASN.1 OPTIONAL, however in
// most cases it is not.
type ECPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
}

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return secp256k1.S256()
}

// PemToECDSA creates a private key with the given openssl pem encoded value.
// TODO: check key type and curve name
func PemToECDSA(pem []byte) (*ecdsa.PrivateKey, error) {
	var privKey ECPEMPrivateKey
	if _, err := asn1.Unmarshal(pem, &privKey); err != nil {
		return nil, err
	}
	return DerToECDSA(privKey.KeyData)
}

// DerToECDSA creates a private key with the given der encoded D value.
func DerToECDSA(derD []byte) (*ecdsa.PrivateKey, error) {
	var privKey ECPrivateKey
	if _, err := asn1.Unmarshal(derD, &privKey); err != nil {
		return nil, errors.New("x509: failed to parse EC private key: " + err.Error())
	}
	if privKey.Version != ecPrivKeyVersion {
		return nil, fmt.Errorf("x509: unknown EC private key version %d", privKey.Version)
	}
	return toECDSA(privKey.PrivateKey, true)
}

// ToECDSA creates a private key with the given D value.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	return toECDSA(d, true)
}

// ToECDSAUnsafe blindly converts a binary blob to a private key. It should almost
// never be used unless you are sure the input is valid and want to avoid hitting
// errors due to bad origin encoding (0 prefixes cut off).
func ToECDSAUnsafe(d []byte) *ecdsa.PrivateKey {
	priv, _ := toECDSA(d, false)
	return priv
}

// DerToPublicKey returns uncompressed ecdsa public key bytes
func DerToPublicKey(derPubKey []byte) ([]byte, error) {
	pubKey := make([]byte, 1)
	var ecPubKey ECPublicKey
	if _, err := asn1.Unmarshal(derPubKey, &ecPubKey); err != nil {
		return pubKey, err
	}
	// uncompress
	return ecPubKey.PublicKey.Bytes, nil
}

// toECDSA creates a private key with the given D value. The strict parameter
// controls whether the key's length should be enforced at the curve size or
// it can also accept legacy encodings (0 prefixes).
func toECDSA(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}

// UnmarshalPubkey converts bytes to a secp256k1 public key.
func UnmarshalPubkey(pub []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(S256(), pub)
	if x == nil {
		return nil, errInvalidPubkey
	}
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y)
}

// Sha256 the data
func Sha256(data []byte) []byte {
	sha256 := sha256.New()
	sha256.Write(data)
	return sha256.Sum(nil)
}

// Sha3 the data
func Sha3(data []byte) []byte {
	hash := sha3.NewKeccak256()
	hash.Write(append(data))
	return hash.Sum(nil)
}

func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}

// PubkeyToAddress returns diode address
func PubkeyToAddress(pubkey []byte) ([]byte, error) {
	hasher := sha3.NewKeccak256()
	hasher.Write(pubkey[1:])
	hashPubkey := hasher.Sum(nil)
	return hashPubkey[12:], nil
}
