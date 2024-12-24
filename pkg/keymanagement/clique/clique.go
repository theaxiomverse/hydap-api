package clique

import (
	"fmt"
	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	"go.dedis.ch/kyber/v4/proof"
	"go.dedis.ch/kyber/v4/suites"
	"go.dedis.ch/kyber/v4/util/random"
)

// CliqueModule implements cryptographic proof functionalities using Kyber's Rep predicate

type CliqueModule struct {
	suite       suites.Suite
	initialized bool
	secret      kyber.Scalar
	publicKey   kyber.Point
	proof       []byte
}

// Name returns the name of the module
func (c *CliqueModule) Name() string {
	return "KeyManagement_Clique"
}

// Initialize sets up the CliqueModule with a cryptographic proof
func (c *CliqueModule) Initialize() error {
	if c.initialized {
		return fmt.Errorf("Clique module is already initialized")
	}

	// Crypto setup
	c.suite = edwards25519.NewBlakeSHA256Ed25519()
	basePoint := c.suite.Point().Base()

	// Create public/private key pair (X, x)
	c.secret = c.suite.Scalar().Pick(random.New())
	c.publicKey = c.suite.Point().Mul(c.secret, nil)

	// Generate a proof of knowledge of the secret key x
	predicate := proof.Rep("X", "x", "B")
	scalarValues := map[string]kyber.Scalar{"x": c.secret}
	pointValues := map[string]kyber.Point{"B": basePoint, "X": c.publicKey}

	prover := predicate.Prover(c.suite, scalarValues, pointValues, nil)
	var err error
	c.proof, err = proof.HashProve(c.suite, "TEST", prover)
	if err != nil {
		return fmt.Errorf("failed to generate cryptographic proof: %w", err)
	}

	c.initialized = true
	fmt.Println("Clique module initialized successfully with cryptographic proof")
	return nil
}

// Terminate clears the CliqueModule state
func (c *CliqueModule) Terminate() error {
	if !c.initialized {
		return fmt.Errorf("Clique module is not initialized")
	}

	c.secret = nil
	c.publicKey = nil
	c.proof = nil
	c.initialized = false
	fmt.Println("Clique module terminated successfully")
	return nil
}

// Signature returns a dummy signature (can be extended for actual signing)
func (c *CliqueModule) Signature() string {
	return "Signature_Clique"
}

// VerifyProof verifies the stored cryptographic proof
func (c *CliqueModule) VerifyProof() error {
	if !c.initialized {
		return fmt.Errorf("module not initialized")
	}

	// Recreate the verifier for the proof
	predicate := proof.Rep("X", "x", "B")
	basePoint := c.suite.Point().Base()
	pointValues := map[string]kyber.Point{"B": basePoint, "X": c.publicKey}

	verifier := predicate.Verifier(c.suite, pointValues)
	err := proof.HashVerify(c.suite, "TEST", verifier, c.proof)
	if err != nil {
		return fmt.Errorf("failed to verify proof: %w", err)
	}

	fmt.Println("Proof verified successfully")
	return nil
}

// GenerateKeyPair returns the generated key pair
func (c *CliqueModule) GenerateKeyPair() (interface{}, interface{}) {
	return c.publicKey, c.secret
}
