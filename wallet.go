package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := &Wallet{private, public}
	return wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, _ := ecdsa.GenerateKey(curve, rand.Reader)
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}

func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)
	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)
	fullpayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullpayload)
	return address
}

func Base58Encode(payload []byte) []byte {
	return []byte(base58.Encode(payload))
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:checkSumLen]

}

func HashPubKey(pubKey []byte) []byte {
	publicSha256 := sha256.Sum256(pubKey)
	ripemd160hasher := ripemd160.New()
	ripemd160hasher.Write(publicSha256[:])
	publicRipemd160 := ripemd160hasher.Sum(nil)
	return publicRipemd160

}
