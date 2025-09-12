package events

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrEventBusStopped      = errors.New("event bus is stopped")
	ErrInvalidEventType     = errors.New("invalid event type")
)

// EventBus 事件总线接口
type EventBus interface {
	// 发布事件
	Publish(ctx context.Context, event Event) error
	PublishAsync(event Event)

	// 订阅事件
	Subscribe(eventType string, handler EventHandler) SubscriptionID
	SubscribeWithFilter(eventType string, filter EventFilter, handler EventHandler) SubscriptionID

	// 取消订阅
	Unsubscribe(id SubscriptionID) error

	// 生命周期管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool
}

// EventBusConfig 事件总线配置
type EventBusConfig struct {
	WorkerPoolSize   int
	BufferSize       int
	MaxRetries       int
	RetryDelay       time.Duration
	ProcessTimeout   time.Duration
	EnableMetrics    bool
	EnablePersist    bool
	DeadLetterQueue  bool
}

// DefaultEventBusConfig 默认配置
func DefaultEventBusConfig() *EventBusConfig {
	return &EventBusConfig{
		WorkerPoolSize:  10,
		BufferSize:      1000,
		MaxRetries:      3,
		RetryDelay:      time.Second * 5,
		ProcessTimeout:  time.Minute * 5,
		EnableMetrics:   true,
		EnablePersist:   true,
		DeadLetterQueue: true,
	}
}

// eventBus 事件总线实现
type eventBus struct {
	config              *EventBusConfig
	subscriptionManager *SubscriptionManager
	eventStore          EventStore
	processor           *EventProcessor
	deadLetterQueue     DeadLetterQueue
	metrics             *EventMetrics

	// 异步事件队列
	asyncQueue chan Event

	// 控制状态
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex
}

// NewEventBus 创建新的事件总线
func NewEventBus(config *EventBusConfig, options ...EventBusOption) EventBus {
	if config == nil {
		config = DefaultEventBusConfig()
	}

	bus := &eventBus{
		config:              config,
		subscriptionManager: NewSubscriptionManager(),
		asyncQueue:          make(chan Event, config.BufferSize),
	}

	// 应用选项
	for _, option := range options {
		option(bus)
	}

	// 初始化组件
	if bus.processor == nil {
		bus.processor = NewEventProcessor(config, bus.subscriptionManager)
	}

	if config.EnableMetrics && bus.metrics == nil {
		bus.metrics = NewEventMetrics()
	}

	return bus
}

// EventBusOption 事件总线选项
type EventBusOption func(*eventBus)

// WithEventStore 设置事件存储
func WithEventStore(store EventStore) EventBusOption {
	return func(bus *eventBus) {
		bus.eventStore = store
	}
}

// WithDeadLetterQueue 设置死信队列
func WithDeadLetterQueue(dlq DeadLetterQueue) EventBusOption {
	return func(bus *eventBus) {
		bus.deadLetterQueue = dlq
	}
}

// WithMetrics 设置指标收集器
func WithMetrics(metrics *EventMetrics) EventBusOption {
	return func(bus *eventBus) {
		bus.metrics = metrics
	}
}

// Start 启动事件总线
func (b *eventBus) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return nil
	}

	b.ctx, b.cancel = context.WithCancel(ctx)
	b.running = true

	// 启动异步队列处理器
	b.wg.Add(1)
	go b.processAsyncQueue()

	log.Println("Event bus started")
	return nil
}

// Stop 停止事件总线
func (b *eventBus) Stop(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return nil
	}

	b.cancel()
	b.running = false

	// 等待所有工作完成或超时
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Event bus stopped gracefully")
	case <-ctx.Done():
		log.Println("Event bus stopped with timeout")
	}

	return nil
}

// IsRunning 检查是否运行中
func (b *eventBus) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// Publish 同步发布事件
func (b *eventBus) Publish(ctx context.Context, event Event) error {
	if !b.IsRunning() {
		return ErrEventBusStopped
	}

	// 记录指标
	if b.metrics != nil {
		b.metrics.EventPublished(event.GetType())
	}

	// 持久化事件
	if b.eventStore != nil && b.config.EnablePersist {
		if err := b.eventStore.Save(event); err != nil {
			log.Printf("Failed to persist event %s: %v", event.GetID(), err)
		}
	}

	// 同步处理事件
	return b.processor.Process(ctx, event)
}

// PublishAsync 异步发布事件
func (b *eventBus) PublishAsync(event Event) {
	if !b.IsRunning() {
		log.Printf("Event bus not running, dropping event: %s", event.GetID())
		return
	}

	select {
	case b.asyncQueue <- event:
		// 成功加入队列
	default:
		// 队列已满，记录错误
		log.Printf("Async queue full, dropping event: %s", event.GetID())
		if b.metrics != nil {
			b.metrics.EventDropped(event.GetType())
		}
	}
}

// Subscribe 订阅事件
func (b *eventBus) Subscribe(eventType string, handler EventHandler) SubscriptionID {
	return b.subscriptionManager.Subscribe(eventType, handler, nil, PriorityNormal)
}

// SubscribeWithFilter 带过滤器订阅事件
func (b *eventBus) SubscribeWithFilter(eventType string, filter EventFilter, handler EventHandler) SubscriptionID {
	return b.subscriptionManager.Subscribe(eventType, handler, filter, PriorityNormal)
}

// Unsubscribe 取消订阅
func (b *eventBus) Unsubscribe(id SubscriptionID) error {
	return b.subscriptionManager.Unsubscribe(id)
}

// processAsyncQueue 处理异步队列
func (b *eventBus) processAsyncQueue() {
	defer b.wg.Done()

	for {
		select {
		case event := <-b.asyncQueue:
			// 处理异步事件
			if err := b.Publish(b.ctx, event); err != nil {
				log.Printf("Failed to process async event %s: %v", event.GetID(), err)
			}

		case <-b.ctx.Done():
			// 处理剩余事件
			b.drainQueue()
			return
		}
	}
}

// drainQueue 清空队列
func (b *eventBus) drainQueue() {
	for {
		select {
		case event := <-b.asyncQueue:
			if err := b.Publish(context.Background(), event); err != nil {
				log.Printf("Failed to process remaining event %s: %v", event.GetID(), err)
			}
		default:
			return
		}
	}
}