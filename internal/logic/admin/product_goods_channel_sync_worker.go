package adminlogic

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// ProductGoodsChannelSyncWorkerOptions 控制商品渠道同步后台任务的运行参数。
type ProductGoodsChannelSyncWorkerOptions struct {
	Interval time.Duration
	Limit    int
}

// ProductGoodsChannelSyncWorker 周期性同步开启开关的商品渠道绑定。
type ProductGoodsChannelSyncWorker struct {
	logic    *ProductGoodsLogic
	interval time.Duration
	limit    int
	ctx      context.Context
	cancel   context.CancelFunc
	stop     chan struct{}
	done     chan struct{}

	startOnce sync.Once
	stopOnce  sync.Once
	started   atomic.Bool
	running   atomic.Bool
}

// NewProductGoodsChannelSyncWorker 创建商品渠道同步后台任务。
func NewProductGoodsChannelSyncWorker(logic *ProductGoodsLogic, opts ProductGoodsChannelSyncWorkerOptions) *ProductGoodsChannelSyncWorker {
	interval := opts.Interval
	if interval <= 0 {
		interval = time.Minute
	}
	limit := opts.Limit
	if limit <= 0 || limit > defaultProductGoodsChannelSyncLimit {
		limit = defaultProductGoodsChannelSyncLimit
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &ProductGoodsChannelSyncWorker{
		logic:    logic,
		interval: interval,
		limit:    limit,
		ctx:      ctx,
		cancel:   cancel,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Start 启动商品渠道同步后台任务。
func (w *ProductGoodsChannelSyncWorker) Start() {
	w.startOnce.Do(func() {
		w.started.Store(true)
		go w.loop()
	})
}

// Stop 停止商品渠道同步后台任务，并等待后台循环退出。
func (w *ProductGoodsChannelSyncWorker) Stop() {
	if !w.started.Load() {
		return
	}
	w.stopOnce.Do(func() {
		w.cancel()
		close(w.stop)
	})
	<-w.done
}

func (w *ProductGoodsChannelSyncWorker) loop() {
	defer close(w.done)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.stop:
			return
		case <-ticker.C:
			w.runOnce()
		}
	}
}

func (w *ProductGoodsChannelSyncWorker) runOnce() {
	if w.logic == nil || !w.tryBeginRun() {
		return
	}
	defer w.finishRun()
	_, _ = w.logic.SyncChannelBindingsOnce(w.ctx, ProductGoodsChannelSyncOptions{Limit: w.limit})
}

func (w *ProductGoodsChannelSyncWorker) tryBeginRun() bool {
	return w.running.CompareAndSwap(false, true)
}

func (w *ProductGoodsChannelSyncWorker) finishRun() {
	w.running.Store(false)
}
