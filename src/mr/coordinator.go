package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

// Master节点的结构状态
type Coordinator struct {
	// Your definitions here.
	State       int        //状态信息
	MapChan     chan *Task //Map任务的信道
	ReduceChan  chan *Task //Reduce任务的信道
	MapTasks    []*Task    //Map任务队列
	ReduceTasks []*Task    //Reduce任务队列
	RemainTasks int        //剩余任务数量(Map,Reduce复用)
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
// 这部分主要去实现RPC远程调用功能
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// worker发来请求，获取任务，并通过Task交互，设置状态
func (c *Coordinator) PostTask(arg *ExampleArgs, reply *Task) error {
	//用锁保证并发安全
	mutex.Lock()
	defer mutex.Unlock()
	//通过判断Master状态判断执行流程，因为需要等待所有的Map任务完成之后才会进行Reduce任务
	if c.State == Map {
		//任务队列有任务，就分配
		if len(c.MapChan) > 0 {
			*reply = *<-c.MapChan
			c.MapTasks[reply.TaskID].Start = time.Now()
		} else {
			if arg.X != Waiting {
				reply.TaskType = Waiting
			} else {
				i := 0
				for i < len(c.MapTasks) {
					//判断超时15秒未完成的任务失效了，需要再次分配
					if c.MapTasks[i].Finished == false && time.Since(c.MapTasks[i].Start) > time.Second*15 {
						*reply = *c.MapTasks[i]
						c.MapTasks[i].Start = time.Now()
						break
					}
					i++
				}
			}
		}
	} else if c.State == Reduce {
		//同上
		if len(c.ReduceChan) > 0 {
			*reply = *<-c.ReduceChan
			c.ReduceTasks[reply.TaskID].Start = time.Now()
		} else {
			if arg.X != Waiting {
				reply.TaskType = Waiting
			} else {
				i := 0
				for i < len(c.ReduceTasks) {
					if c.ReduceTasks[i].Finished == false && time.Since(c.ReduceTasks[i].Start) > time.Second*15 {
						*reply = *c.ReduceTasks[i]
						c.ReduceTasks[i].Start = time.Now()
						break
					}
					i++
				}
			}
		}
	} else {
		//都不是则置状态为0表示结束
		reply.TaskType = 0
	}
	return nil
}

// worker发来完成请求，相应完成后进行中间chan的任务设置，并检测
func (c *Coordinator) FinishTask(task *Task, reply *ExampleReply) error {
	if task.TaskType == Map {
		mutex.Lock()
		if c.MapTasks[task.TaskID].Finished == false {
			c.MapTasks[task.TaskID].Finished = true
			c.RemainTasks--
		}
		mutex.Unlock()
	} else if task.TaskType == Reduce {
		mutex.Lock()
		if c.ReduceTasks[task.TaskID].Finished == false {
			c.ReduceTasks[task.TaskID].Finished = true
			c.RemainTasks--
		}
		mutex.Unlock()
	}
	//检查Map任务是否全部完成，如果完成则开始分配Reduce任务
	if c.Check(task.NReduce) {
		//等待Reduce任务完成
		c.State = Waiting
	}
	return nil
}

func (c *Coordinator) Check(NReduce int) bool {
	if c.RemainTasks == 0 {
		if len(c.ReduceTasks) == 0 {
			c.CreateReduce(NReduce)
			return false
		}
		return true
	}
	return false
}

// 创建Nreduce个reduce任务
func (c *Coordinator) CreateReduce(Nreduce int) {
	i := 0
	for i < Nreduce {
		task := Task{
			TaskType: Reduce,
			Filename: "mr-",
			NReduce:  len(c.MapTasks),
			TaskID:   i,
			Finished: false,
			Start:    time.Now(),
		}
		c.ReduceChan <- &task
		c.ReduceTasks = append(c.ReduceTasks, &task)
		i++
	}
	c.State = Reduce
	c.RemainTasks = Nreduce
}

// start a thread that listens for RPCs from worker.go
//socket unix是用本地套接字通信的
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	//状态信息是Map完成之后
	ret := (c.State == Waiting)

	// Your code here.

	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
// 调用的是传来文件和做reduce任务的数量，可以理解为有几台服务器在工作
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		State:       Map,
		MapChan:     make(chan *Task, len(files)),
		ReduceChan:  make(chan *Task, nReduce),
		MapTasks:    make([]*Task, len(files)),
		ReduceTasks: make([]*Task, 0),
		RemainTasks: len(files),
	}
	i := 0
	for _, file := range files {
		task := Task{
			TaskType: Map,
			Filename: file,
			NReduce:  nReduce,
			TaskID:   i,
			Finished: false,
			Start:    time.Now(),
		}
		c.MapChan <- &task
		c.MapTasks[i] = &task
		i++
	}
	// Your code here.
	c.server()
	return &c
}
