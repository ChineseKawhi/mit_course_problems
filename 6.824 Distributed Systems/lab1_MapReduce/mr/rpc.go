package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type WorkerArgs struct {
	PID int
}

type SettingReply struct {
	NReduce int
}

type RequestMapReply struct {
	FileName string
	Status   int
}

type ReportMapArgs struct {
	PID      int
	FileName string
}

type RequestReduceReply struct {
	ReduceId int
	Status   int
}

type ReportReduceArgs struct {
	PID      int
	ReduceId int
}

type ReportReply struct {
	Status int
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
