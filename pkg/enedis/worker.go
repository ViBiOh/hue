package enedis

import (
	"encoding/json"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/iot/pkg/provider"
)

func (a *App) handleConsumptionWorker(message *provider.WorkerMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var data Consumption
	if err := json.Unmarshal([]byte(message.Payload), &data); err != nil {
		return errors.WithStack(err)
	}

	a.consumption = &data

	if a.prometheus {
		a.updatePrometheus()
	}

	return nil
}

// WorkerHandler handle commands receive from worker
func (a *App) WorkerHandler(p *provider.WorkerMessage) error {
	if p.Action == ConsumptionAction {
		return a.handleConsumptionWorker(p)
	}

	return provider.ErrWorkerUnknownAction
}
