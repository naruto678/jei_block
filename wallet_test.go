package main

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcutil/base58"
)

func TestBase58(t *testing.T) {

	fmt.Println(base58.Encode([]byte("This is base58 encoded")))
}

func TestNewWallet(t *testing.T) {
	wallet := NewWallet()
	fmt.Println(string(wallet.GetAddress()))
}
