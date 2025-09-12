package events

import (
	"context"
	"log"
	"time"
	"xsha-backend/database"
)

// TestEventSystem 测试事件系统功能
func TestEventSystem() {
	log.Println("Starting event system test...")

	// 创建事件总线配置
	config := &EventBusConfig{
		WorkerPoolSize:  2,
		BufferSize:      100,
		MaxRetries:      3,
		RetryDelay:      time.Second * 2,
		ProcessTimeout:  time.Minute * 1,
		EnableMetrics:   true,
		EnablePersist:   false, // 测试时不启用持久化
		DeadLetterQueue: false,
	}

	// 创建事件总线
	eventBus := NewEventBus(config)

	// 启动事件总线
	ctx := context.Background()
	if err := eventBus.Start(ctx); err != nil {
		log.Printf("Failed to start event bus: %v", err)
		return
	}

	// 定义测试事件处理器
	testHandler := func(event Event) error {
		log.Printf("Processing test event: %s, Type: %s, Timestamp: %v", 
			event.GetID(), event.GetType(), event.GetTimestamp())
		
		if payload := event.GetPayload(); payload != nil {
			log.Printf("Event payload: %v", payload)
		}
		
		return nil
	}

	// 订阅测试事件
	subscriptionID := eventBus.Subscribe("test.event", testHandler)
	log.Printf("Subscribed to test.event with ID: %s", subscriptionID)

	// 创建并发布测试事件
	testEvent := &BaseEvent{
		ID:        "test-event-1",
		Type:      "test.event", 
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{
			"message": "Hello from event system!",
			"data":    123,
		},
		Metadata: map[string]string{
			"source": "test",
			"version": "1.0",
		},
	}

	// 同步发布事件
	log.Println("Publishing test event synchronously...")
	if err := eventBus.Publish(ctx, testEvent); err != nil {
		log.Printf("Failed to publish event: %v", err)
	}

	// 异步发布事件
	log.Println("Publishing test event asynchronously...")
	eventBus.PublishAsync(testEvent)

	// 等待事件处理完成
	time.Sleep(time.Second * 3)

	// 取消订阅
	if err := eventBus.Unsubscribe(subscriptionID); err != nil {
		log.Printf("Failed to unsubscribe: %v", err)
	} else {
		log.Printf("Successfully unsubscribed from %s", subscriptionID)
	}

	// 停止事件总线
	shutdownCtx, cancel := context.WithTimeout(ctx, time.Second * 10)
	defer cancel()
	
	if err := eventBus.Stop(shutdownCtx); err != nil {
		log.Printf("Failed to stop event bus: %v", err)
	} else {
		log.Println("Event bus stopped successfully")
	}

	log.Println("Event system test completed!")
}

// TestTaskEvents 测试任务事件
func TestTaskEvents() {
	log.Println("Testing task events...")

	// 创建事件总线
	config := DefaultEventBusConfig()
	config.EnablePersist = false
	eventBus := NewEventBus(config)

	ctx := context.Background()
	eventBus.Start(ctx)

	// 创建任务事件处理器
	taskEventHandler := func(event Event) error {
		switch e := event.(type) {
		case *TaskCreatedEvent:
			log.Printf("Task Created: ID=%d, Title=%s, ProjectID=%d", 
				e.TaskID, e.Title, e.ProjectID)
		case *TaskStatusChangedEvent:
			log.Printf("Task Status Changed: ID=%d, %s -> %s", 
				e.TaskID, e.OldStatus, e.NewStatus)
		default:
			log.Printf("Unknown task event type: %s", event.GetType())
		}
		return nil
	}

	// 订阅任务事件
	eventBus.Subscribe(EventTypeTaskCreated, taskEventHandler)
	eventBus.Subscribe(EventTypeTaskStatusChanged, taskEventHandler)

	// 创建模拟任务数据
	mockTask := &database.Task{
		ID:        1,
		Title:     "Test Task",
		ProjectID: 100,
		StartBranch: "main",
		WorkBranch:  "task-1-test",
		CreatedBy:  "test-user",
	}

	// 发布任务创建事件
	createEvent := NewTaskCreatedEvent(mockTask)
	eventBus.PublishAsync(createEvent)

	// 发布任务状态变更事件
	statusEvent := NewTaskStatusChangedEvent(1, 100, "todo", "in_progress", "test-user", "Started working")
	eventBus.PublishAsync(statusEvent)

	// 等待处理完成
	time.Sleep(time.Second * 2)

	// 停止事件总线
	shutdownCtx, cancel := context.WithTimeout(ctx, time.Second * 5)
	defer cancel()
	eventBus.Stop(shutdownCtx)

	log.Println("Task events test completed!")
}