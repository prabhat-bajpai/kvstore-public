package cluster

import (
	"fmt"
	"time"
//	"os"
	"strconv"
	"testing"
	"io/ioutil"
)

const (SERVER = 5)
const (NOOFMSG = 1000)
const (NOOFMSGRECV = NOOFMSG*4 + NOOFMSG*4 + 4)



func TestCluster(t *testing.T) {
	var id [SERVER]int
fmt.Println("1")
	for i:=0;i<SERVER;i++{
		id[i]=i+1
	}
//variable type array
	done := make(chan string, SERVER)
	//myid,_ = strconv.Atoi(os.Args[1])

	for _, myid1 := range id {
		go func(myid int) {
			count:=0
  			server := New(myid, /* config file */"peer.json")
			fmt.Println(server)
			go server.Send()
			go server.Receive()

			go func(myid int,count int) {
				for {
					select {
       					case <-  server.Inbox(): 
					count=count+1
       					case <- time.After(10 * time.Second):
					if count != NOOFMSGRECV {
					t.Error("Packets are lost")
					}					
					done <- "1"
					time.After(5 * time.Second)				
   					}
				}
			}(myid,count)
// // This for loop for broadcast msgs 			
				for i:=0 ; i<NOOFMSG ; i++ {
					sender_pid:="-1"			
					id,err := strconv.Atoi(sender_pid)
					msg := fmt.Sprintf("%d",myid)
					if err == nil {
						server.Outbox() <- &Envelope{Pid: id, Msg: msg}
					}
				}
time.Sleep(time.Second)
// This for loop send msgs to individual servers 
				for i:=0 ; i<NOOFMSG ; i++ {
					for _, myid = range id {
						sender_pid:=fmt.Sprintf("%d",myid)		
						id,err := strconv.Atoi(sender_pid)
						msg := fmt.Sprintf("%d",myid)
						if err == nil {
							server.Outbox() <- &Envelope{Pid: id, Msg: msg}
						}
					}
				}

time.Sleep(time.Second)

//Open a file and read in msg here a read a file of 3.4 MB and it receive

 file, e := ioutil.ReadFile("./a.pdf")
if e != nil {
fmt.Printf("File error: %v\n", e)
}
				for _, myid = range id {
						sender_pid:=fmt.Sprintf("%d",myid)		
						id,err := strconv.Atoi(sender_pid)
						msg := string(file)
						if err == nil {
							server.Outbox() <- &Envelope{Pid: id, Msg: msg}
						}
					}
		}(myid1)
}
	count:=0    
	for _ = range done {
		count=count+1
		if count == SERVER {
			break
		}
	}
}
