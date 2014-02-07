Sir i am sorry i am late in submitting this assignment actually sir i complete this assignment by sunday but i am facing some difficulty to resolve High Water Mark problem. Sir here i use Push and Pull type sockets with DONTWAIT type send which resolve the problem if their are any of the server DOWN but if the receiver are slow then this can drop the packets which are for the server which is UP. To resolve this problem we use following methord which din't work unortunately.
These Methords are :
 
1) I use Push type sockets for bind different server and i use this sockets for send data to other servers of the cloud with no send flag set means it's a blocking type socket when i used . It work well with all the servers UP but if any of the server DOWN then  it block all other servers too because their is a common output channel.And this methord fail.

2) Then i use Router typesocket  to bind and Dealer type socket for connect here i use Router because it can handle the situation of blocking and if any of the server is block then it start droping the data of the other server but keep working with other servers. But if the receiver is slow then sender then this fail too because it start droping the data of servers which are UP .

3) Then finally i use PUSH-PULL combiation with sent DONTWAIT type still i facing the same prolem of  if the receiver is slow then sender then this fail too because it start droping the data of servers which are UP.

4) FUNCTIONS and STRUTURE in cluster.go :-

a) type Cluster struct {
	myid int
	peer [MAX_SERVER]int
	no_of_peer int
	my_cluster *zmq.Socket
	cluster [MAX_SERVER]*zmq.Socket
	inbox chan *Envelope
	outbox chan *Envelope
}

its the objects of server

b) func New(id int,f string) Cluster

This function initialize server by initializing its cluster object

c) func (s Cluster) Send()

This is send function which send data from output channel.

d) func (s Cluster) Receive()

This read data from network and pipe them in to the Inbox channel.



 
