package main

import "testing"

func TestCreation(t *testing.T) {
	// This test tests the creation of a blockchain with a specific address
	mining_wallet := NewWallet()
	address := mining_wallet.GetAddress()

	chain := CreateBlockchain(string(address))
	if chain == nil {
		t.Fatal("Chain was not created ")
	}

}
