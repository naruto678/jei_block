package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
)

const (
	targetBits          = 6
	maxNonce            = math.MaxInt
	dbFile              = "block.db"
	blocksBucket        = "blocks"
	subsidy             = 100
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
	scriptPubKey        = "ArnabWalletAddress"
)

type CLI struct {
	bc *Blockchain
}

func (cli *CLI) AddBlock(data string) {
	cli.bc.AddBlock(data)
	fmt.Println("Success!!")
}
func (cli *CLI) ValidateArgs() bool { return true }
func (cli *CLI) PrintChain() {
	cli.bc = NewBlockchain("")
	bci := cli.bc.Iterator()
	for {
		block := bci.Next()
		fmt.Printf("Prev.Hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Data: %s\n", block.Data)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) createBlockchain(address string) {
	cli.bc = CreateBlockchain(address)
	cli.bc.db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	cli.bc = NewBlockchain(address)
	defer cli.bc.db.Close()
	balance := 0
	UTXOs := cli.bc.FindUTXO(address)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of '%s': %d\n", address, balance)

}

func (cli *CLI) sendBalance(to, from string, amount int) {
	cli.bc = NewBlockchain(from)
	defer cli.bc.db.Close()
	tx := NewUTXOTransaction(from, to, amount, cli.bc)
	err := cli.bc.MineBlock([]*Transaction{tx})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Success!!")
}

func (cli *CLI) Run() {
	cli.ValidateArgs()
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printBlockCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send the genesis block reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address of the balance")
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendCmdTo := sendCmd.String("to", "", "address of the receiver")
	sendCmdFrom := sendCmd.String("from", "", "address of the sender")
	sendCmdAmount := sendCmd.Int("amount", 0, "amount of money to send")

	switch os.Args[1] {
	case "createblockchain":
		createBlockchainCmd.Parse(os.Args[2:])
	case "printchain":
		printBlockCmd.Parse(os.Args[2:])
	case "getbalance":
		getBalanceCmd.Parse(os.Args[2:])

	case "send":
		sendCmd.Parse(os.Args[2:])

	default:
		fmt.Println("Unknown command . Exiting")
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}
	if printBlockCmd.Parsed() {
		cli.PrintChain()
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}
	if sendCmd.Parsed() {
		if *sendCmdTo == "" || *sendCmdFrom == "" {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.sendBalance(*sendCmdTo, *sendCmdFrom, *sendCmdAmount)
	}

}

func main() {
	cli := CLI{}
	cli.Run()
}
