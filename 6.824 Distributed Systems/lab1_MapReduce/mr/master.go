package mr

import (
	"container/list"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const NA = -1

// result
const Error = -1
const Done = 1

// status
const PleaseWait = 0
const Mapping = 1
const MapComplete = 2
const Reducing = 3
const ReduceComplete = 4

type MapTask struct {
	worker_id int
	start_at  time.Time
	filename  string
}

type ReduceTask struct {
	worker_id int
	start_at  time.Time
	reduce_id int
}

type Master struct {
	// Your definitions here.
	NReduce           int
	files             []string
	map_waiting       list.List
	map_processing    list.List
	map_done          list.List
	reduce_waiting    list.List
	reduce_processing list.List
	reduce_done       list.List
	register          []int
	ack_reducing      []int

	info_lock sync.Mutex
}

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (m *Master) Setting(args *WorkerArgs, reply *SettingReply) error {
	m.info_lock.Lock()
	m.register = append(m.register, args.PID)
	m.info_lock.Unlock()
	reply.NReduce = m.NReduce
	return nil
}

func (m *Master) ResponseMap(args *WorkerArgs, reply *RequestMapReply) error {
	m.info_lock.Lock()
	// if m.map_waiting.Len == 0 && m.map_processing.Len == 0 {
	// 	m.info_lock.Unlock()
	// 	return errors.New(fmt.Sprintf("Process %v start bucketing", args.PID))
	// }
	// fmt.Printf("Receive process %v's mapping request at %v\n", args.PID, time.Now())
	if m.map_waiting.Len() > 0 {
		doing := m.map_waiting.Front()
		m.map_waiting.Remove(doing)
		maptask, _ := (doing.Value).(MapTask)
		maptask.start_at = time.Now()
		maptask.worker_id = args.PID
		reply.FileName = maptask.filename
		reply.Status = Mapping
		m.map_processing.PushBack(maptask)
	} else if m.map_processing.Len() > 0 {
		tag := true
		for doing := m.map_processing.Front(); doing != nil; doing = doing.Next() {
			// do something with e.Value
			maptask, _ := (doing.Value).(MapTask)
			if time.Now().Sub(maptask.start_at) > 10*time.Second {
				fmt.Printf("Process %v fails mapping at %v \n", maptask.worker_id, time.Now())
				maptask.start_at = time.Now()
				maptask.worker_id = args.PID
				reply.FileName = maptask.filename
				reply.Status = Mapping
				tag = false
				m.map_processing.Remove(doing)
				m.map_processing.PushBack(maptask)
				break
			}
		}
		if tag {
			reply.FileName = ""
			reply.Status = PleaseWait
		}

	} else {
		reply.FileName = ""
		reply.Status = MapComplete
	}

	m.info_lock.Unlock()

	return nil
}

func (m *Master) ResponseMapReport(args *ReportMapArgs, reply *ReportReply) error {
	m.info_lock.Lock()
	reply.Status = Error
	for doing := m.map_processing.Front(); doing != nil; doing = doing.Next() {
		maptask, _ := (doing.Value).(MapTask)
		if maptask.filename == args.FileName && maptask.worker_id == args.PID {
			m.map_processing.Remove(doing)
			maptask.start_at = time.Now()
			m.map_done.PushBack(maptask)
			reply.Status = Done
			break
		}
	}

	m.info_lock.Unlock()

	return nil
}

func (m *Master) ResponseReduce(args *WorkerArgs, reply *RequestReduceReply) error {
	m.info_lock.Lock()

	if m.reduce_waiting.Len() > 0 {
		doing := m.reduce_waiting.Front()
		m.reduce_waiting.Remove(doing)
		reducetask, _ := (doing.Value).(ReduceTask)
		reducetask.start_at = time.Now()
		reducetask.worker_id = args.PID
		reply.ReduceId = reducetask.reduce_id
		reply.Status = Reducing
		m.reduce_processing.PushBack(reducetask)
	} else if m.reduce_processing.Len() > 0 {
		tag := true
		for doing := m.reduce_processing.Front(); doing != nil; doing = doing.Next() {
			reducetask, _ := (doing.Value).(ReduceTask)
			if time.Now().Sub(reducetask.start_at).Seconds() > 10 {
				fmt.Printf("Process %v fails reducing at %v \n", reducetask.worker_id, time.Now())
				reducetask.start_at = time.Now()
				reducetask.worker_id = args.PID
				reply.ReduceId = reducetask.reduce_id
				reply.Status = Reducing
				tag = false
				m.reduce_processing.Remove(doing)
				m.reduce_processing.PushBack(reducetask)
				break
			}
		}
		if tag {
			reply.ReduceId = NA
			reply.Status = PleaseWait
		}
	} else {
		reply.ReduceId = NA
		reply.Status = ReduceComplete
	}

	m.info_lock.Unlock()

	return nil
}

func (m *Master) ResponseReduceReport(args *ReportReduceArgs, reply *ReportReply) error {
	m.info_lock.Lock()
	reply.Status = Error
	for doing := m.reduce_processing.Front(); doing != nil; doing = doing.Next() {
		reducetask, _ := (doing.Value).(ReduceTask)
		if reducetask.reduce_id == args.ReduceId && reducetask.worker_id == args.PID {
			m.reduce_processing.Remove(doing)
			reducetask.start_at = time.Now()
			m.reduce_done.PushBack(reducetask)
			reply.Status = Done
			break
		}
	}

	m.info_lock.Unlock()

	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	ret := false

	m.info_lock.Lock()
	if m.reduce_done.Len() == m.NReduce {
		ret = true
	}
	m.info_lock.Unlock()

	return ret
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	// Your code here.
	m.NReduce = nReduce
	m.info_lock.Lock()
	m.files = files
	for _, filename := range m.files {
		cur := MapTask{}
		cur.worker_id = NA
		cur.start_at = time.Now()
		cur.filename = filename
		m.map_waiting.PushBack(cur)
	}
	i := 0
	for i < nReduce {
		cur := ReduceTask{}
		cur.worker_id = NA
		cur.start_at = time.Now()
		cur.reduce_id = i
		m.reduce_waiting.PushBack(cur)
		i++
	}
	m.info_lock.Unlock()
	fmt.Printf("reduce is: %d\n", m.NReduce)
	m.server()
	return &m
}
