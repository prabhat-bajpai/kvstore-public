package cluster

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
	"io/ioutil"
	"strings"
)

const (
	BROADCAST = -1
)
const (
	BUFFER_LENGTH_IN = 100
)
const (
	BUFFER_LENGTH_OUT = 100
)
const (
	MAX_SERVER = 5
)

type jsonobject struct {
	Object ObjectType
}

type ObjectType struct {
	Buffer_size int
	Peers       []Peer1
}

type Peer1 struct {
	Id      int
	Host    string
	Timeout int
	Logdir  string
}

type Envelope struct {
	// On the sender side, Pid identifies the receiving peer. If instead, Pid is
	// set to cluster.BROADCAST, the message is sent to all peers. On the receiver side, the
	// Id is always set to the original sender. If the Id is not found, the message is silently dropped
	Pid int

	// An id that globally and uniquely identifies the message, meant for duplicate detection at
	// higher levels. It is opaque to this package.
	//MsgId int64

	// the actual message.
	Msg interface{}
}
type Cluster struct {
	myid             int
	peer             [MAX_SERVER]int
	no_of_peer       int
	logdir           string
	election_timeout int
	my_cluster       *zmq.Socket
	cluster          [MAX_SERVER]*zmq.Socket
	inbox            chan *Envelope
	outbox           chan *Envelope
	network          bytes.Buffer
}

func (c Cluster) Timeout() int {
	return c.election_timeout
}

func (c Cluster) Pid() int {
	return c.myid
}

func (c Cluster) Peer() [MAX_SERVER]int {
	return c.peer
}

func (c Cluster) Inbox() chan *Envelope {
	return c.inbox
}

func (c Cluster) Outbox() chan *Envelope {
	return c.outbox
}

type Server interface {
	// Id of this server
	Pid() int

	// array of other servers' ids in the same cluster
	Peers() []int

	// the channel to use to send messages to other peers
	// Note that there are no guarantees of message delivery, and messages
	// are silently dropped
	Outbox() chan *Envelope

	// the channel to receive messages from other peers.
	Inbox() chan *Envelope
}

func New(id int, f string) Cluster {

	var myid int
	var peer [MAX_SERVER]int
	var cluster [MAX_SERVER]*zmq.Socket
	var mycluster *zmq.Socket
	var no_of_p int
	var server Cluster
	file, _ := ioutil.ReadFile(f)
	var jsontype jsonobject
	var logfile string
	var timeout int
	var network_cd bytes.Buffer

	json.Unmarshal(file, &jsontype)
	myid = id
	no_of_p = jsontype.Object.Buffer_size

	for i := 0; i < jsontype.Object.Buffer_size; i++ {
		if jsontype.Object.Peers[i].Id != myid {
			peer[i] = jsontype.Object.Peers[i].Id
			cluster[i], _ = zmq.NewSocket(zmq.PUSH)
			cluster[i].Connect("tcp://" + jsontype.Object.Peers[i].Host)
		} else {
			mycluster, _ = zmq.NewSocket(zmq.PULL)
			mycluster.SetIdentity(string(id))
			a := strings.Split(jsontype.Object.Peers[i].Host, ":")
			mycluster.Bind("tcp://*:" + a[1])
			logfile = jsontype.Object.Peers[i].Logdir
			timeout = jsontype.Object.Peers[i].Timeout

		}
	}

	server = Cluster{myid: id, peer: peer, no_of_peer: no_of_p, logdir: logfile, election_timeout: timeout, my_cluster: mycluster, cluster: cluster, inbox: make(chan *Envelope, BUFFER_LENGTH_IN), outbox: make(chan *Envelope, BUFFER_LENGTH_OUT), network: network_cd}

	go server.Send()
	go server.Receive()

	return server
}

func (s Cluster) Send() {
	for {
		data := <-s.Outbox()
		msg := data.Msg.(int)*16 + s.myid
		//encoding
		write := new(bytes.Buffer)
		encoder := gob.NewEncoder(write)
		encoder.Encode(msg)

		if data.Pid == BROADCAST {
			for i := 0; i < s.no_of_peer; i++ {
				if s.cluster[i] != nil {
					_, err := s.cluster[i].SendBytes(write.Bytes(), 0)
					if err != nil {
						//panic(err)
					}
				}
			}
		} else {
			for i := 0; i < s.no_of_peer; i++ {
				if data.Pid == s.Peer()[i] {
					if s.cluster[i] != nil {
						_, err := s.cluster[i].SendBytes(write.Bytes(), 0)
						if err != nil {
							//panic(err)
						}
					}
				}
			}
		}
	}
}

func (s Cluster) Receive() {
	for {
		msg, err := s.my_cluster.RecvBytes(0)
		if err != nil {
			//panic(err)
		}
		var msg1 int
		//Decoding
		read := bytes.NewBuffer(msg)
		decoder := gob.NewDecoder(read)
		decoder.Decode(&msg1)

		id1 := msg1 % 16
		msg1 = msg1 / 16
		s.Inbox() <- &Envelope{Pid: id1, Msg: msg1}
	}
}
