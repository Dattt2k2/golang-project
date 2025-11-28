package cron

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron       *cron.Cron
	aggregator *PendingReviewAggregator
}

func NewScheduler(aggregator *PendingReviewAggregator) *Scheduler {
	c := cron.New(cron.WithSeconds())
	return &Scheduler{
		cron:       c,
		aggregator: aggregator,
	}
}

func (s *Scheduler) Start() {
	// Chạy mỗi 30 phút
	s.cron.AddFunc("0 */1 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		if err := s.aggregator.Run(ctx); err != nil {
			log.Printf("Pending review aggregation failed: %v", err)
		}
	})

	// Hoặc mỗi 1 giờ: "0 0 * * * *"
	// Hoặc mỗi ngày 2h sáng: "0 0 2 * * *"
	// Hoặc mỗi 15 phút: "0 */15 * * * *"

	s.cron.Start()
	log.Println("Review aggregation scheduler started - Running every 30 minutes")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("Review aggregation scheduler stopped")
}
