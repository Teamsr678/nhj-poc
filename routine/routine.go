package routine

import (
	"context"
	"log"
	"nhj-poc/service"
	"time"

	"github.com/go-co-op/gocron"
)

func StartUpdatePaymentStatusJob(ctx context.Context) (*gocron.Scheduler, error) {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return nil, err
	}

	s := gocron.NewScheduler(loc)

	_, err = s.Every(1).Day().At("15:54").Do(func() {
		log.Println("🔄 Daily UpdatePaymentStatus job starting")
		service.UpdatePaymentStatusByIDs([]int{})
		log.Println("✅ Daily UpdatePaymentStatus job finished")
	})
	if err != nil {
		return nil, err
	}

	s.StartAsync()

	return s, nil
}
