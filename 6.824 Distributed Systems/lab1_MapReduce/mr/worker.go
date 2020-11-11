package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strings"
	"time"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return ihash(a[i].Key) < ihash(a[j].Key) }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// setting
	worker_id := os.Getpid()
	nReduce := GetSetting(worker_id)
	fmt.Printf("Process %v gets setting\n", worker_id)
	// map

	for {
		status, filename := RequestMap(worker_id)
		if status == MapComplete {
			fmt.Printf("Process %v finds all mapping finished at %v \n", worker_id, time.Now())
			break
		} else if status == PleaseWait {
			fmt.Printf("Process %v is told to wait others to map\n", worker_id)
			time.Sleep(100 * time.Millisecond)
		} else if status == Mapping {
			fmt.Printf("Process %v starts mapping %v\n", worker_id, filename)
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("cannot open %v", filename)
			}
			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("cannot read %v", filename)
			}
			file.Close()
			m_intermediate := []KeyValue{}
			kva := mapf(filename, string(content))
			m_intermediate = append(m_intermediate, kva...)
			// write m_intermediate
			if len(m_intermediate) > 0 {
				i := 0
				reduce_id := 0
				iname := ""
				bucket := [][]KeyValue{}
				for i < nReduce {
					bucket = append(bucket, []KeyValue{})
					i++
				}
				i = 0
				for i < len(m_intermediate) {
					reduce_id = ihash(m_intermediate[i].Key) % nReduce
					bucket[reduce_id] = append(bucket[reduce_id], m_intermediate[i])
					i++
				}
				i = 0
				for i < nReduce {
					iname = fmt.Sprintf("mr-%v-%d", filename[6:], i)
					ifile, _ := os.Create(iname)
					enc := json.NewEncoder(ifile)
					enc.Encode(bucket[i])
					i++
				}
			}
			status := ReportMap(worker_id, filename)
			if status == Done {
				fmt.Printf("Process %v finishes mapping %v\n", worker_id, filename)
			}
		}
	}

	// reduce
	for {
		status, reduce_id := RequestReduce(worker_id)
		if status == ReduceComplete {
			fmt.Printf("Process %v finds all reducing finished at %v \n", worker_id, time.Now())
			break
		} else if status == PleaseWait {
			fmt.Printf("Process %v is told to wait others to reduce\n", worker_id)
			time.Sleep(100 * time.Millisecond)
		} else if status == Reducing {
			fmt.Printf("Process %v starts reducing %v\n", worker_id, reduce_id)
			files := []string{}
			dir, err := ioutil.ReadDir("../mr-tmp")
			if err != nil {
				log.Fatalf("cannot open %v", dir)
			}
			for _, fi := range dir {
				if fi.IsDir() {
					log.Fatalf("wrong file %v", fi.Name())
				} else {
					suffix := fmt.Sprintf("-%v", reduce_id)
					ok := strings.HasSuffix(fi.Name(), suffix)
					if ok {
						files = append(files, "../mr-tmp/"+fi.Name())
					}
				}
			}
			r_intermediate := []KeyValue{}
			for _, filename := range files {
				file, err := os.Open(filename)
				if err != nil {
					log.Fatalf("cannot open %v", filename)
				}
				temp := []KeyValue{}
				dec := json.NewDecoder(file)
				for {
					if err := dec.Decode(&temp); err != nil {
						break
					}
				}
				// for _, kv := range temp {
				// 	r_intermediate = append(r_intermediate, kv)
				// }
				r_intermediate = append(r_intermediate, temp...)
			}
			sort.Sort(ByKey(r_intermediate))
			oname := fmt.Sprintf("mr-out-%d", reduce_id)
			ofile, _ := os.Create(oname)
			i := 0
			for i < len(r_intermediate) {
				j := i + 1
				for j < len(r_intermediate) && r_intermediate[j].Key == r_intermediate[i].Key {
					j++
				}
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, r_intermediate[k].Value)
				}
				output := reducef(r_intermediate[i].Key, values)

				// this is the correct format for each line of Reduce output.
				fmt.Fprintf(ofile, "%v %v\n", r_intermediate[i].Key, output)

				i = j
			}
			ofile.Close()
			status := ReportReduce(worker_id, reduce_id)
			if status == Done {
				fmt.Printf("Process %v finishes reducing %v\n", worker_id, reduce_id)
			}
		} else {
			return
		}
	}
}

//
// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
//
func GetSetting(pid int) int {
	args := WorkerArgs{}

	// fill in the argument(s).
	args.PID = pid

	// declare a reply structure.
	reply := SettingReply{}

	// send the RPC request, wait for the reply.
	call("Master.Setting", &args, &reply)

	return reply.NReduce
}

func RequestMap(pid int) (int, string) {
	// declare an argument structure.
	args := WorkerArgs{}

	// fill in the argument(s).
	args.PID = pid

	// declare a reply structure.
	reply := RequestMapReply{}

	// send the RPC request, wait for the reply.
	if !call("Master.ResponseMap", &args, &reply) {
		return Error, ""
	}

	return reply.Status, reply.FileName
}

func ReportMap(pid int, filename string) int {
	// declare an argument structure.
	args := ReportMapArgs{}

	// fill in the argument(s).
	args.PID = pid
	args.FileName = filename

	// declare a reply structure.
	reply := ReportReply{}

	// send the RPC request, wait for the reply.
	if !call("Master.ResponseMapReport", &args, &reply) {
		return Error
	}

	return reply.Status
}

func RequestReduce(pid int) (int, int) {
	args := WorkerArgs{}

	// fill in the argument(s).
	args.PID = pid

	// declare a reply structure.
	reply := RequestReduceReply{}

	// send the RPC request, wait for the reply.
	if !call("Master.ResponseReduce", &args, &reply) {
		return Error, NA
	}

	return reply.Status, reply.ReduceId
}

func ReportReduce(pid int, reduce_id int) int {
	args := ReportReduceArgs{}

	// fill in the argument(s).
	args.PID = pid
	args.ReduceId = reduce_id

	// declare a reply structure.
	reply := ReportReply{}

	// send the RPC request, wait for the reply.
	call("Master.ResponseReduceReport", &args, &reply)

	return reply.Status
}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
