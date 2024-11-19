package main

import (
	"encoding/json"
	"root/bitcoin"
	"root/lsp"
	"fmt"
	"os"
)

// Attempt to connect miner as a client to the server.
func joinWithServer(hostport string) (lsp.Client, error) {

	miner, err := lsp.NewClient(hostport, lsp.NewParams())
	if err != nil {
		fmt.Println("NewClient ERR")
	}
	Mes := bitcoin.NewJoin()
	MaMes, err := json.Marshal(Mes)
	if err != nil {
		fmt.Println("Marshal ERR")
	}
	miner.Write(MaMes)

	return miner, err
}

func main() {
	const numArgs = 2
	if len(os.Args) != numArgs {
		fmt.Printf("Usage: ./%s <hostport>", os.Args[0])
		return
	}

	hostport := os.Args[1]
	miner, err := joinWithServer(hostport)
	if err != nil {
		fmt.Println("Failed to join with server:", err)
		return
	}

	defer miner.Close()

	var MyMes bitcoin.Message
	for {
		ResMe, err := miner.Read()
		if err != nil {
			fmt.Println("Read ERR")
			miner.Close()
		}
		if ResMe != nil {
			err = json.Unmarshal(ResMe, &MyMes)
			if err != nil {
				fmt.Print("unmarshal ERR")
			}

			LoHash := bitcoin.Hash(MyMes.Data, MyMes.Lower)
			var i uint64
			var Nonce uint64
			Nonce = MyMes.Lower
			for i = MyMes.Lower; i <= MyMes.Upper; i++ {
				Temp := bitcoin.Hash(MyMes.Data, i)
				if Temp < LoHash {
					LoHash = Temp
					Nonce = i
				}
			}
			ReMes := bitcoin.NewResult(LoHash, Nonce)
			MarReMes, err := json.Marshal(ReMes)
			if err != nil {
				fmt.Println("Marshal ERR")
			}
			miner.Write(MarReMes)

		} else {
			miner.Close()

		}
	}
}
