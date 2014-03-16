package main

import (
	"fmt"
	cluster "github.com/prabhat-bajpai/kvstore-public/tree/master/raft/cluster"
	zmq "github.com/pebbe/zmq4"
	"time"
	//"math"
	"math/rand"
	"os"
	"strconv"
)

// This is the structure of Raft Object
type Replicator struct {
	server cluster.Cluster // Cluster Object
	term   int
	leader bool
}

// This is methods supported by Raft object
type Raft interface {
	Get_Term() int
	Get_Isleader() bool
	Outbox() chan *cluster.Envelope
	Inbox() chan *cluster.Envelope
}

// Methord Return the current term of the Raft
func (r Replicator) Get_Term() int {
	return r.term
}

// Methord Return true if is leader
func (r Replicator) Get_Isleader() bool {
	return r.leader
}

// Return Election time of this raft server
func (r Replicator) Timeout() int {
	return r.server.Timeout()
}

// Raft layer Inbox
func (r Replicator) Inbox() chan *cluster.Envelope {
	return r.server.Inbox()
}

// Raft layer Outnbox
func (r Replicator) Outbox() chan *cluster.Envelope {
	return r.server.Outbox()
}

// This Function initialize the parameters of Raft Object
func New(id int, f string) Replicator {
	server := cluster.New(id, f)
	raft := Replicator{server: server, term: 0, leader: false}

	return raft
}

// This methord is for communication of raft server to testing program using Response sockets and this server is close if
// receive msg is 0 if msg is 1 then return phase to testing program
func raft_test(port string, server *Replicator, done chan bool, phase *int) {
	var reply string
	responder, _ := zmq.NewSocket(zmq.REP)
	responder.Bind("tcp://*:" + port)
	for {

		msg, _ := responder.Recv(0)
		if msg == "0" {
			done <- true
			break
		} else {
			reply = fmt.Sprintf("%d", *phase)
		}
		responder.Send(reply, 0)
	}
}

// This methord take care of receive msg from other raft servers
// Least significant 2 bits of msg tell the type of msg if envelope.Msg % 4 is 0 :- vote request msg , 1:- Heartbeat msg from
// leader , 2 :- grant msg from follower
// other bits tell term of the sender
func receive_m(server *Replicator, heartbeat_interm chan int, candidate_interm chan int, done chan bool, lead chan bool, phase *int, count *int) {
	for {
		envelope := <-server.Inbox() //If any massage in inbox

		if envelope.Msg.(int)%4 == 1 {
			if *phase == 2 {
				term := envelope.Msg.(int) / 4
				if term > server.Get_Term() {
					heartbeat_interm <- term
				}
			} else {
				term := envelope.Msg.(int) / 4
				heartbeat_interm <- term
			}
		} else if envelope.Msg.(int)%4 == 0 {
			if *phase == 0 {
				term := envelope.Msg.(int) / 4
				if term > server.Get_Term() {
					candidate_interm <- term
					server.term = term
					server.Outbox() <- &cluster.Envelope{Pid: envelope.Pid, Msg: server.Get_Term()*4 + 2}
					fmt.Println("Term is = " + fmt.Sprintf("%d", term))
				}
			} else {
				continue
			}
		} else {
			if *phase == 1 {
				term := envelope.Msg.(int) / 4
				if term == server.Get_Term() {
					*count = *count + 1
					if *count >= 3 {
						lead <- true
					}
				}
				continue
			}
		}
		<-done
	}
}

// This function for send msg but most important this function implement leader election algo using receive_m function
func send_m(server *Replicator, heartbeat_interm chan int, candidate_interm chan int, done chan bool, lead chan bool, phase *int, count *int) {
	for {
		if *phase == 2 {
			timer1 := time.NewTimer(time.Millisecond * 200)
			select {
			case term := <-heartbeat_interm:
				server.leader = false
				server.term = term
				*phase = 0
				done <- true
			case <-timer1.C:
				server.Outbox() <- &cluster.Envelope{Pid: -1, Msg: server.Get_Term()*4 + 1}
			}
		} else if *phase == 0 {
			// Election timeout timer
			timer1 := time.NewTimer(time.Millisecond * time.Duration(server.Timeout()))
			select {
			case term := <-heartbeat_interm:
				server.term = term
				done <- true
			case term := <-candidate_interm:
				server.term = term
				done <- true
			case <-timer1.C:
				*phase = 1
			}
		} else {
			timer1 := time.NewTimer(time.Millisecond * time.Duration(server.Timeout()))
			server.term = server.term + 1
			*count = 1
			server.Outbox() <- &cluster.Envelope{Pid: -1, Msg: server.Get_Term() * 4}
			select {
			case term := <-heartbeat_interm:
				server.term = term
				*phase = 0
				done <- true
			case <-lead:
				server.leader = true
				fmt.Print("leader")
				*phase = 2
				done <- true
			case <-timer1.C:
				*phase = 0
				// Wait for some random time after candidate time out
				//timer2 := time.NewTimer(time.Millisecond * time.Duration(server.Timeout()*rand.Intn(3)))
				time.Sleep(time.Millisecond * time.Duration(server.Timeout()*rand.Intn(3)))
				*phase = 1
			}
		}
	}
}

// This is main funtion which initialize different channel for commuication between receive_m and send_m function
func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	ex_done := make(chan bool, 1)

	// Id and port(For communication to test program) of Raft server
	id := os.Args[1]
	port := os.Args[2]

	myid, _ := strconv.Atoi(id)

	// Initialize server object
	server := New(myid /* config file */, "github.com/prabhat-bajpai/kvstore-public/tree/master/raft/peer.json")

	// This channel contain term of leader
	heartbeat_interm := make(chan int, 1)
	// This channel contain term of candidate voted by this raft server
	candidate_interm := make(chan int, 1)
	done := make(chan bool, 1)
	// If this raft server is become Leader
	lead := make(chan bool, 1)
	phase := 0 // 0 :- Follower, 1:- Candidate, 2:- Leader
	count := 0 // count no of vote grant msg

	go raft_test(port, &server, ex_done, &phase)
	go receive_m(&server, heartbeat_interm, candidate_interm, done, lead, &phase, &count)
	go send_m(&server, heartbeat_interm, candidate_interm, done, lead, &phase, &count)

	// Signal from testing program to close this server
	<-ex_done
}
