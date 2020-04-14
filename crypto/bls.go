// +build relic

package crypto

// #cgo CFLAGS: -g -Wall -std=c99 -I./ -I./relic/build/include
// #cgo LDFLAGS: -Lrelic/build/lib -l relic_s
// #include "bls_include.h"
import "C"

import (
	"errors"
	"fmt"
	"sync"

	"github.com/dapperlabs/flow-go/crypto/hash"
)

// blsBls12381Algo, embeds SignAlgo
type blsBls12381Algo struct {
	// points to Relic context of BLS12-381 with all the parameters
	context ctx
	// the signing algo
	algo SigningAlgorithm
}

//  Once variables to use a unique instance
var blsInstance *blsBls12381Algo
var once sync.Once

// returns a new BLS signer on curve BLS12-381
func newBlsBLS12381() *blsBls12381Algo {
	once.Do(func() {
		blsInstance = &(blsBls12381Algo{
			algo: BlsBls12381,
		})
		blsInstance.init()
	})
	return blsInstance
}

// Sign signs an array of bytes
// This function does not modify the private key, even temporarily
// If the hasher used is KMAC128, it is not modified by the function, even temporarily
func (sk *PrKeyBlsBls12381) Sign(data []byte, kmac hash.Hasher) (Signature, error) {
	if kmac == nil {
		return nil, errors.New("Sign requires a Hasher")
	}
	// hash the input to 128 bytes
	h := kmac.ComputeHash(data)
	return newBlsBLS12381().blsSign(&sk.scalar, h), nil
}

// BLS_KMACFunction is the customizer used for KMAC in BLS
const BLS_KMACFunction = "H2C"

// NewBLS_KMAC returns a new KMAC128 instance with the right parameters
// chosen for BLS signatures and verifications
// tag is the domain separation tag
func NewBLS_KMAC(tag string) hash.Hasher {
	// the error is ignored as the parameter lengths are in the correct range of kmac
	kmac, _ := hash.NewKMAC_128([]byte(tag), []byte("BLS_KMACFunction"), opSwUInputLenBlsBls12381)
	return kmac
}

// Verify verifies a signature of a byte array using the public key
// The function assumes the public key is in the valid G2 subgroup as it is
// either generated by the library or read through the DecodePublicKey function.
// This function does not modify the public key, even temporarily
// If the hasher used is KMAC128, it is not modified by the function, even temporarily
func (pk *PubKeyBlsBls12381) Verify(s Signature, data []byte, kmac hash.Hasher) (bool, error) {
	if kmac == nil {
		return false, errors.New("VerifyBytes requires a Hasher")
	}
	// hash the input to 128 bytes
	h := kmac.ComputeHash(data)

	return newBlsBLS12381().blsVerify(&pk.point, s, h), nil
}

// generatePrivateKey generates a private key for BLS on BLS12381 curve
// The minimum size of the input seed is 48 bytes (for a sceurity of 128 bits)
func (a *blsBls12381Algo) generatePrivateKey(seed []byte) (PrivateKey, error) {
	if len(seed) < KeyGenSeedMinLenBlsBls12381 {
		return nil, fmt.Errorf("seed should be at least %d bytes",
			KeyGenSeedMinLenBlsBls12381)
	}

	sk := &PrKeyBlsBls12381{
		// public key is not computed
		pk: nil,
	}

	// maps the seed to a private key
	mapKeyZr(&(sk.scalar), seed)
	return sk, nil
}

func (a *blsBls12381Algo) decodePrivateKey(privateKeyBytes []byte) (PrivateKey, error) {
	if len(privateKeyBytes) != prKeyLengthBlsBls12381 {
		return nil, fmt.Errorf("the input length has to be equal to %d", prKeyLengthBlsBls12381)
	}
	sk := &PrKeyBlsBls12381{
		pk: nil,
	}
	readScalar(&sk.scalar, privateKeyBytes)
	if sk.scalar.checkMembershipZr() {
		return sk, nil
	}
	return nil, errors.New("the private key is not a valid BLS12-381 curve key")
}

func (a *blsBls12381Algo) decodePublicKey(publicKeyBytes []byte) (PublicKey, error) {
	if len(publicKeyBytes) != pubKeyLengthBlsBls12381 {
		return nil, fmt.Errorf("the input length has to be equal to %d", pubKeyLengthBlsBls12381)
	}
	var pk PubKeyBlsBls12381
	if readPointG2(&pk.point, publicKeyBytes) != nil {
		return nil, errors.New("the input slice does not encode a public key")
	}
	if pk.point.checkMembershipG2() {
		return &pk, nil
	}
	return nil, errors.New("the public key is not a valid BLS12-381 curve key")

}

// PrKeyBlsBls12381 is the private key of BLS using BLS12_381, it implements PrivateKey
type PrKeyBlsBls12381 struct {
	// public key
	pk *PubKeyBlsBls12381
	// private key data
	scalar scalar
}

func (sk *PrKeyBlsBls12381) Algorithm() SigningAlgorithm {
	return BlsBls12381
}

