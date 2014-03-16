#Raft Leader Election

`Raft` is a algorithm Raft is a consensus algorithm for managinga replicated log. Here in `Raft Leader Election` i implement only leader election part. In this algorithm one of the raft server work as a leader and here i choose that leader using this algorithm.

#Usage

(note: In the examples below, $ is the shell prompt, we use one argument raft id to run the server "----------------"
#### Code snippet in command line:
```
$ go run github.com/prabhat-bajpai/kvstore-public/tree/master/raft/raft.go raft.id
---------------------------------

```

# Testing

I made 6 different test cases which test the program in 6 different situation and cover all the configuration of raft server at runtime. 

```
go get github.com/prabhat-bajpai/kvstore-public/tree/master/raft
go test github.com/prabhat-bajpai/kvstore-public/tree/master/raft
```

# The `cluster` package

`cluster` Cluster package use in this program is lower layer of raft which provide lower layer interface for seng and receive packets from ather raft server.

### How it works

It just the golang implementation of raft algorithm's leader election part. Here we use a new function which initialize the raft object and make it able to communicate with lower cluster layer.
	The detail discription of other function are write with the function in program.
