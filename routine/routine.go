package routine

import (
	"context"
	"log"
	"nhj-poc/services"
	"time"

	"github.com/go-co-op/gocron"
)

func StartUpdatePaymentStatusJob(ctx context.Context) (*gocron.Scheduler, error) {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return nil, err
	}

	s := gocron.NewScheduler(loc)

	_, err = s.Every(1).Day().At("14:49").Do(func() {
		log.Println("🔄 Daily UpdatePaymentStatus job starting")
		services.UpdatePaymentStatusByIDs([]int{})
		log.Println("✅ Daily UpdatePaymentStatus job finished")
	})
	if err != nil {
		return nil, err
	}

	s.StartAsync()

	return s, nil
}
