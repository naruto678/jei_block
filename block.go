package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Transactions  []*Transaction
	Nonce         int
}

func (block *Block) SetHash() {
	timestamp := strconv.FormatInt(block.Timestamp, 10)
	headers := bytes.Join([][]byte{
		[]byte(timestamp),
		block.Data,
		block.PrevBlockHash,
	}, []byte{})

	hash := sha256.Sum256(headers)
	block.Hash = hash[:]
}

func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}
func NewBlock(transactions []*Transaction, data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Transactions:  transactions,
	}
	block.SetHash()
	return block
}

func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func DeserializeBlock(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		panic(err)
	}
	return &block
}
