package worker

import (
	"context"
	"log"
	"time"

	"order_processing/internal/service"
)

type StatusUpdater struct {
	service  *service.OrderService
	interval time.Duration
}

func NewStatusUpdater(svc *service.OrderService, interval time.Duration) *StatusUpdater {
	return &StatusUpdater{
		service:  svc,
		interval: interval,
	}
}

func (w *StatusUpdater) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("status updater started (interval=%s)", w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("status updater stopped")
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *StatusUpdater) runOnce(ctx context.Context) {
	count, err := w.service.ProcessPendingOrders(ctx)
	if err != nil {
		log.Printf("status updater error: %v", err)
		return
	}
	if count > 0 {
		log.Printf("status updater moved %d order(s) from PENDING to PROCESSING", count)
	}
}
