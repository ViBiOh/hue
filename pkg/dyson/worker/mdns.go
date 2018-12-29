package worker

import (
	"context"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/grandcat/zeroconf"
)

func findDysonMQTTServices() (map[string]*zeroconf.ServiceEntry, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	services := make(map[string]*zeroconf.ServiceEntry)
	entries := make(chan *zeroconf.ServiceEntry)

	go func() {
		for service := range entries {
			services[service.Instance] = service
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	err = resolver.Browse(ctx, `_dyson_mqtt._tcp`, ``, entries)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	<-ctx.Done()
	return services, nil
}