func (sk *PrKeyBlsBls12381) KeySize() int {
	return PrKeyLenBlsBls12381
}

// computePublicKey generates the public key corresponding to
// the input private key. The function makes sure the piblic key
// is valid in G2
func (sk *PrKeyBlsBls12381) computePublicKey() {
	var newPk PubKeyBlsBls12381
	// compute public key pk = g2^sk
	_G2scalarGenMult(&(newPk.point), &(sk.scalar))
	sk.pk = &newPk
}

func (sk *PrKeyBlsBls12381) PublicKey() PublicKey {
	if sk.pk != nil {
		return sk.pk
	}
	sk.computePublicKey()
	return sk.pk
}

func (a *PrKeyBlsBls12381) Encode() ([]byte, error) {
	dest := make([]byte, prKeyLengthBlsBls12381)
	writeScalar(dest, &a.scalar)
	return dest, nil
}

func (sk *PrKeyBlsBls12381) Equals(other PrivateKey) bool {
	otherBLS, ok := other.(*PrKeyBlsBls12381)
	if !ok {
		return false
	}
	return sk.scalar.equals(&otherBLS.scalar)
}

// PubKeyBlsBls12381 is the public key of BLS using BLS12_381,
// it implements PublicKey
type PubKeyBlsBls12381 struct {
	// public key data
	point pointG2
}

func (pk *PubKeyBlsBls12381) Algorithm() SigningAlgorithm {
	return BlsBls12381
}

func (pk *PubKeyBlsBls12381) KeySize() int {
	return PubKeyLenBlsBls12381
}

func (a *PubKeyBlsBls12381) Encode() ([]byte, error) {
	dest := make([]byte, pubKeyLengthBlsBls12381)
	writePointG2(dest, &a.point)
	return dest, nil
}

func (pk *PubKeyBlsBls12381) Equals(other PublicKey) bool {
	otherBLS, ok := other.(*PubKeyBlsBls12381)
	if !ok {
		return false
	}
	return pk.point.equals(&otherBLS.point)
}

// Get Macro definitions from the C layer as Cgo does not export macros
var signatureLengthBlsBls12381 = int(C.getSignatureLengthBLS_BLS12381())
var pubKeyLengthBlsBls12381 = int(C.getPubKeyLengthBLS_BLS12381())
var prKeyLengthBlsBls12381 = int(C.getPrKeyLengthBLS_BLS12381())

// init sets the context of BLS12381 curve
func (a *blsBls12381Algo) init() error {
	// Inits relic context and sets the B12_381 parameters
	if err := a.context.initContext(); err != nil {
		return err
	}
	a.context.precCtx = C.init_precomputed_data_BLS12_381()

	// compare the Go and C layer constants as a sanity check
	if signatureLengthBlsBls12381 != SignatureLenBlsBls12381 ||
		pubKeyLengthBlsBls12381 != PubKeyLenBlsBls12381 ||
		prKeyLengthBlsBls12381 != PrKeyLenBlsBls12381 {
		return errors.New("BLS on BLS-12381 settings are not correct")
	}
	return nil
}

// reInit the context of BLS12381 curve assuming there was a previous call to init()
// If the implementation evolves and relic has multiple contexts,
// reinit should be called at every a. operation.
func (a *blsBls12381Algo) reInit() {
	a.context.reInitContext()
}

// TEST/DEBUG/BENCH
// wraps a call to optimized SwU algorithm since cgo can't be used
// in go test files
func OpSwUUnitTest(output []byte, input []byte) {
	C.opswu_test((*C.uchar)(&output[0]),
		(*C.uchar)(&input[0]),
		SignatureLenBlsBls12381)
}

// computes a bls signature
func (a *blsBls12381Algo) blsSign(sk *scalar, data []byte) Signature {
	s := make([]byte, SignatureLenBlsBls12381)

	C._blsSign((*C.uchar)(&s[0]),
		(*C.bn_st)(sk),
		(*C.uchar)(&data[0]),
		(C.int)(len(data)))
	return s
}

// Checks the validity of a bls signature
func (a *blsBls12381Algo) blsVerify(pk *pointG2, s Signature, data []byte) bool {
	if len(s) != signatureLengthBlsBls12381 {
		return false
	}
	verif := C._blsVerify((*C.ep2_st)(pk),
		(*C.uchar)(&s[0]),
		(*C.uchar)(&data[0]),
		(C.int)(len(data)))

	return (verif == valid)
}

// checkMembershipZr checks a scalar is less than the groups order (r)
func (sk *scalar) checkMembershipZr() bool {
	verif := C.checkMembership_Zr((*C.bn_st)(sk))
	return verif == valid
}

// membershipCheckG2 runs a membership check of BLS public keys on BLS12-381 curve.
// Returns true if the public key is on the correct subgroup of the curve
// and false otherwise
// It is necessary to run this test once for every public key before
// it is used to verify BLS signatures. The library calls this function whenever
// it imports a key through the function DecodePublicKey.
func (pk *pointG2) checkMembershipG2() bool {
	verif := C.checkMembership_G2((*C.ep2_st)(pk))
	return verif == valid
}
