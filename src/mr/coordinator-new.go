package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
	"sync"
)

type Coordinator struct{
	State int
	m sync.RWMutex
	MapChan     chan *Task 
	ReduceChan  chan *Task 
	MapTasks    []*Task    //Map任务队列
	ReduceTasks []*Task    //Reduce任务队列
	Connection  map[string][]ConnectionInfo
}

//注册信息结构体，Name是功能的名字，Port是监听的端口号，默认全部使用的都是net.Dial("tcp", Port)去连接。
type ConnectionInfo struct{
	Name string
	Port string
}

func (c *Coordinator) AddConnect(funcName string,keepaliveFunc string,port string) bool{
	//注册功能
	//rpc调用的
	mu.Lock()
	defer mu.Unlock()
	c.Connection[funcName] = append(c.Connection[funcName],ConnectionInfo{Name:keepaliveFunc,Port:port})
	return true
}

//暴力遍历所有的连接信息
func (c *Coordinator) keepAlive(){
	go func(){
		ConnectionNew := make(map[string][]ConnectionInfo)
		for key,values := range c.Connection {
			valuesNew := make([]ConnectionInfo,0)
			for _,value := range values{
				if isServiceAvailable(value.Port){
					valuesNew = append(valuesNew,value)
				}
			}
			if len(valuesNew) != 0{
				ConnectionNew[key] = valuesNew
			}
		}
		c.Connection = ConnectionNew
		time.Sleep(10*time.Second)
	}()
}

//一个简单的hash函数
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func isServiceAvailable(addr string) bool {
    // 假设这里使用 TCP 连接来检查服务可达性
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return false
    }
    conn.Close()
    return true
}

func (c *Coordinator) connection(){
	rpc.Register(c)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		panic("Error starting server: " + e.Error())
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			panic("Error accepting connection: " + err.Error())
		}
		go rpc.ServeConn(conn)
	}
	http.Serve(l,nil)
}

func (c *Coordinator) doMapFunc(name string){
	value, ok := c.Connection[name]
	if ok{
		totalNum := len(value)
		
	}else {
		
	}
}

func (c *Coordinator) server(){
	i:= 1 
	for i == 1{
		switch c.State {
			case Map:{
				doMapFunc()
			}
			case Reduce:{
				doReduceFunc()
			}
			case Waiting:{

			}
			case Done:{
				i = 2
				time.Sleep(2*time.Second)
			}
		}
	}

}

func (c *Coordinator) Done() bool{
	ret := (c.State == Done)
	return ret
}

func MakeCoordinator(files []string, nReduce int) *Coordinator{
	c := Coordinator{
		State: 		Map,
		MapChan: 	make(chan *Task,len(files)),
		ReduceChan: make(chan *Task, nReduce),
		MapTasks:    make([]*Task, len(files)),
		ReduceTasks: make([]*Task, nReduce),
		Connection : make(map[string][]ConnectionInfo,0),
	}
	i := 0
	for _,file := range files{
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

	go c.connection()

	c.keepAlive()

	go c.server()

	return &c
}