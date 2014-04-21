package worker

import (
	"encoding/json"
	zmq "github.com/innotech/hydra/vendors/github.com/alecthomas/gozmq"
	"log"
	"time"
	// DEBUG
	"fmt"
)

const (
	SIGNAL_READY      = "\001"
	SIGNAL_REQUEST    = "\002"
	SIGNAL_REPLY      = "\003"
	SIGNAL_HEARTBEAT  = "\004"
	SIGNAL_DISCONNECT = "\005"

	HEARTBEAT_LIVENESS = 3
)

type Worker interface {
	close()
	recv([][]byte) [][]byte
	Run(func([]map[string]interface{}, map[string]string) []interface{})
}

type lbWorker struct {
	broker  string // Hydra Load Balancer address
	context *zmq.Context
	service string
	verbose bool
	worker  *zmq.Socket

	heartbeat   time.Duration
	heartbeatAt time.Time
	liveness    int
	reconnect   time.Duration

	expectReply bool
	replyTo     []byte
}

func NewWorker(broker, service string, verbose bool) Worker {
	context, _ := zmq.NewContext()
	self := &lbWorker{
		broker:    broker,
		context:   context,
		service:   service,
		verbose:   verbose,
		heartbeat: 2500 * time.Millisecond,
		liveness:  0,
		reconnect: 2500 * time.Millisecond,
	}
	self.reconnectToBroker()
	return self
}

func (self *lbWorker) reconnectToBroker() {
	if self.worker != nil {
		self.worker.Close()
	}
	self.worker, _ = self.context.NewSocket(zmq.DEALER)
	// Pending messages shall be discarded immediately when the socket is closed with Close()
	self.worker.SetLinger(0)
	self.worker.Connect(self.broker)
	if self.verbose {
		log.Printf("Connecting to broker at %s...\n", self.broker)
	}
	self.sendToBroker(SIGNAL_READY, []byte(self.service), nil)
	self.liveness = HEARTBEAT_LIVENESS
	self.heartbeatAt = time.Now().Add(self.heartbeat)
}

func (self *lbWorker) sendToBroker(command string, option []byte, msg [][]byte) {
	if len(option) > 0 {
		msg = append([][]byte{option}, msg...)
	}

	msg = append([][]byte{nil, []byte(command)}, msg...)
	if self.verbose {
		log.Printf("Sending %X to broker\n", command)
		//Dump(msg)
	}
	self.worker.SendMultipart(msg, 0)
}

func (self *lbWorker) close() {
	if self.worker != nil {
		self.worker.Close()
	}
	self.context.Close()
}

func (self *lbWorker) recv(reply [][]byte) (msg [][]byte) {

	Dump(reply)
	//  Format and send the reply if we were provided one
	if len(reply) == 0 && self.expectReply {
		panic("Error reply")
	}
	log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! ******************************** !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

	if len(reply) > 0 {
		if len(self.replyTo) == 0 {
			panic("Error replyTo")
		}
		reply = append([][]byte{self.replyTo, nil}, reply...)
		log.Println("SENDING TO BROKER >>>>>>>>>>>>")
		Dump(reply)
		self.sendToBroker(SIGNAL_REPLY, nil, reply)
	}

	self.expectReply = true

	for {
		items := zmq.PollItems{
			zmq.PollItem{Socket: self.worker, Events: zmq.POLLIN},
		}

		_, err := zmq.Poll(items, self.heartbeat)
		if err != nil {
			panic(err) //  Interrupted
		}

		log.Printf("RECV %d", len(items))
		log.Printf("RECV items[0] %d", items[0])
		log.Printf("RECV POLLIN %d", items[0].REvents&zmq.POLLIN)

		if item := items[0]; item.REvents&zmq.POLLIN != 0 {

			log.Printf("RECV2 %d", len(items))

			msg, _ = self.worker.RecvMultipart(0)

			log.Printf("RECV3 %d", len(msg))

			if self.verbose {
				log.Println("Received message from broker: ")
				//Dump(msg)
			}
			self.liveness = HEARTBEAT_LIVENESS
			Dump(msg)
			if len(msg) < 2 {
				panic("Invalid msg") //  Interrupted
			}

			switch command := string(msg[1]); command {
			case SIGNAL_REQUEST:
				log.Printf("Signal REQUEST received")
				//  We should pop and save as many addresses as there are
				//  up to a null part, but for now, just save one...
				log.Println("PRE REQUEST FROM WORKER")
				Dump(msg)
				self.replyTo = msg[2]
				log.Println("REPLY TO: " + string(self.replyTo))
				msg = msg[4:6]
				log.Println("DUMP 2")
				Dump(msg)
				log.Println("END DUMP")
				return
			case SIGNAL_HEARTBEAT:
				log.Printf("Signal HEARBEAT received")
				// do nothin
			case SIGNAL_DISCONNECT:
				log.Printf("Signal DISCONNECT received")
				self.reconnectToBroker()
			default:
				log.Println("Invalid input message:")
				//Dump(msg)
			}
		} else if self.liveness--; self.liveness <= 0 {
			if self.verbose {
				log.Println("Disconnected from broker - retrying...")
			}
			time.Sleep(self.reconnect)
			self.reconnectToBroker()
		}

		//  Send HEARTBEAT if it's time
		if self.heartbeatAt.Before(time.Now()) {
			self.sendToBroker(SIGNAL_HEARTBEAT, nil, nil)
			self.heartbeatAt = time.Now().Add(self.heartbeat)
		}
	}

	return
}

func (self *lbWorker) Run(fn func([]map[string]interface{}, map[string]string) []interface{}) {
	log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! ******************************** !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	for reply := [][]byte{}; ; {
		log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! ******************************** !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		request := self.recv(reply)
		log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! ******************************** !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		log.Printf("WORKER LOOP: %#v", request)
		if len(request) == 0 {
			log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! BREAK !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			break
		}
		// You should code your logic here
		var instances []map[string]interface{}
		if err := json.Unmarshal(request[0], &instances); err != nil {
			log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! FATAL INSTANCES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			log.Fatalln("Bad message: invalid instances")
			// TODO: Set REPLY and return
		}

		var args map[string]string
		if err := json.Unmarshal(request[1], &args); err != nil {
			log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! FATAL INSTANCES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			log.Fatalln("Bad message: invalid args")
			// TODO: Set REPLY and return
		}

		log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! computedInstances !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		computedInstances := fn(instances, args)

		instancesResult, _ := json.Marshal(computedInstances)
		reply = [][]byte{instancesResult}
		log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! CONTINUE !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}

// DEBUG
func Dump(msg [][]byte) {
	for _, part := range msg {
		isText := true
		fmt.Printf("[%03d] ", len(part))
		for _, char := range part {
			if char < 32 || char > 127 {
				isText = false
				break
			}
		}
		if isText {
			fmt.Printf("%s\n", part)
		} else {
			fmt.Printf("%X\n", part)
		}
	}
}
