package vss

import (
	"encoding/base64"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"hydap/pkg/keymanagement"
	"hydap/pkg/keymanagement/pb"
	"math/big"

	"github.com/open-quantum-safe/liboqs-go/oqs"
	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	"go.dedis.ch/kyber/v4/share"
	"go.dedis.ch/kyber/v4/util/random"
)

const (
	PrimeModulus            = (1 << 127) - 1
	ScaleFactor             = 1e8 // Adjusted for precision
	ErrFailedSignatureCheck = "signature verification failed for share"
)

type VSS struct {
	suite         kyber.Group
	threshold     int
	kyberEnc      oqs.KeyEncapsulation
	falconSig     oqs.Signature
	keyManagement keymanagement.KeyManagement // Introduced KeyManagement dependency
	sigManagement keymanagement.KeyManagement
}

func NewVSS(threshold int, algorithm pb.Algorithm) (*VSS, error) {
	vss := &VSS{
		suite:         edwards25519.NewBlakeSHA256Ed25519(),
		threshold:     threshold,
		keyManagement: keymanagement.NewKeyManager(algorithm), // Inject KeyManagement instance
	}

	sigManager, err := keymanagement.NewKeyManager(algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize signature manager: %w", err)
	}
	vss.sigManagement = sigManager

	// Initialize Kyber Encapsulation
	if err := vss.kyberEnc.Init(getKyberAlgorithmName(algorithm), nil); err != nil {
		return nil, fmt.Errorf("failed to initialize Kyber: %w", err)
	}

	// Ensure keys are ready
	vss.keyManagement.Init(algorithm)

	// Load secret key if required
	if err := vss.keyManagement.LoadSecretKey(); err != nil {
		return nil, fmt.Errorf("failed to load secret key: %w", err)
	}

	// Initialize Falcon Signature
	vss.sigManagement.Init(algorithm)

	return vss, nil
}

func (vss *VSS) SplitSecret(coordinates []float64, threshold, numShares int) ([][][4]interface{}, error) {
	allShares := [][][4]interface{}{}
	for _, coord := range coordinates {
		coordScalar, err := convertCoordinateToScalar(coord)
		if err != nil {
			return nil, fmt.Errorf("coordinate conversion failed: %w", err)
		}

		// Generate Shamir's Secret Sharing polynomial
		shamirPolynomial := share.NewPriPoly(vss.suite, threshold, vss.suite.Scalar().SetInt64(coordScalar), random.New())
		shamirShares := shamirPolynomial.Shares(numShares)

		// Encrypt and sign shares
		coordShares := [][4]interface{}{}
		for _, shamirShare := range shamirShares {
			encryptedShare, err := vss.encryptAndSignShareWithKeyManagement(shamirShare)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt and sign share: %w", err)
			}
			coordShares = append(coordShares, encryptedShare)
		}

		allShares = append(allShares, coordShares)
	}

	return allShares, nil
}

func convertCoordinateToScalar(coord float64) (int64, error) {
	scaledValue := big.NewFloat(coord * ScaleFactor)
	scaledInt, _ := scaledValue.Int64() // Lossless conversion for reasonable precision
	if scaledInt < 0 {
		return 0, errors.New("negative values not supported in secret sharing")
	}
	return scaledInt, nil
}

