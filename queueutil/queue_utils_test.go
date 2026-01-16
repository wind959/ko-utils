package queueutil

import (
	"errors"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestQueueUtils(t *testing.T) {
	// 1. 创建队列（容量100）
	q := NewQueue[string](100)
	defer q.Close()

	// 2. 生产者
	go producer(q)

	// 3. 消费者
	go consumer(q)

	// 4. 带重试的生产者
	go retryProducer(q)

	// 5. 带重试的消费者
	go retryConsumer(q)

	time.Sleep(5 * time.Second)
}

type User struct {
	ID   int
	Name string
}

func TestQueueStruct(t *testing.T) {
	userQueue := NewQueue[*User](100)
	defer userQueue.Close()

	// 放入结构体指针
	user := &User{ID: 1, Name: "张三"}
	_ = userQueue.Put(user)

	// 获取
	gotUser, err := userQueue.Get()
	if err != nil {
		// 处理错误
	}
	fmt.Println(gotUser.Name)

	// 批量操作
	users := []*User{
		{ID: 2, Name: "李四"},
		{ID: 3, Name: "王五"},
	}
	userQueue.PutBatch(users)

	// 批量获取
	batch, _ := userQueue.GetBatch(10)
	for _, u := range batch {
		fmt.Println(u.Name)
	}
}

func producer(q *Queue[string]) {
	for i := 0; i < 10; i++ {
		if err := q.Put(fmt.Sprintf("任务-%d", i)); err != nil {
			if errors.Is(err, ErrQueueClosed) {
				log.Println("队列已关闭，生产者退出")
				return
			}
			log.Printf("生产失败: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func consumer(q *Queue[string]) {
	for {
		item, err := q.Get()
		if err != nil {
			if errors.Is(err, ErrQueueClosed) {
				log.Println("队列已关闭，消费者退出")
				return
			}
			log.Printf("消费失败: %v", err)
			continue
		}

		log.Printf("处理: %v", item)
		// 处理业务...
	}
}

func retryProducer(q *Queue[string]) {
	item := "重要任务"

	// 最大重试3次，每次延迟1秒
	if err := q.RetryPut(item, 3, time.Second); err != nil {
		log.Printf("重试生产失败: %v", err)
	}
}

func retryConsumer(q *Queue[string]) {
	// 处理失败时重试3次，每次延迟500ms
	err := q.ProcessWithRetry(func(item string) error {
		log.Printf("处理: %v", item)

		// 模拟随机失败
		if time.Now().UnixNano()%3 == 0 {
			return fmt.Errorf("处理失败")
		}
		return nil
	}, 3, 500*time.Millisecond)

	if err != nil {
		log.Printf("处理失败: %v", err)
	}
}
