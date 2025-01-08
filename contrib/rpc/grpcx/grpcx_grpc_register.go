package grpcx

import (
	"context"
)

func (s *GrpcServer) WithMetadata(ctx context.Context, metadata map[string]interface{}) {
	if s.registrar == nil {
		return
	}
	s.serviceMu.Lock()
	defer s.serviceMu.Unlock()
	if len(s.services) == 0 {
		return
	}
	var (
		err error
	)
	// Register service list after server starts.
	for i, service := range s.services {
		service.SetMetadata(metadata)
		s.Logger().Debugf(ctx, `service register: %+v`, service)
		if len(service.GetEndpoints()) == 0 {
			s.Logger().Warningf(ctx, `no endpoints found to register service, abort service registering`)
			return
		}
		if service, err = s.registrar.Register(ctx, service); err != nil {
			s.Logger().Fatalf(ctx, `%+v`, err)
		}
		s.services[i] = service
	}
}
