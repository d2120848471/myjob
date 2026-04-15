package audit

import (
	"context"
	"sync"

	modelruntime "myjob/internal/model/runtime"
)

// Writer 将操作审计事件写入存储（支持同步/异步缓冲两种模式）。
type Writer struct {
	syncMode bool
	insert   func(context.Context, modelruntime.OperationEvent) error
	ch       chan modelruntime.OperationEvent
	stop     chan struct{}
	wg       sync.WaitGroup
}

// NewWriter 创建一个审计写入器。
//
// - syncMode=true：Write 直接同步落库
// - syncMode=false：Write 写入缓冲队列，由后台协程落库
func NewWriter(syncMode bool, bufferSize int, insert func(context.Context, modelruntime.OperationEvent) error) *Writer {
	if bufferSize <= 0 {
		bufferSize = 8
	}
	return &Writer{
		syncMode: syncMode,
		insert:   insert,
		ch:       make(chan modelruntime.OperationEvent, bufferSize),
		stop:     make(chan struct{}),
	}
}

// Start 启动异步落库协程（syncMode=true 时为空操作）。
func (w *Writer) Start() {
	if w.syncMode {
		return
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-w.stop:
				return
			case evt := <-w.ch:
				// 异步模式下由后台协程落库，避免操作链路被审计 IO 拖慢。
				_ = w.insert(context.Background(), evt)
			}
		}
	}()
}

// Write 写入一条操作审计事件。
func (w *Writer) Write(ctx context.Context, evt modelruntime.OperationEvent) {
	if w.syncMode {
		_ = w.insert(ctx, evt)
		return
	}
	select {
	case w.ch <- evt:
	default:
		// 缓冲打满时直接同步写入，优先保证日志不丢。
		_ = w.insert(ctx, evt)
	}
}

// Close 停止异步写入并等待后台协程退出。
func (w *Writer) Close() {
	close(w.stop)
	w.wg.Wait()
}
