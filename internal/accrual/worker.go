package accrual

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nikitaw13/gophermart/internal/model"
	"go.uber.org/zap"
)

var backoffSchedule = []time.Duration{
	100 * time.Millisecond,
	200 * time.Millisecond,
	400 * time.Millisecond,
}

// AccrualWorker periodically polls the accrual system for order status updates.
type AccrualWorker struct {
	client      *AccrualClient
	repository  AccrualRepository
	logger      *zap.Logger
	interval    time.Duration
	workerCount int
	pauseUntil  atomic.Int64
	done        chan struct{}
}

// NewAccrualWorker creates a worker that polls the accrual system at the given interval.
func NewAccrualWorker(client *AccrualClient, repository AccrualRepository, logger *zap.Logger, workerCount int, interval time.Duration) *AccrualWorker {
	if workerCount < 1 {
		workerCount = 1
	}
	return &AccrualWorker{
		client:      client,
		repository:  repository,
		logger:      logger,
		interval:    interval,
		workerCount: workerCount,
		done:        make(chan struct{}),
	}
}

// Done returns a channel that is closed once ProcessNewOrders has fully stopped.
func (aw *AccrualWorker) Done() <-chan struct{} {
	return aw.done
}

// ProcessNewOrders periodically fetches new orders from the repository
// and updates their accrual status until the context is cancelled.
func (aw *AccrualWorker) ProcessNewOrders(ctx context.Context) error {
	defer close(aw.done)
	var wg sync.WaitGroup
	defer wg.Wait()

	orderNumberStream := make(chan string, aw.workerCount)
	defer close(orderNumberStream)

	for i := 0; i < aw.workerCount; i++ {
		wg.Go(func() {
			aw.consumeJobs(ctx, orderNumberStream)
		})

	}

	ticker := time.NewTicker(aw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			orders, err := aw.repository.FetchNewOrders(ctx)
			if err != nil {
				aw.logger.Error("error fetching new orders", zap.Error(err))
				continue
			}
			for _, orderNumber := range orders {
				select {
				case orderNumberStream <- orderNumber:
				case <-ctx.Done():
					return nil
				}
			}
		}
	}
}

func (aw *AccrualWorker) consumeJobs(ctx context.Context, orderNumberStream <-chan string) {
	for orderNumber := range orderNumberStream {
		err := aw.processOrder(ctx, orderNumber)
		if errors.Is(err, ErrOrderNotRegistered) {
			aw.logger.Debug("order is not yet registered in the accrual system", zap.Error(err), zap.String("order", orderNumber))
		} else {
			aw.logger.Error("error processing an order", zap.Error(err), zap.String("order", orderNumber))
		}

	}
}

func (aw *AccrualWorker) processOrder(ctx context.Context, orderNumber string) error {
	order, err := aw.getStatusWithRetries(ctx, orderNumber)
	if err != nil {
		return err
	}

	if order.Status == model.StatusRegistered {
		return nil
	}

	err = aw.repository.UpdateOrder(ctx, orderNumber, order.Status, order.Accrual)
	if err != nil {
		return err
	}
	return nil
}

func (aw *AccrualWorker) getStatusWithRetries(ctx context.Context, orderNumber string) (Order, error) {
	var (
		lastError error
		order     Order
		err       error
	)

	for _, delay := range backoffSchedule {
		for {
			remaining := aw.pauseUntil.Load() - time.Now().Unix()
			if remaining <= 0 {
				break
			}

			select {
			case <-ctx.Done():
				return order, ctx.Err()
			case <-time.After(time.Duration(remaining) * time.Second):
			}
		}

		order, err = aw.client.GetStatus(ctx, orderNumber)

		if err == nil {
			return order, nil
		}

		if Classify(err) == NonRetriable {
			return order, err
		}

		// Use the backoff delay unless the server requested a longer Retry-After wait.
		wait := delay

		if ra, ok := errors.AsType[*RetryAfterError](err); ok {
			// Extend the pause deadline if the new one is later.
			for {
				oldPauseUntil := aw.pauseUntil.Load()
				newPauseUntil := time.Now().Unix() + int64(ra.After)

				if newPauseUntil <= oldPauseUntil {
					break
				}
				if aw.pauseUntil.CompareAndSwap(oldPauseUntil, newPauseUntil) {
					break
				}
			}

			wait = time.Duration(ra.After) * time.Second
		}

		lastError = err

		select {
		case <-ctx.Done():
			return order, ctx.Err()
		case <-time.After(wait):
		}

	}

	return Order{}, fmt.Errorf("all attempts unsuccessful: %w", lastError)
}
