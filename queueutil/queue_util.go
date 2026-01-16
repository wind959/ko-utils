package queueutil

import (
	"errors"
	"sync"
	"time"
)

// 错误定义
var (
	ErrQueueClosed = errors.New("queue closed")
	ErrQueueFull   = errors.New("queue full")
	ErrTimeout     = errors.New("timeout")
	ErrNoItem      = errors.New("no item available")
)

// Queue 线程安全的内存队列
type Queue[T any] struct {
	items     chan T
	closeOnce sync.Once
	closed    chan struct{}
}

// NewQueue 创建队列
// capacity: 队列容量，0=无缓冲（同步），>0=缓冲队列
func NewQueue[T any](capacity int) *Queue[T] {
	if capacity < 0 {
		capacity = 0
	}
	return &Queue[T]{
		items:  make(chan T, capacity),
		closed: make(chan struct{}),
	}
}

// Put 放入元素（阻塞直到成功或队列关闭）
func (q *Queue[T]) Put(item T) error {
	select {
	case q.items <- item:
		return nil
	case <-q.closed:
		return ErrQueueClosed
	}
}

// TryPut 尝试放入元素（非阻塞）
func (q *Queue[T]) TryPut(item T) error {
	select {
	case q.items <- item:
		return nil
	case <-q.closed:
		return ErrQueueClosed
	default:
		return ErrQueueFull
	}
}

// PutWithTimeout 放入元素（带超时）
func (q *Queue[T]) PutWithTimeout(item T, timeout time.Duration) error {
	if timeout <= 0 {
		return q.TryPut(item)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case q.items <- item:
		return nil
	case <-q.closed:
		return ErrQueueClosed
	case <-timer.C:
		return ErrTimeout
	}
}

// PutBatch 批量放入元素
func (q *Queue[T]) PutBatch(items []T) error {
	for _, item := range items {
		if err := q.Put(item); err != nil {
			return err
		}
	}
	return nil
}

// Get 获取元素（阻塞直到有元素或队列关闭）
func (q *Queue[T]) Get() (T, error) {
	var zero T
	select {
	case item, ok := <-q.items:
		if !ok {
			return zero, ErrQueueClosed
		}
		return item, nil
	case <-q.closed:
		return zero, ErrQueueClosed
	}
}

// TryGet 尝试获取元素（非阻塞）
// 返回元素、是否成功、错误信息
func (q *Queue[T]) TryGet() (item T, ok bool, err error) {
	select {
	case item, ok := <-q.items:
		if !ok {
			var zero T
			return zero, false, ErrQueueClosed
		}
		return item, true, nil
	case <-q.closed:
		var zero T
		return zero, false, ErrQueueClosed
	default:
		var zero T
		return zero, false, nil // 没取到，但不是错误
	}
}

func (q *Queue[T]) TryGetSimple() (T, error) {
	var zero T
	item, ok, err := q.TryGet()
	if err != nil {
		return zero, err
	}
	if !ok {
		return zero, ErrNoItem
	}
	return item, nil
}

// GetWithTimeout 获取元素（带超时）
func (q *Queue[T]) GetWithTimeout(timeout time.Duration) (T, error) {
	var zero T
	if timeout <= 0 {
		return q.TryGetSimple()
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case item, ok := <-q.items:
		if !ok {
			return zero, ErrQueueClosed
		}
		return item, nil
	case <-q.closed:
		return zero, ErrQueueClosed
	case <-timer.C:
		return zero, ErrTimeout
	}
}

// GetBatch 批量获取元素
func (q *Queue[T]) GetBatch(max int) ([]T, error) {
	if max <= 0 {
		return []T{}, nil
	}
	items := make([]T, 0, max)
	for i := 0; i < max; i++ {
		item, ok, err := q.TryGet()
		if err != nil {
			return items, err
		}
		if !ok {
			break // 没有更多元素了
		}
		items = append(items, item)
	}
	return items, nil
}

// Close 关闭队列
func (q *Queue[T]) Close() {
	q.closeOnce.Do(func() {
		close(q.closed)
		close(q.items)
	})
}

// IsClosed 检查队列是否已关闭
func (q *Queue[T]) IsClosed() bool {
	select {
	case <-q.closed:
		return true
	default:
		return false
	}
}

// Len 当前队列长度
func (q *Queue[T]) Len() int {
	return len(q.items)
}

// Cap 队列容量
func (q *Queue[T]) Cap() int {
	return cap(q.items)
}

// IsEmpty 队列是否为空
func (q *Queue[T]) IsEmpty() bool {
	return q.Len() == 0
}

// IsFull 队列是否已满
func (q *Queue[T]) IsFull() bool {
	return q.Len() == q.Cap()
}

// Range 遍历队列所有元素
func (q *Queue[T]) Range(fn func(item T) bool) {
	for {
		select {
		case item, ok := <-q.items:
			if !ok {
				return
			}
			if !fn(item) {
				return
			}
		case <-q.closed:
			return
		default:
			return
		}
	}
}

// Clear 清空队列并返回清空的数量
func (q *Queue[T]) Clear() int {
	if q.IsClosed() {
		return 0
	}
	count := 0
	for {
		select {
		case _, ok := <-q.items:
			if !ok {
				return count
			}
			count++
		default:
			return count
		}
	}
}

// ClearWithHandler 清空队列并对每个元素执行处理
func (q *Queue[T]) ClearWithHandler(handler func(item T)) int {
	if q.IsClosed() {
		return 0
	}

	count := 0
	for {
		select {
		case item, ok := <-q.items:
			if !ok {
				return count
			}
			count++
			if handler != nil {
				handler(item)
			}
		default:
			return count
		}
	}
}

// RetryPut 重试放入（失败后延迟重试）
func (q *Queue[T]) RetryPut(item T, maxRetries int, delay time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		err := q.TryPut(item)
		if err == nil {
			return nil
		}

		if errors.Is(err, ErrQueueClosed) {
			return err
		}

		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}
	return ErrQueueFull
}

// ProcessWithRetry 处理元素并自动重试
func (q *Queue[T]) ProcessWithRetry(
	processor func(item T) error,
	maxRetries int,
	retryDelay time.Duration,
) error {
	for {
		item, err := q.Get()
		if err != nil {
			return err
		}

		for retry := 0; retry < maxRetries; retry++ {
			if err := processor(item); err == nil {
				break
			}

			if retry < maxRetries-1 {
				time.Sleep(retryDelay)
				continue
			}

			// 最终失败，重新放回队列
			if putErr := q.Put(item); putErr != nil {
				return putErr
			}
			time.Sleep(retryDelay)
		}
	}
}

// Drain 获取并清空所有元素
func (q *Queue[T]) Drain() []T {
	items := make([]T, 0, q.Len())
	for {
		select {
		case item, ok := <-q.items:
			if !ok {
				return items
			}
			items = append(items, item)
		default:
			return items
		}
	}
}
