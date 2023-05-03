package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/btcsuite/btcutil/base58"
)

type Blockchain struct {
	db  *bolt.DB
	tip []byte
}

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (chain *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{
		currentHash: chain.tip,
		db:          chain.db,
	}
	return bci
}

func (iterator *BlockchainIterator) Next() *Block {
	var block *Block
	err := iterator.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(iterator.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		panic(err)
	}
	iterator.currentHash = block.PrevBlockHash
	return block
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{
		block:  block,
		target: target,
	}
	return pow
}

func (pow *ProofOfWork) PrepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.HashTransactions(),
		[]byte(strconv.FormatInt(pow.block.Timestamp, 16)),
		[]byte(strconv.FormatInt(targetBits, 16)),
		[]byte(strconv.FormatInt(int64(nonce), 16)),
	},
		[]byte{})
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	fmt.Printf("Mining the block containing \"%s\" \n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.PrepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")
	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.PrepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}
func (chain *Blockchain) AddBlock(data string) {

	var lastHash []byte

	err := chain.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		panic(err)
	}

	newBlock := NewBlock(nil, data, lastHash)
	pow := NewProofOfWork(newBlock)
	nonce, hash := pow.Run()
	newBlock.Hash = hash
	newBlock.Nonce = nonce

	chain.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		b.Put([]byte("l"), newBlock.Hash)
		chain.tip = newBlock.Hash
		b.Put(newBlock.Hash, newBlock.Serialize())
		return nil
	})

}

func (bc *Blockchain) MineBlock(transactions []*Transaction) error {
	if !dbExists() {
		fmt.Println("Blockchain does not eixst. Create a new blockchain ")
		os.Exit(1)
	}
	err := bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		if bucket == nil {
			log.Panic("bucket not found")
			os.Exit(1)
		}

		last_hash := bucket.Get([]byte("l"))
		new_block := NewBlock(transactions, "", last_hash)
		pow := NewProofOfWork(new_block)
		nonce, hash := pow.Run()
		new_block.Nonce = nonce
		new_block.Hash = hash
		bc.tip = new_block.Hash
		bucket.Put([]byte("l"), new_block.Hash)
		bucket.Put(new_block.Hash, new_block.Serialize())
		return nil
	})
	return err
}

func NewBlockchain(address string) *Blockchain {
	var tip []byte
	if !dbExists() {
		fmt.Println("Blockchain does not exist. create a blockchain")
		os.Exit(1)
	}
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))

		if bucket == nil {
			return err
		}

		tip = bucket.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &Blockchain{
		db:  db,
		tip: tip,
	}

}

func CreateBlockchain(address string) *Blockchain {
	var tip []byte
	if dbExists() {
		fmt.Println("Blockchain already exists. Exiting")
		os.Exit(1)
	}

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		// 1. create a new coin base transaction
		// 2. crate a new genesis block and put the coin base transaction

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}
		cbtx := NewCoinBaseTX(address, genesisCoinbaseData)
		genesisBlock := NewGenesisBlock(cbtx)

		b.Put(genesisBlock.Hash, genesisBlock.Serialize())
		b.Put([]byte("l"), genesisBlock.Hash)
		tip = genesisBlock.Hash
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{
		db:  db,
		tip: tip,
	}

}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction
	pubKeyHash := base58.Decode(address)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txId := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				// was the output spent
				if spentTXOs[txId] != nil {
					for _, spentOut := range spentTXOs[txId] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if !tx.Coinbase() {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxId := hex.EncodeToString(in.Txid)
						spentTXOs[inTxId] = append(spentTXOs[inTxId], in.Vout)
					}
				}
			}

		}
		if len(block.PrevBlockHash) == 0 {
			break
		}

	}
	return unspentTxs

}

func (bc *Blockchain) FindUTXO(address string) []TxOutput {
	var UTXOs []TxOutput
	pubKeyHash := base58.Decode(address)
	unspentTransactions := bc.FindUnspentTransactions(address)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func dbExists() bool {
	_, err := os.Stat(dbFile)
	return err == nil
}

func NewGenesisBlock(transaction *Transaction) *Block {
	start_time := time.Now()
	genesis_block := NewBlock([]*Transaction{transaction}, "Genesis Block", []byte{})
	pow := NewProofOfWork(genesis_block)
	nonce, hash := pow.Run()
	genesis_block.Nonce = nonce
	genesis_block.Hash = hash
	fmt.Printf("new block mined in %f seconds\n", time.Since(start_time).Seconds())
	return genesis_block
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := bc.FindUnspentTransactions(address)
	pubKeyHash := base58.Decode(address)
	accumulated := 0
Work:
	for _, tx := range unspentTxs {
		txId := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txId] = append(unspentOutputs[txId], outIdx)
			}
			if accumulated >= amount {
				break Work
			}
		}
	}
	return accumulated, unspentOutputs
}
