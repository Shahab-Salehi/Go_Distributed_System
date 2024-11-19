// Contains the implementation of a LSP server.

package lsp

import "errors"
import "root/lspnet"
import "encoding/json"
import "strconv"
import "fmt"
import "time"

type server struct {
	connection *lspnet.UDPConn
	clossee    chan int
	MyClients  map[int]*myClient
	NewID      int
	Pa         *Params
	outgoing   chan *Smes
}
type myClient struct {
	addr        *lspnet.UDPAddr
	incoming    chan Message
	id          int
	sqN         int
	incomingAck chan *Message
	DataAck     map[int]int
	CanRead     map[int]int
	RecivedData map[int]int
	closee      chan int
}

type Smes struct {
	Payload []byte
	ConnID  int
}

// NewServer creates, initiates, and returns a new server. This function should
// NOT block. Instead, it should spawn one or more goroutines (to handle things
// like accepting incoming client connections, triggering epoch events at
// fixed intervals, synchronizing events using a for-select loop like you saw in
// project 0, etc.) and immediately return. It should return a non-nil error if
// there was an error resolving or listening on the specified port number.
func NewServer(port int, params *Params) (Server, error) {
	time.Sleep(300 * time.Millisecond)
	StPort := strconv.FormatInt(int64(port), 10)
	ADre, err := lspnet.ResolveUDPAddr("udp", ":"+StPort)
	if err != nil {
		return nil, err
	}
	Conn, ER := lspnet.ListenUDP("udp", ADre)

	s := &server{
		connection: Conn,
		clossee:    make(chan int),
		MyClients:  make(map[int]*myClient),
		NewID:      1,
		Pa:         params,
		outgoing:   make(chan *Smes, 800),

	}

	go func() {

		var mes = Message{}
		var buffer []byte = make([]byte, 2000)
		for {

			select {
			case <-s.clossee:
				return
			default:
				len, addr, ERROR := s.connection.ReadFromUDP(buffer)

				if ERROR != nil {
					return
				}
				json.Unmarshal(buffer[0:len], &mes)

				if mes.Type == MsgData {

					AckM := &Message{
						Type:   MsgAck,
						ConnID: mes.ConnID,
						SeqNum: mes.SeqNum,
					}

					MeToSe, _ := json.Marshal(AckM)
					_, EERROO := s.connection.WriteToUDP(MeToSe, addr)
					if EERROO != nil {
						fmt.Println("MsgType errDA")

					}
					if _, ok := s.MyClients[mes.ConnID].RecivedData[mes.SeqNum]; ok {

					} else {

						s.MyClients[mes.ConnID].RecivedData[mes.SeqNum] = 1
						s.MyClients[mes.ConnID].incoming <- mes
					}

				} else if mes.Type == MsgAck {

					s.MyClients[mes.ConnID].incomingAck <- &mes
					s.MyClients[mes.ConnID].DataAck[mes.SeqNum] = 1

				} else if mes.Type == MsgConnect {
					AckM := &Message{
						Type:   MsgAck,
						ConnID: s.NewID,
						SeqNum: 0,
					}

					NewClient := &myClient{
						addr:        addr,
						incoming:    make(chan Message),
						id:          s.NewID,
						sqN:         1,
						incomingAck: make(chan *Message, 600),
						DataAck:     make(map[int]int),

						CanRead:     make(map[int]int),
						RecivedData: make(map[int]int),
						closee:      make(chan int),
					}

					NewClient.CanRead[0] = 1
					NewClient.DataAck[0] = 1
					s.MyClients[s.NewID] = NewClient
					s.NewLis(NewClient)
					s.NewID++

					MeToSe, _ := json.Marshal(AckM)
					_, EERROO := s.connection.WriteToUDP(MeToSe, addr)

					if EERROO != nil {
					}
				}

			}

		}

	}()

	return s, ER
}

func (s *server) NewLis(Cli *myClient) {

	go func() {
		for {
			select {
			case NeMes := <-Cli.incoming:

				go func() {
					SSEQQ := NeMes.SeqNum
					DData := NeMes.Payload
					CCID := NeMes.ConnID
					for {
						if Cli.CanRead[SSEQQ-1] == 1 {

							Mes := Smes{
								Payload: DData,
								ConnID:  CCID,
							}
							s.outgoing <- &Mes
							Cli.CanRead[SSEQQ] = 1
							break
						} else {
							time.Sleep(200 * time.Millisecond)

						}

					}
				}()
			case <-Cli.closee:
				break

			}
		}

	}()

}

func (s *server) Read() (int, []byte, error) {

	select {
	case Mes := <-s.outgoing:
		if _, ok := s.MyClients[Mes.ConnID]; ok {

			return Mes.ConnID, Mes.Payload, nil
		} else {

			return Mes.ConnID, nil, errors.New("READ ID ERROR")

		}

	} // Blocks indefinitely.

}

func (s *server) Write(connID int, payload []byte) error {

	if _, ok := s.MyClients[connID]; ok {

		MyData := &Message{

			Type:    MsgData,
			ConnID:  connID,
			SeqNum:  s.MyClients[connID].sqN,
			Size:    len(payload),
			Payload: payload,
		}
		s.MyClients[connID].sqN++
		go func() {
			for {

				if s.MyClients[connID].DataAck[MyData.SeqNum-s.Pa.WindowSize] == 1 || MyData.SeqNum-s.Pa.WindowSize < 0 {
					MeToSe, _ := json.Marshal(MyData)
					_, EERROO := s.connection.WriteToUDP(MeToSe, s.MyClients[connID].addr)
					if EERROO != nil {
						fmt.Println("MsgType errDS")
					} else {
						s.MyClients[connID].DataAck[MyData.SeqNum] = 0
					}

					for i := 0; i < s.Pa.EpochLimit; i++ {

						if s.MyClients[connID].DataAck[MyData.SeqNum] == 1 {
							break
						}

						time.Sleep(time.Millisecond * time.Duration(s.Pa.EpochMillis))
						s.connection.WriteToUDP(MeToSe, s.MyClients[connID].addr)

					}

					break

				}
				time.Sleep(200 * time.Millisecond)

			}
		}()
		return nil
	} else {
		fmt.Println("WRITE ID ERROR")
		return errors.New("ID ERROR")
	}

}

func (s *server) CloseConn(connID int) error {
	fmt.Println("CloseConn")
	s.MyClients[connID].closee <- 1
	time.Sleep(200 * time.Millisecond)
	delete(s.MyClients, connID)
	return nil
}

func (s *server) Close() error {
	fmt.Println("Close")
	return nil
}
