package orderlogic

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Worker 负责后台提交待处理开放订单，并轮询已提交订单的上游状态。
type Worker struct {
	logic    *OrderLogic
	stop     chan struct{}
	done     chan struct{}
	submitCh chan string

	startOnce sync.Once
	stopOnce  sync.Once
	started   atomic.Bool
}

// NewWorker 创建订单后台 worker，并挂接到对应的 OrderLogic。
func NewWorker(logic *OrderLogic) *Worker {
	worker := &Worker{
		logic:    logic,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		submitCh: make(chan string, 64),
	}
	logic.worker = worker
	return worker
}

// Start 启动订单后台 worker。
func (w *Worker) Start() {
	w.startOnce.Do(func() {
		w.started.Store(true)
		go w.loop()
	})
}

// Stop 停止订单后台 worker，并等待当前循环退出。
func (w *Worker) Stop() {
	if !w.started.Load() {
		return
	}
	w.stopOnce.Do(func() {
		close(w.stop)
	})
	<-w.done
}

// Trigger 非阻塞地通知 worker 尽快提交订单；队列满时交给定时扫描兜底。
func (w *Worker) Trigger(orderNo string) {
	select {
	case <-w.stop:
		return
	case w.submitCh <- orderNo:
	default:
	}
}

func (w *Worker) loop() {
	defer close(w.done)

	cfg := w.logic.core.Config().OpenOrder
	submitTicker := time.NewTicker(time.Duration(cfg.SubmitScanIntervalSeconds) * time.Second)
	defer submitTicker.Stop()
	pollTicker := time.NewTicker(time.Duration(cfg.PollIntervalSeconds) * time.Second)
	defer pollTicker.Stop()

	for {
		select {
		case <-w.stop:
			return
		case <-w.submitCh:
			// Trigger 只作为唤醒信号；实际提交仍扫描待提交订单，避免遗漏同一时间进来的多笔订单。
			_ = w.logic.SubmitPendingOnce(context.Background())
		case <-submitTicker.C:
			_ = w.logic.SubmitPendingOnce(context.Background())
		case <-pollTicker.C:
			_ = w.logic.PollDueOnce(context.Background())
		}
	}
}