func (vss *VSS) encryptAndSignShareWithKeyManagement(shamirShare *share.PriShare) ([4]interface{}, error) {
	// Retrieve encoded public key from keymanagement
	publicKeyBytes := vss.keyManagement.GetKey()

	// Decode public key into native type (if necessary)
	publicKey, err := base64Decode(publicKeyBytes) // Assume this function decodes Base64-encoded keys
	if err != nil {
		return [4]interface{}{}, fmt.Errorf("failed to decode public key: %w", err)
	}

	// Encrypt the share using Kyber
	enc := oqs.KeyEncapsulation{}
	var sk []byte
	sk, _ = proto.Marshal(vss.keyManagement.GetPrivate())
	err = enc.Init(getKyberAlgorithmName(pb.Algorithm_KYBER512), sk)
	ciphertext, sharedSecret, err := enc.EncapSecret(publicKey)
	if err != nil {
		return [4]interface{}{}, fmt.Errorf("failed to encapsulate secret: %w", err)
	}

	// Sign the ciphertext
	signedPubKey := vss.keyManagement.SignedPublicKey()
	if signedPubKey == "" {
		return [4]interface{}{}, fmt.Errorf("failed to sign ciphertext using key management")
	}

	return [4]interface{}{shamirShare.I, ciphertext, sharedSecret, signedPubKey}, nil
}

func (vss *VSS) ReconstructSecret(allEncryptedShares [][][4]interface{}, publicKeyBytes string) ([]float64, error) {
	if len(allEncryptedShares) < vss.threshold {
		return nil, errors.New("not enough shares provided for reconstruction")
	}

	// Decode public key
	publicKey, err := base64Decode(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	reconstructedCoords := []float64{}
	for _, coordShares := range allEncryptedShares {
		sharesForReconstruction := [][2]int64{}
		for _, share := range coordShares {
			index := share[0].(int64)
			ciphertext := share[1].([]byte)
			sharedSecret := share[2].([]byte)
			signature := share[3].([]byte)

			// Verify the signature
			if !vss.verifySignature(ciphertext, signature, publicKey) {
				return nil, errors.New(ErrFailedSignatureCheck)
			}

			shareInt := new(big.Int).SetBytes(sharedSecret).Int64()
			sharesForReconstruction = append(sharesForReconstruction, [2]int64{index, shareInt})
		}

		// Perform Lagrange interpolation and scale back
		coordInt, err := vss.reconstructFromShares(sharesForReconstruction)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct from shares: %w", err)
		}
		reconstructedCoords = append(reconstructedCoords, float64(coordInt)/ScaleFactor)
	}

	return reconstructedCoords, nil
}

func getKyberAlgorithmName(algorithm pb.Algorithm) string {
	switch algorithm {
	case pb.Algorithm_KYBER512:
		return "Kyber-512"
	case pb.Algorithm_KYBER768:
		return "Kyber-768"
	case pb.Algorithm_KYBER1024:
		return "Kyber-1024"
	default:
		return ""
	}
}

func getFalconAlgorithmName(algorithm pb.Algorithm) string {
	switch algorithm {
	case pb.Algorithm_FALCON512:
		return "Falcon-512"
	default:
		return ""
	}
}

func base64Decode(encoded string) ([]byte, error) {
	// Decode a base64-encoded string into raw bytes
	return base64.StdEncoding.DecodeString(encoded)
}

func (vss *VSS) verifySignature(ciphertext, signature, publicKey []byte) bool {
	sig := oqs.Signature{}
	err := sig.Init(getFalconAlgorithmName(vss.sigManagement.GetAlgorithm()), nil)
	if err != nil {
		return false
	}
	return sig.Verify(ciphertext, signature, publicKey)
}

func (vss *VSS) reconstructFromShares(shares [][2]int64) (int64, error) {
	if len(shares) < vss.threshold {
		return 0, errors.New("insufficient shares for reconstruction")
	}

	// Convert shares to Kyber format
	priShares := make([]*share.PriShare, len(shares))
	for i, s := range shares {
		priShares[i] = &share.PriShare{
			I: int(s[0]),
			V: vss.suite.Scalar().SetInt64(s[1]),
		}
	}

	// Reconstruct using Lagrange interpolation
	secret, err := share.RecoverSecret(vss.suite, priShares, vss.threshold, len(shares))
	if err != nil {
		return 0, fmt.Errorf("reconstruction failed: %w", err)
	}

	// Convert back to int64
	return secret.V.Int64(), nil
}
