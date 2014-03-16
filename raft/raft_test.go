package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"os/exec"
	"testing"
	"time"
)

// This is simple send function here i communicate with raft using simple request-response socket and ask to raft their status
// like phase of raft and we can also close raft server using this function.
// Here msg :- 1 means msg for ask phase of server and msg:- 0 means close the raft server
func send(msg int, requester *zmq.Socket) string {
	// Decoding of phse receive by type 0 msg is 0 :- Follower, 1:- Candidate, 2:- Leader

	// By default phase :- "3" means no response from server and server is close
	phase := "3"
	requester.Send(fmt.Sprintf("%d", msg), 0)
	if msg == 1 {
		phase, _ = requester.Recv(0)
	}
	if phase == "0" {
		phase = "Follower"
	} else if phase == "1" {
		phase = "Candidate"
	} else if phase == "2" {
		phase = "Leader"
	}
	return phase
}

// This function is used for start a Raft server from external program here we send the id of raft server
// and port for the response socket so that this program can communicate with server
func execl(myid, port int) {
	lsCmd := exec.Command("go", "run", "raft.go", fmt.Sprintf("%d", myid), fmt.Sprintf("%d", port))

	_, err := lsCmd.Output()
	if err != nil {
		panic(err)
	}
}

// This function check the phase of server at different test cases in each test case i change the configuration of server
// and run this method
// This method ask phase of Raft 100 times in 50 millisecond interval
func run(msg []int, requester []*zmq.Socket, up []bool, phase []string, t_case int, t *testing.T) {
	// These variables for checking the status of whole custer
	var leader bool
	two_lead := false
	no_leader := 0
	no_leader_s := 0

	for k := 1; k < 100; k++ {
		for i := 0; i < 5; i++ {
			if up[i] == true {
				phase[i] = send(msg[i], requester[i])
			}
		}
		for i := 0; i < 5; i++ {
			leader = false
			if phase[i] == "Leader" {
				leader = true
				t := i + 1
				for j := t; j < 5; j++ {
					if phase[j] == "Leader" {
						two_lead = true
					}
					i = j
				}
			}

		}
		if leader == false {
			if t_case != 4 {
				if k == (no_leader_s + no_leader) {
					no_leader = no_leader + 1
				} else {
					no_leader = 1
					no_leader_s = k
				}
				// Now this time is 50*50 = 2500 ms app 4 time out
				if no_leader > 50 {
					fmt.Println("Nolader")
					t.Error("NO Leader")
				}
			}
		}
		if t_case == 4 {
			if leader == true {
				t.Error("Leader is elected in majority no of Failure")
			}
		}

		if two_lead == true {
			t.Error("Two leader at the same time")
		}
		time.Sleep(time.Millisecond * 50)
	}
	time.Sleep(time.Second * 2)
}

func TestRaft(t *testing.T) {
	// This is the sockets for communicate with Raft server
	requester := make([]*zmq.Socket, 5)
	// This variable contain status of Raft server whether it 'up' or 'down'
	up := make([]bool, 5)
	// Msg for different servers
	msg := make([]int, 5) // 0 :- exit, 1:- what_phase
	// This is the phase of Different Raft server at that time
	phase := make([]string, 5)
	// Which test case
	t_case := 0
	port := 5410
	for i := 0; i < 5; i++ {
		port_i := port + i
		go execl(i+1, port_i)
		requester[i], _ = zmq.NewSocket(zmq.REQ)
		requester[i].Connect("tcp://localhost:" + fmt.Sprintf("%d", port_i))
		up[i] = true
		msg[i] = 1
	}

	// Test Case :- 1
	// 1st test case check one and only one leader at a time
	t_case = 1
	run(msg, requester, up, phase, t_case, t)

	// Test Case :- 2
	// Now a Follower is dead
	// check phase after each 50 milisecond
	t_case = 2
	for i := 0; i < 5; i++ {
		if phase[i] == "Follower" {
			send(0, requester[i])
			up[i] = false
			break
		}
	}
	run(msg, requester, up, phase, t_case, t)

	// Test Case :- 3
	// Now a Leader is Down and minority no of failure
	// check phase after each 50 milisecond
	t_case = 3
	for i := 0; i < 5; i++ {
		if phase[i] == "Leader" {
			phase[i] = "Follower"
			send(0, requester[i])
			up[i] = false
			break
		}
	}
	run(msg, requester, up, phase, t_case, t)

	// Test Case :- 4
	// Now a Leader is Down and majority no of failure this tym no one is Leader
	// check phase after each 50 milisecond
	t_case = 4
	for i := 0; i < 5; i++ {
		if phase[i] == "Leader" {
			phase[i] = "Follower"
			send(0, requester[i])
			up[i] = false
			break
		}
	}
	run(msg, requester, up, phase, t_case, t)

	// Test Case :- 5
	// Now a Raft is UP and minority no of failure this tym choose one as Leader
	// check phase after each 50 milisecond
	t_case = 5
	for i := 0; i < 5; i++ {
		if up[i] == false {
			phase[i] = "Follower"
			go execl(i+1, port+i)
			time.Sleep(time.Millisecond * 10)
			requester[i].Close()
			time.Sleep(time.Millisecond * 100)
			requester[i], _ = zmq.NewSocket(zmq.REQ)
			requester[i].Connect("tcp://localhost:" + fmt.Sprintf("%d", port+i))
			time.Sleep(time.Millisecond * 10)
			up[i] = true
			break
		}
	}
	run(msg, requester, up, phase, t_case, t)

	// Test Case :- 6
	// Now a Follower is down so Leader is old one and no effect on leader
	// check phase after each 50 milisecond
	t_case = 6
	for i := 0; i < 5; i++ {
		if up[i] == true {
			if phase[i] == "Follower" {
				send(0, requester[i])
				up[i] = false
				break
			}
		}
	}
	run(msg, requester, up, phase, t_case, t)

	// Now down all the server and their respective process
	for i := 0; i < 5; i++ {
		if up[i] == true {
			send(0, requester[i])
			up[i] = false
		}
	}

	// Some Wait so is it can do their all task
	time.Sleep(time.Second * 2)
}
