package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
	"sync"
	"time"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.
// 设计实现一个RPC传输的结构
type Task struct {
	TaskType int       //服务器状态信息，对应map\reduce\wait
	TaskID   int       //任务ID
	Filename string    //传输文件名
	NReduce  int       //能卸载的服务器数量？
	Finished bool      //是否完成
	Start    time.Time //开始时间
}

const Map = 1
const Reduce = 2
const Waiting = 3

var mutex sync.Mutex

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
