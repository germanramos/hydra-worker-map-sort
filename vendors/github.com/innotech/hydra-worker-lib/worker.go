package worker

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"

	// DEBUG
	"fmt"

	"github.com/innotech/hydra-worker-map-sort/vendors/github.com/innotech/hydra-worker-lib/vendors/github.com/BurntSushi/toml"
	zmq "github.com/innotech/hydra-worker-map-sort/vendors/github.com/innotech/hydra-worker-lib/vendors/github.com/alecthomas/gozmq"
)

const (
	SIGNAL_READY      = "\001"
	SIGNAL_REQUEST    = "\002"
	SIGNAL_REPLY      = "\003"
	SIGNAL_HEARTBEAT  = "\004"
	SIGNAL_DISCONNECT = "\005"

	DEFAULT_PRIORITY_LEVEL     = 0 // Local worker
	DEFAULT_VERBOSE            = false
	DEFAULT_HEARTBEAT_INTERVAL = 2600 * time.Millisecond
	DEFAULT_HEARTBEAT_LIVENESS = 3
	DEFAULT_RECONNECT_INTERVAL = 2500 * time.Millisecond
)

type Worker interface {
	close()
	recv([][]byte) [][]byte
	Run(func([]interface{}, map[string]interface{}) []interface{})
}

type lbWorker struct {
	HydraServerAddr string `toml:"hydra_server_address"` // Hydra Load Balancer address
	context         *zmq.Context
	PriorityLevel   int    `toml:"priority_level"`
	ServiceName     string `toml:"service_name"`
	Verbose         bool   `toml:"verbose"`
	socket          *zmq.Socket

	HeartbeatInterval time.Duration `toml:"heartbeat_interval"`
	heartbeatAt       time.Time
	Liveness          int `toml:"liveness"`
	livenessCounter   int
	ReconnectInterval time.Duration `toml:"reconnect_interval"`

	expectReply bool
	replyTo     []byte
}

func NewWorker(arguments []string) Worker {
	self := new(lbWorker)

	context, _ := zmq.NewContext()
	self.context = context
	self.HeartbeatInterval = DEFAULT_HEARTBEAT_INTERVAL
	self.PriorityLevel = DEFAULT_PRIORITY_LEVEL
	self.Liveness = DEFAULT_HEARTBEAT_LIVENESS
	self.ReconnectInterval = DEFAULT_RECONNECT_INTERVAL
	self.Verbose = DEFAULT_VERBOSE

	if err := self.Load(arguments); err != nil {
		panic(err.Error())
	}

	// Validate worker configuration
	if !self.isValid() {
		log.Printf("%#v", self)
		panic("You must set all required configuration options")
	}

	self.livenessCounter = self.Liveness
	self.reconnectToBroker()
	return self
}

// Load configures hydra-worker, it can be loaded from both
// custom file or command line arguments and the values extracted from
// files they can be overriden with the command line arguments.
func (self *lbWorker) Load(arguments []string) error {
	var path string
	f := flag.NewFlagSet("hydra-worker", flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)
	f.StringVar(&path, "config", "", "path to config file")
	f.Parse(arguments[1:])

	if path != "" {
		// Load from config file specified in arguments.
		if err := self.loadConfigFile(path); err != nil {
			return err
		}
	}

	// Load from command line flags.
	if err := self.loadFlags(arguments); err != nil {
		return err
	}

	return nil
}

// LoadFile loads configuration from a file.
func (self *lbWorker) loadConfigFile(path string) error {
	_, err := toml.DecodeFile(path, &self)
	return err
}

// LoadFlags loads configuration from command line flags.
func (self *lbWorker) loadFlags(arguments []string) error {
	var ignoredString string

	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)
	f.StringVar(&self.HydraServerAddr, "hydra-server-addr", self.HydraServerAddr, "")
	f.DurationVar(&self.HeartbeatInterval, "heartbeat-interval", self.HeartbeatInterval, "")
	f.IntVar(&self.Liveness, "Liveness", self.Liveness, "")
	f.IntVar(&self.PriorityLevel, "priority-level", self.PriorityLevel, "")
	f.DurationVar(&self.ReconnectInterval, "reconnect-interval", self.ReconnectInterval, "")
	f.StringVar(&self.ServiceName, "service-name", self.ServiceName, "")
	f.BoolVar(&self.Verbose, "v", self.Verbose, "")
	f.BoolVar(&self.Verbose, "Verbose", self.Verbose, "")

	// BEGIN IGNORED FLAGS
	f.StringVar(&ignoredString, "config", "", "")
	// END IGNORED FLAGS

	return nil
}

func (self *lbWorker) isValid() bool {
	if self.HydraServerAddr == "" {
		return false
	}
	if self.ServiceName == "" {
		return false
	}
	return true
}

