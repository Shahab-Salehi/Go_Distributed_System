package main

import (
	"encoding/json"
	"root/bitcoin"
	"root/lsp"
	"fmt"
	"log"
	"os"
	"strconv"
)

type server struct {
	lspServer lsp.Server
}
type myserver struct {
	Myminer    map[int]*minerst
	number     int
	SS         chan int
	LL         chan int
	MyminerMID map[int]int
}
type minerst struct {
	MID  int
	FF   chan int
	RecM chan *bitcoin.Message
}

func startServer(port int) (*server, error) {
	// TODO: implement this!

	NewS, err := lsp.NewServer(port, lsp.NewParams())
	if err != nil {
		fmt.Println("StartNewServer ERR")
	}
	Serv := &server{
		lspServer: NewS,
	}

	return Serv, err
}

var LOGF *log.Logger

func main() {
	s := &myserver{
		Myminer:    make(map[int]*minerst),
		number:     1,
		SS:         make(chan int, 1000),
		LL:         make(chan int, 1),
		MyminerMID: make(map[int]int),
	}
	s.LL <- 1

	// You may need a logger for debug purpose
	const (
		name = "log.txt"
		flag = os.O_RDWR | os.O_CREATE
		perm = os.FileMode(0666)
	)

	file, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return
	}
	defer file.Close()

	LOGF = log.New(file, "", log.Lshortfile|log.Lmicroseconds)
	// Usage: LOGF.Println() or LOGF.Printf()

	const numArgs = 2
	if len(os.Args) != numArgs {
		fmt.Printf("Usage: ./%s <port>", os.Args[0])
		return
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Port must be a number:", err)
		return
	}

	srv, err := startServer(port)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Server listening on port", port)

	defer srv.lspServer.Close()

	for {
		var MyMes bitcoin.Message

		CID, Data, err := srv.lspServer.Read()
		if err != nil {
			fmt.Println("ServerRead ERR")

		} else {
			err = json.Unmarshal(Data, &MyMes)
			if err != nil {
				fmt.Print("ServerUnmarshal ERR")
			}
			if MyMes.Type == bitcoin.Request {
				go func() {
					k := 0
					CCID := CID
					MMyMes := MyMes
					FID := 1
					for {

						for {
							<-s.LL
							for i := 1; i < s.number; i++ {
								if len(s.Myminer[i].FF) == 0 {
									k = 1
									FID = i
									s.Myminer[i].FF <- 1
									break

								}

							}
							s.LL <- 1
							if k == 1 {
								break
							}
							<-s.SS

						}

						MuMyMes, err := json.Marshal(MMyMes)
						if err != nil {
							fmt.Println("SJsoning ERR")

						}
						eeRR := srv.lspServer.Write(s.Myminer[FID].MID, MuMyMes)
						if eeRR == nil {
							break
						}
					}

					NewRecM := <-s.Myminer[FID].RecM
					MuNewRecM, err := json.Marshal(NewRecM)
					if err != nil {
						fmt.Println("SSJsoning ERR")
					}
					<-s.Myminer[FID].FF
					srv.lspServer.Write(CCID, MuNewRecM)

				}()

			} else if MyMes.Type == bitcoin.Join {
				NewMiner := &minerst{
					MID:  CID,
					FF:   make(chan int, 1),
					RecM: make(chan *bitcoin.Message, 100),
				}
				s.Myminer[s.number] = NewMiner
				s.MyminerMID[CID] = s.number
				s.number++

			} else if MyMes.Type == bitcoin.Result {
				s.Myminer[s.MyminerMID[CID]].RecM <- &MyMes
				s.SS <- 1

			}

		}

	}
}
