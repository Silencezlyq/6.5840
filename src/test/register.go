package main

import (
	"fmt"
	"time"
	"math/rand"
	"sync"
)

type Publisher struct {
	//申请读写锁和管道
	mu           sync.RWMutex
	subscribers  map[chan string]struct{}
	once 		 sync.Once
}

func NewPublisher() *Publisher {
	return &Publisher{
		subscribers: make(map[chan string]struct{}),
	}
}

func (p *Publisher) Subscribe(subscriber chan string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers[subscriber] = struct{}{}
}

func (p *Publisher) Unsubscribe(subscriber chan string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.subscribers, subscriber)
	p.once.Do(func(){
		close(subscriber)
	})
}

func (p *Publisher) Publish(message string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for subscriber := range p.subscribers {
		subscriber <- message
	}
}


func main() {
	//发布者，管理订阅者并发布消息
	publisher := NewPublisher()

	rand.Seed(time.Now().UnixNano())

	// 生成一个随机整数，范围在 [0, 100)
	randomInt := rand.Intn(100)

	subscriber1 := make(chan string)
	subscriber2 := make(chan string)

	publisher.Subscribe(subscriber1)
	publisher.Subscribe(subscriber2)

	go func() {
		for msg := range subscriber1 {
			fmt.Println("Subscriber 1 received:", msg)
		}
	}()

	go func() {
		for msg := range subscriber2 {
			fmt.Println("Subscriber 2 received:", msg)
		}
	}()

	go func(){
		for {
			randomInt = rand.Intn(100)
			if randomInt % 5 == 0{
				publisher.Unsubscribe(subscriber1)
			}else if randomInt % 5 == 1{
				publisher.Unsubscribe(subscriber2)
			}else{
				publisher.Publish("waiting!")
			}
			time.Sleep(2*time.Second)
		}
	}()

	publisher.Publish("Hello, subscribers!")

	// Wait for a while to allow subscribers to process messages
	<-time.After(20 * time.Second)

	// publisher.Unsubscribe(subscriber1)
	// publisher.Unsubscribe(subscriber2)

	numbers := []int{1, 2, 3, 4, 5}

    // 删除索引为2的元素
    indexToDelete := 4
    if indexToDelete >= 0 && indexToDelete < len(numbers) {
        numbers = append(numbers[:indexToDelete], numbers[indexToDelete+1:]...)
    } else {
        fmt.Println("Invalid index to delete")
    }

    // 遍历切片
    for _, number := range numbers {
        fmt.Println(number)
    }
}