// reconnectToBroker connects worker to hydra load balancer server (broker)
func (self *lbWorker) reconnectToBroker() {
	if self.socket != nil {
		self.socket.Close()
	}
	self.socket, _ = self.context.NewSocket(zmq.DEALER)
	// Pending messages shall be discarded immediately when the socket is closed with Close()
	// Set random identity to make tracing easier
	self.socket.SetLinger(0)
	self.socket.Connect(self.HydraServerAddr)
	if self.Verbose {
		log.Printf("Connecting to broker at %s...\n", self.HydraServerAddr)
	}
	self.sendToBroker(SIGNAL_READY, []byte(self.ServiceName), [][]byte{[]byte(strconv.Itoa(self.PriorityLevel))})
	self.heartbeatAt = time.Now().Add(self.HeartbeatInterval)
}

// sendToBroker dispatchs messages to hydra load balancer server (broker)
func (self *lbWorker) sendToBroker(command string, option []byte, msg [][]byte) {
	if len(option) > 0 {
		msg = append([][]byte{option}, msg...)
	}

	msg = append([][]byte{nil, []byte(command)}, msg...)
	if self.Verbose {
		log.Printf("Sending %X to broker\n", command)
	}
	self.socket.SendMultipart(msg, 0)
}

// close
func (self *lbWorker) close() {
	if self.socket != nil {
		self.socket.Close()
	}
	self.context.Close()
}

// recv receives messages from hydra load balancer server (broker) and send the responses back
func (self *lbWorker) recv(reply [][]byte) (msg [][]byte) {
	//  Format and send the reply if we were provided one
	if len(reply) == 0 && self.expectReply {
		panic("Error reply")
	}

	if len(reply) > 0 {
		if len(self.replyTo) == 0 {
			panic("Error replyTo")
		}
		reply = append([][]byte{self.replyTo, nil}, reply...)
		self.sendToBroker(SIGNAL_REPLY, nil, reply)
	}

	self.expectReply = true

	for {
		items := zmq.PollItems{
			zmq.PollItem{Socket: self.socket, Events: zmq.POLLIN},
		}

		_, err := zmq.Poll(items, self.HeartbeatInterval)
		if err != nil {
			panic(err) //  Interrupted
		}

		if item := items[0]; item.REvents&zmq.POLLIN != 0 {
			msg, _ = self.socket.RecvMultipart(0)
			if self.Verbose {
				log.Println("Received message from broker")
			}
			self.livenessCounter = self.Liveness
			if len(msg) < 2 {
				panic("Invalid msg") //  Interrupted
			}

			switch command := string(msg[1]); command {
			case SIGNAL_REQUEST:
				// log.Println("SIGNAL_REQUEST")
				//  We should pop and save as many addresses as there are
				//  up to a null part, but for now, just save one...
				self.replyTo = msg[2]
				msg = msg[4:6]
				return
			case SIGNAL_HEARTBEAT:
				// log.Println("SIGNAL_HEARTBEAT")
				// do nothin
			case SIGNAL_DISCONNECT:
				// log.Println("SIGNAL_DISCONNECT")
				self.reconnectToBroker()
			default:
				// TODO: catch error
				log.Println("Invalid input message")
			}
		} else if self.livenessCounter--; self.livenessCounter <= 0 {
			if self.Verbose {
				log.Println("Disconnected from broker - retrying...")
			}
			time.Sleep(self.ReconnectInterval)
			self.reconnectToBroker()
		}

		//  Send HEARTBEAT if it's time
		if self.heartbeatAt.Before(time.Now()) {
			self.sendToBroker(SIGNAL_HEARTBEAT, nil, nil)
			self.heartbeatAt = time.Now().Add(self.HeartbeatInterval)
		}
	}

	return
}

// Run executes the worker permanently
func (self *lbWorker) Run(fn func([]interface{}, map[string]interface{}) []interface{}) {
	for reply := [][]byte{}; ; {
		request := self.recv(reply)
		if len(request) == 0 {
			break
		}
		var instances []interface{}
		if err := json.Unmarshal(request[0], &instances); err != nil {
			log.Fatalln("Bad message: invalid instances")
			// TODO: Set REPLY and return
		}

		var args map[string]interface{}
		if err := json.Unmarshal(request[1], &args); err != nil {
			log.Fatalln("Bad message: invalid args")
			// TODO: Set REPLY and return
		}

		var processInstances func(levels []interface{}, ci *[]interface{}, iteration int) []interface{}
		processInstances = func(levels []interface{}, ci *[]interface{}, iteration int) []interface{} {
			levelIteration := 0
			for _, level := range levels {
				if level != nil {
					kind := reflect.TypeOf(level).Kind()
					if kind == reflect.Slice || kind == reflect.Array {
						o := make([]interface{}, 0)
						*ci = append(*ci, processInstances(level.([]interface{}), &o, levelIteration))
					} else {
						args["iteration"] = iteration
						t := fn(levels, args)
						return t
					}
					levelIteration = levelIteration + 1
				}
			}
			return *ci
		}
		var tmpInstances []interface{}
		computedInstances := processInstances(instances, &tmpInstances, 0)

		instancesResult, _ := json.Marshal(computedInstances)
		reply = [][]byte{instancesResult}
	}
}

// DEBUG: prints the message legibly
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
