package async

import (
	"go.uber.org/zap"
	"sync"
	"time"
)

type Task func()

type WorkerPool struct {
	tasks        chan Task
	workersCount int
	logger       *zap.Logger
	wg           sync.WaitGroup
	shutdown     chan struct{}
	isShutDown   bool
	mu           sync.Mutex
}

func NewWorkerPool(workersCount int, queueSize int, logger *zap.Logger) *WorkerPool {
	pool := &WorkerPool{
		tasks:        make(chan Task, queueSize),
		workersCount: workersCount,
		logger:       logger.With(zap.String("component", "worker_pool")),
		shutdown:     make(chan struct{}),
	}
	pool.start()
	return pool
}

func (p *WorkerPool) start() {
	p.logger.Info("Starting worker pool", zap.Int("workers", p.workersCount))

	for i := 0; i < p.workersCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	p.logger.Debug("Worker started", zap.Int("worker_id", id))

	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				p.logger.Debug("Worker shutting down (channel closed)", zap.Int("worker_id", id))
				return
			}
			startTime := time.Now()

			func() {
				defer func() {
					if r := recover(); r != nil {
						p.logger.Error("Task panicked", zap.Any("panic", r), zap.Int("worker_id", id))
					}
				}()
				task()
			}()
			p.logger.Debug("Task completed",
				zap.Int("worker_id", id),
				zap.Duration("duration", time.Since(startTime)))
		case <-p.shutdown:
			p.logger.Debug("Worker shutting down (shutdown signal)", zap.Int("worker_id", id))
			return
		}
	}
}

func (p *WorkerPool) Submit(task Task) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isShutDown {
		return false
	}

	select {
	case p.tasks <- task:
		return true
	default:
		p.logger.Warn("The queue is full")
		return false
	}
}

func (p *WorkerPool) Shutdown(wait bool) {
	p.mu.Lock()
	if p.isShutDown {
		p.mu.Unlock()
		return
	}
	p.isShutDown = true
	p.mu.Unlock()
	p.logger.Info("Shutting down worker pool")

	close(p.tasks)

	close(p.shutdown)

	if wait {
		p.logger.Debug("Waiting for all workers to finish")
		p.wg.Wait()
		p.logger.Info("Worker pool shutdown complete")
	}

}
