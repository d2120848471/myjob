package audit

import (
	"context"
	"sync"

	modelruntime "myjob/internal/model/runtime"
)

type Writer struct {
	syncMode bool
	insert   func(context.Context, modelruntime.OperationEvent) error
	ch       chan modelruntime.OperationEvent
	stop     chan struct{}
	wg       sync.WaitGroup
}

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

func (w *Writer) Close() {
	close(w.stop)
	w.wg.Wait()
}
