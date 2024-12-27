package keymanagement

import (
	"encoding/base64"
	"errors"
	"github.com/open-quantum-safe/liboqs-go/oqs"
github.com/theaxiomverse/hydap-api/pkg/crypto"
github.com/theaxiomverse/hydap-api/pkg/keymanagement/pb"
)
type KeyManagement interface {
	GetPublicKey() string
	LoadSecretKey(string) error
	DeriveKey() string
	Init(algorithm pb.Algorithm, secretKey string) error
	Sign([]byte) ([]byte, error)
	GetPrivate() []byte
}

type keygen struct {
	publicKey  []byte
	privateKey []byte
	alg        pb.Algorithm
}

func (k *keygen) Init(algorithm pb.Algorithm, secretKey string) error {
	k.alg = algorithm
	if algorithm == pb.Algorithm_NONE {
		return ErrUnsupportedAlgorithm
	}

	if secretKey != "" {
		if err := k.LoadSecretKey(secretKey); err != nil {
			return err
		}
	}

	// Generate new keypair if no secret key provided
	pk, sk, err := k.generateKeyPair()
	if err != nil {
		return err
	}

	k.publicKey = pk
	if secretKey == "" {
		k.privateKey = sk
	}

	return nil
}

func (k *keygen) GetPublicKey() string {
	return base64.StdEncoding.EncodeToString(k.publicKey)
}

func (k *keygen) LoadSecretKey(secretKey string) error {
	decodedKey, err := base64.StdEncoding.DecodeString(secretKey)

	if err != nil {
		return ErrInvalidSecretKey
	}

	k.privateKey = decodedKey
	return nil
}

func (k *keygen) Sign(message []byte) ([]byte, error) {
	if k.privateKey == nil {
		return nil, ErrPrivateKeyNotLoaded
	}
	signer, err := initOqsSigner("Falcon-"+getKeySecurityLevel(k.alg), k.privateKey)
	if err != nil {
		return nil, err
	}
	return signer.Sign(message)
}

func (k *keygen) GetPrivate() []byte {
	return k.privateKey
}

func NewKeyManager(algorithm pb.Algorithm, secretKey string) (KeyManagement, error) {
	k := &keygen{}
	err := k.Init(algorithm, secretKey)
	if err != nil {
		return nil, err
	}
	return k, nil
}

func (k *keygen) generateKeyPair() ([]byte, []byte, error) {
	switch k.alg {
	case pb.Algorithm_FALCON512:
		signer, err := initOqsSigner("Falcon-512", nil)
		if err != nil {
			return nil, nil, err
		}
		pk, err := signer.GenerateKeyPair()
		sk := signer.ExportSecretKey()

		return pk, sk, nil

	case pb.Algorithm_KYBER512, pb.Algorithm_KYBER768, pb.Algorithm_KYBER1024:
		kem, err := initOqsKEM("Kyber-"+getKeySecurityLevel(k.alg), nil)
		if err != nil {
			return nil, nil, err
		}
		pk, err := kem.GenerateKeyPair()
		sk := kem.ExportSecretKey()
		if err != nil {
			return nil, nil, err
		}
		return pk, sk, nil

	default:
		return nil, nil, ErrUnsupportedAlgorithm
	}
}
func initOqsSigner(keySecurityLevel string, secretKey []byte) (oqs.Signature, error) {
	signer := oqs.Signature{}
	err := signer.Init(keySecurityLevel, secretKey)
	return signer, err
}

func initOqsKEM(keySecurityLevel string, secretKey []byte) (oqs.KeyEncapsulation, error) {
	kem := oqs.KeyEncapsulation{}
	err := kem.Init(keySecurityLevel, secretKey)
	return kem, err
}

func getKeySecurityLevel(algorithm pb.Algorithm) string {
	switch algorithm {
	case pb.Algorithm_KYBER512:
		return "512"
	case pb.Algorithm_KYBER768:
		return "768"
	case pb.Algorithm_KYBER1024:
		return "1024"
	case pb.Algorithm_FALCON512:
		return "512"
	default:
		return ""
	}
}

var (
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
	ErrPrivateKeyNotLoaded  = errors.New("private key not loaded")
	ErrInvalidSecretKey     = errors.New("invalid secret key")
)

func (k *keygen) DeriveKey() string {
	if k.privateKey == nil {
		return ""
	}
	hasher := crypto.NewBlake3()
	return hasher.HashToBase64(k.privateKey)
}

func (k *keygen) GetAlgorithm() pb.Algorithm {
	return k.alg
}
