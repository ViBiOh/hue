package worker

import (
	"context"
	"time"

	"github.com/ViBiOh/httputils/pkg/logger"
)

const (
	notificationInterval = time.Hour * 24
	retryInterval        = time.Minute * 10
)

func (a *App) getTimer(hour int, minute int, interval time.Duration) *time.Timer {
	currentTime := time.Now()

	nextTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), hour, minute, 0, 0, a.location)
	if !nextTime.After(currentTime.In(a.location)) {
		nextTime = nextTime.Add(interval)
	}

	logger.Info("Next scrapping at %v", nextTime)

	return time.NewTimer(time.Until(nextTime))
}

func (a *App) scheduler() {
	timer := a.getTimer(a.hour, a.minute, notificationInterval)

	for {
		select {
		case <-timer.C:
			ctx := context.Background()

			if err := a.fetchAndSaveData(ctx); err != nil {
				logger.Error(`%+v`, err)

				timer.Reset(retryInterval)
				logger.Warn("Retrying in 10 minutes")
			} else {
				return
			}
		}
	}
}

func (a *App) startScheduler() {
	if a.location == nil {
		logger.Warn("location is missing for sending notification")
		return
	}

	if a.db == nil {
		logger.Warn("storage is missing for saving enedis consumption")
		return
	}

	for {
		a.scheduler()
	}
}
