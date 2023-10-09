package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
// 就是一个简单的自动生成hash值的函数
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// 向Master发出请求，获取任务，设置对应状态
func GetTask(reply *Task) bool {
	args := ExampleArgs{}
	if reply.TaskType == Waiting {
		args.X = Waiting
	}
	//向Master发出请求
	ok := call("Coordinator.PostTask", &args, reply)
	return ok
}

// 封装了完成任务的函数，设置对应状态
func FinishTask(task *Task) {
	reply := ExampleReply{}
	call("Coordinator.FinishTask", task, &reply)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	flag := false

	i := 1

	for i == 1 {
		task := Task{}
		//任务完成之后置为等待状态
		if flag {
			task.TaskType = Waiting
			flag = false
		}
		if GetTask(&task) {
			switch task.TaskType {
			case Waiting:
				{
					time.Sleep(time.Second)
					flag = true
				}
			case Map:
				{
					DoMap(mapf, &task)
					FinishTask(&task)
				}
			case Reduce:
				{
					DoReduce(reducef, &task)
					FinishTask(&task)
				}
			default:
				{
					i = 2
				}
			}
		} else {
			break
		}
	}

	//这里应该写出错的处理，之后补全。

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()

}

// 定义一个KV结构体数组，并设置各种函数，类似类定义的操作(为了实现Sort类接口，使用sort排序方法)。
type ByKey []KeyValue

func (a ByKey) Len() int {
	return len(a)
}
func (a ByKey) Swap(i int, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByKey) Less(i, j int) bool {
	return a[i].Key < a[j].Key
}

// 从中间文件读取所有值，统计数量，进行reduce任务，最终输出结果。
func DoReduce(reducef func(string, []string) string, task *Task) {
	//同样从本地直接读取KeyValue文件
	i := 0
	intermediate := []KeyValue{}
	for i < task.NReduce {
		filepath := task.Filename + strconv.Itoa(i) + "-" + strconv.Itoa(task.TaskId)
		file, _ := os.Open(filepath)
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			intermediate = append(intermediate, kv)
		}
		file.Close()
		i++
	}
	sort.Sort(ByKey(intermediate))

	i = 0
	oname := "mr-out-" + strconv.Itoa(task.TaskID)
	ofile, _ := os.Create(oname)
	for i < len(intermediate) {
		j := i + 1
		//排好序了，统计重复词数据
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		//
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		//wc.go中实现的reducefunc就是给定values的len直接获取的
		output := reducef(intermediate[i].Key, values)

		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)
		i = j
	}

}

// 主要是要读取所有输入文件，将文件传给map函数，计算结果并输出
func DoMap(mapf func(string, string) []KeyValue, task *Task) {
	//文件难道不是RPC传来的嘛？？这里是本地读，相当于同一台机器上，他只传了路径。
	intermidiate := []KeyValue{}
	Filename := task.Filename
	file, err := os.Open(Filename)
	if err != nil {
		log.Fatalf("cannot open %v", Filename)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", Filename)
	}
	file.Close()

	kva := mapf(Filename, string(content))
	//将kva元素展开，追加到intermidiate切片里
	intermidiate = append(intermidiate, kva...)

	//创建NReduce个[]KeyValue数组的map（整个MAP过程？）
	intermidiateMap := make([][]KeyValue, task.NReduce)

	for _, tmp := range intermidiate {
		//任务分配
		intermidiateMap[ihash(tmp.Key)%task.NReduce] = append(intermidiateMap[ihash(tmp.Key)%task.NReduce], tmp)
	}
	//用排序的方式，进行word统计
	sort.Sort(ByKey(intermidiate))

	//把每个任务写入文档中，也是本地写？？有些奇怪
	i := 0
	for i < task.NReduce {
		oname := "mr-" + strconv.Itoa(task.TaskID) + "-" + strconv.Itoa(i)
		ofile, _ := os.Create(oname)
		enc := json.NewEncoder(ofile)
		for _, kv := range intermidiateMap[i] {
			enc.Encode(&kv)
		}
		i++
		ofile.Close()
	}
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
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
