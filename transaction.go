package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcutil/base58"
)

type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]

}

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}
type TxInput struct {
	Txid      []byte
	Vout      int
	PubKey    []byte
	Signature []byte
}

func NewCoinBaseTX(address, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to %s", address)
	}
	pubKeyHash := base58.Decode(address)
	txin := TxInput{
		Txid:      []byte{},
		Vout:      -1,
		PubKey:    pubKeyHash,
		Signature: []byte{},
	}
	txout := TxOutput{subsidy, pubKeyHash}
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
	tx.SetID()
	return &tx

}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	inHash := HashPubKey(in.PubKey)
	return bytes.Compare(inHash, pubKeyHash) == 0
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func (out *TxOutput) Lock(address []byte) {
	// the address is base58 encoded .
	pubKeyHash := base58.Decode(string(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (tx *Transaction) Coinbase() bool {
	return tx.ID == nil

}

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput
	inPubKeyHash := base58.Decode(from)
	outPubKeyHash := base58.Decode(to)
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			input := TxInput{
				Txid:   txID,
				Vout:   out,
				PubKey: inPubKeyHash,
			} //TxInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}
	outputs = append(outputs, TxOutput{amount, outPubKeyHash})
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, inPubKeyHash})
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}
