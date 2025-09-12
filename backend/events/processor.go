package events

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

// EventProcessor 事件处理器
type EventProcessor struct {
	config              *EventBusConfig
	subscriptionManager *SubscriptionManager
	workerPool          *WorkerPool
	retryQueue          chan *RetryEvent
	metrics             *EventMetrics

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// RetryEvent 重试事件
type RetryEvent struct {
	Event       Event
	Subscription *Subscription
	Attempt     int
	LastError   error
	NextRetry   time.Time
}

// NewEventProcessor 创建新的事件处理器
func NewEventProcessor(config *EventBusConfig, subscriptionManager *SubscriptionManager) *EventProcessor {
	return &EventProcessor{
		config:              config,
		subscriptionManager: subscriptionManager,
		workerPool:          NewWorkerPool(config.WorkerPoolSize),
		retryQueue:          make(chan *RetryEvent, config.BufferSize),
	}
}

// Process 处理事件
func (ep *EventProcessor) Process(ctx context.Context, event Event) error {
	eventType := event.GetType()
	subscriptions := ep.subscriptionManager.FilterSubscriptions(eventType, event)

	if len(subscriptions) == 0 {
		if ep.metrics != nil {
			ep.metrics.EventProcessed(eventType, "no_subscribers")
		}
		return nil
	}

	var errors []error
	
	for _, subscription := range subscriptions {
		if subscription.Async {
			// 异步处理
			ep.processAsync(ctx, event, subscription)
		} else {
			// 同步处理
			if err := ep.processSync(ctx, event, subscription); err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		if ep.metrics != nil {
			ep.metrics.EventProcessed(eventType, "error")
		}
		return fmt.Errorf("event processing errors: %v", errors)
	}

	if ep.metrics != nil {
		ep.metrics.EventProcessed(eventType, "success")
	}

	return nil
}

// processSync 同步处理事件
func (ep *EventProcessor) processSync(ctx context.Context, event Event, subscription *Subscription) error {
	startTime := time.Now()
	
	// 设置超时上下文
	if ep.config.ProcessTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ep.config.ProcessTimeout)
		defer cancel()
	}

	// 处理事件
	err := ep.handleEventWithRecovery(ctx, event, subscription)
	
	// 记录处理时间
	if ep.metrics != nil {
		ep.metrics.EventLatency(event.GetType(), time.Since(startTime))
	}

	// 如果失败且配置了重试，加入重试队列
	if err != nil && subscription.MaxRetries > 0 {
		retryEvent := &RetryEvent{
			Event:        event,
			Subscription: subscription,
			Attempt:      1,
			LastError:    err,
			NextRetry:    time.Now().Add(subscription.RetryDelay),
		}
		
		select {
		case ep.retryQueue <- retryEvent:
		default:
			log.Printf("Retry queue full, dropping retry for event %s", event.GetID())
		}
	}

	return err
}

// processAsync 异步处理事件
func (ep *EventProcessor) processAsync(ctx context.Context, event Event, subscription *Subscription) {
	task := &WorkerTask{
		Event:        event,
		Subscription: subscription,
		Context:      ctx,
	}
	
	ep.workerPool.Submit(task)
}

// handleEventWithRecovery 带恢复机制的事件处理
func (ep *EventProcessor) handleEventWithRecovery(ctx context.Context, event Event, subscription *Subscription) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// 获取堆栈信息
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stack := string(buf[:n])
			
			err = fmt.Errorf("panic in event handler: %v\nStack: %s", r, stack)
			log.Printf("Panic in event handler for event %s: %v\n%s", event.GetID(), r, stack)
			
			if ep.metrics != nil {
				ep.metrics.EventError(event.GetType(), "panic")
			}
		}
	}()

	return subscription.Handler(event)
}

// Start 启动处理器
func (ep *EventProcessor) Start(ctx context.Context) error {
	ep.ctx, ep.cancel = context.WithCancel(ctx)
	
	// 启动工作池
	ep.workerPool.Start(ep.ctx)
	
	// 启动重试处理器
	ep.wg.Add(1)
	go ep.processRetryQueue()
	
	return nil
}

// Stop 停止处理器
func (ep *EventProcessor) Stop(ctx context.Context) error {
	if ep.cancel != nil {
		ep.cancel()
	}
	
	// 停止工作池
	ep.workerPool.Stop(ctx)
	
	// 等待重试处理器停止
	done := make(chan struct{})
	go func() {
		ep.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// processRetryQueue 处理重试队列
func (ep *EventProcessor) processRetryQueue() {
	defer ep.wg.Done()
	
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	
	var retryEvents []*RetryEvent
	
	for {
		select {
		case retryEvent := <-ep.retryQueue:
			retryEvents = append(retryEvents, retryEvent)
			
		case <-ticker.C:
			// 处理重试事件
			now := time.Now()
			remaining := retryEvents[:0]
			
			for _, retryEvent := range retryEvents {
				if now.After(retryEvent.NextRetry) {
					ep.processRetry(retryEvent)
				} else {
					remaining = append(remaining, retryEvent)
				}
			}
			
			retryEvents = remaining
			
		case <-ep.ctx.Done():
			return
		}
	}
}

// processRetry 处理重试事件
func (ep *EventProcessor) processRetry(retryEvent *RetryEvent) {
	err := ep.handleEventWithRecovery(ep.ctx, retryEvent.Event, retryEvent.Subscription)
	
	if err == nil {
		// 重试成功
		if ep.metrics != nil {
			ep.metrics.EventRetrySuccess(retryEvent.Event.GetType())
		}
		return
	}
	
	// 重试失败
	retryEvent.Attempt++
	retryEvent.LastError = err
	
	if retryEvent.Attempt <= retryEvent.Subscription.MaxRetries {
		// 继续重试
		retryEvent.NextRetry = time.Now().Add(retryEvent.Subscription.RetryDelay)
		
		select {
		case ep.retryQueue <- retryEvent:
		default:
			log.Printf("Retry queue full, dropping retry for event %s", retryEvent.Event.GetID())
		}
		
		if ep.metrics != nil {
			ep.metrics.EventRetryFailed(retryEvent.Event.GetType())
		}
	} else {
		// 重试次数用完，发送到死信队列
		log.Printf("Event %s failed after %d retries, sending to dead letter queue", 
			retryEvent.Event.GetID(), retryEvent.Attempt)
		
		if ep.metrics != nil {
			ep.metrics.EventDeadLetter(retryEvent.Event.GetType())
		}
	}
}