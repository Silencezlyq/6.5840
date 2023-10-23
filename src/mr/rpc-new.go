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
const Done = 4

var mu sync.mutex

func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}