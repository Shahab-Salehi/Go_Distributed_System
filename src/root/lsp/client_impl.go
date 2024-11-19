// Contains the implementation of a LSP client.

package lsp

//import "errors"
import "root/lspnet"
import "encoding/json"
import "fmt"

import "time"

type client struct {
	// TODO: implement this!
	connection  *lspnet.UDPConn
	id          int
	close       chan int
	incoming    chan *Message
	//Counter     int
	//lleenn      chan int
	incomingAck chan *Message
	outgoing    chan *Cmes
	Pa          *Params
	CanStart    chan int
	CanRead     map[int]int
	RecivedData map[int]int
	DataAck     map[int]int
	sqN         int
	start       int
}

type Cmes struct {
	Payload []byte
}

// NewClient creates, initiates, and returns a new client. This function
// should return after a connection with the server has been established
// (i.e., the client has received an Ack message from the server in response
// to its connection request), and should return a non-nil error if a
// connection could not be made (i.e., if after K epochs, the client still
// hasn't received an Ack message from the server in response to its K
// connection requests).
//
// hostport is a colon-separated string identifying the server's host address
// and port number (i.e., "localhost:9999").
func NewClient(hostport string, params *Params) (Client, error) {
	time.Sleep(300 * time.Millisecond)
	ADre, err := lspnet.ResolveUDPAddr("udp", hostport)
	if err != nil {
		return nil, err
	}
	Conn, ER := lspnet.DialUDP("udp", nil, ADre)
	if ER != nil {
		return nil, err
	}
	c := &client{
		connection:  Conn,
		id:          1,
		close:       make(chan int, 100),
		incoming:    make(chan *Message, 800),
		incomingAck: make(chan *Message, 800),
		outgoing:    make(chan *Cmes, 900),
		Pa:          params,
		CanStart:    make(chan int, 1),
		CanRead:     make(map[int]int),
		RecivedData: make(map[int]int),
		DataAck:     make(map[int]int),
		sqN:         1,
		start:       -1,
	}
	c.CanRead[0] = 1
	ConMes := &Message{
		Type:   MsgConnect,
		ConnID: 0,
		SeqNum: 0,
	}

	MeToSe, _ := json.Marshal(ConMes)
	_, EERROO := c.connection.Write(MeToSe)
	if EERROO != nil {
		fmt.Println("MsgConnect errCCM")
	}

	go func() {

		var mes = Message{}
		var buffer []byte = make([]byte, 2000)
		c.NewLis()

		for {

			select {
			case <-c.close:
				return
			default:
				len, _, ERROR := c.connection.ReadFromUDP(buffer)

				if ERROR != nil {
					return
				}
				EERR := json.Unmarshal(buffer[0:len], &mes)
				if EERR != nil {
					fmt.Println("MsgType C EERR")
				}

				if mes.Type == MsgData {

					AckM := &Message{
						Type:   MsgAck,
						ConnID: mes.ConnID,
						SeqNum: mes.SeqNum,
					}

					MeToSe, _ := json.Marshal(AckM)
					_, EERROO := c.connection.Write(MeToSe)
					if EERROO != nil {
						fmt.Println("MsgType C errDA")
					}

					if _, ok := c.RecivedData[mes.SeqNum]; ok {

					} else {

						c.RecivedData[mes.SeqNum] = 1
						c.incoming <- &mes
					}

				} else if mes.Type == MsgAck {
					if mes.SeqNum == 0 {
						c.id = mes.ConnID
						c.start = 1

						c.CanStart <- 1

					} else {

						c.incomingAck <- &mes
						c.DataAck[mes.SeqNum] = 1

					}

				}

			}
		}

	}()

	return c, ER

}

func (c *client) NewLis() {

	go func() {
		for {
			select {
			case NeMes := <-c.incoming:
				go func() {
					DD := NeMes.Payload
					SEQQ := NeMes.SeqNum

					for {
						if c.CanRead[SEQQ-1] == 1 {

							Mes := Cmes{
								Payload: DD,
							}

							c.outgoing <- &Mes
							c.CanRead[SEQQ] = 1
							break
						} else {
							time.Sleep(200 * time.Millisecond)

						}

					}
				}()

			}
		}

	}()

}

func (c *client) ConnID() int {
	return c.id
}

func (c *client) Read() ([]byte, error) {
	select {
	case Mes := <-c.outgoing:

		return Mes.Payload, nil

	}
}

func (c *client) Write(payload []byte) error {

	if c.start == -1 {
		<-c.CanStart
	}

	MyData := &Message{

		Type:    MsgData,
		ConnID:  c.id,
		SeqNum:  c.sqN,
		Size:    len(payload),
		Payload: payload,
	}
	c.sqN++
	go func() {
		for {
		//	fmt.Println(MyData.SeqNum-c.Pa.WindowSize)

			if c.DataAck[MyData.SeqNum-c.Pa.WindowSize] == 1 || MyData.SeqNum-c.Pa.WindowSize<= 0 {

				MeToSe, _ := json.Marshal(MyData)
				_, EERROO := c.connection.Write(MeToSe)
				if EERROO != nil {
					fmt.Println("MsgType errDS")
				} else {

					c.DataAck[MyData.SeqNum] = 0
				}
				t := c.Pa.EpochMillis
				for i := 0; i < c.Pa.EpochLimit; i++ {
					if c.DataAck[MyData.SeqNum] == 1 {

						break
					}
					time.Sleep(time.Duration(t) * time.Millisecond)
					_, EERROO := c.connection.Write(MeToSe)
					if EERROO != nil {
						fmt.Println("MsgType errDSGo")
					}

				}
				break

			}
			time.Sleep(200 * time.Millisecond)

		}
	}()

	return nil
}

func (c *client) Close() error {
	c.close <- 1

	return nil
}
