package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreation(t *testing.T) {
	// This test tests the creation of a blockchain with a specific address
	mining_wallet := NewWallet()
	address := mining_wallet.GetAddress()

	chain := CreateBlockchain(string(address))
	if chain == nil {
		t.Fatal("Chain was not created ")
	}

	utxo_list := chain.FindUTXO(string(address))
	total_balance := 0
	for _, utxo := range utxo_list {
		total_balance += utxo.Value
	}
	// check if the miner got the  correct subsidy reward
	assert.Equal(t, total_balance, subsidy)
	fmt.Println("This is the total balance", total_balance)

	// do clean up here

	err := os.Remove("./" + dbFile)
	if err != nil {
		t.Fatal(err)
	}

}
