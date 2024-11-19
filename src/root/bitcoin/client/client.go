package main

import (
	"encoding/json"
	"root/bitcoin"
	"root/lsp"
	"fmt"
	"os"
	"strconv"
)

func main() {
	const numArgs = 4
	if len(os.Args) != numArgs {
		fmt.Printf("Usage: ./%s <hostport> <message> <maxNonce>", os.Args[0])
		return
	}
	hostport := os.Args[1]
	message := os.Args[2]
	maxNonce, err := strconv.ParseUint(os.Args[3], 10, 64)
	if err != nil {
		fmt.Printf("%s is not a number.\n", os.Args[3])
		return
	}

	client, err := lsp.NewClient(hostport, lsp.NewParams())
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}

	defer client.Close()

	NReq := bitcoin.NewRequest(message, 0, maxNonce)
	MNReq, err := json.Marshal(NReq)
	if err != nil {
		fmt.Println("Jsoning ERR")
	}

	client.Write(MNReq)

	RecMes, err := client.Read()
	if err != nil {
		fmt.Println(" Read ERR")
	}
	var MyMes bitcoin.Message
	err = json.Unmarshal(RecMes, &MyMes)
	if err != nil {
		fmt.Print("unmarshal ERR")
	}

	printResult(MyMes.Hash, MyMes.Nonce)
	client.Close()
	printDisconnected()

}

// printResult prints the final result to stdout.
func printResult(hash, nonce uint64) {
	fmt.Println("Result", hash, nonce)
}

// printDisconnected prints a disconnected message to stdout.
func printDisconnected() {
	fmt.Println("Disconnected")
}
