// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package etcd

import (
	"context"
	"github.com/gogf/gf/v2/os/grpool"
	etcd3 "go.etcd.io/etcd/client/v3"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/gsvc"
)

// Register registers `service` to Registry.
// Note that it returns a new Service if it changes the input Service with custom one.
func (r *Registry) Register(ctx context.Context, service gsvc.Service) (gsvc.Service, error) {
	service = NewService(service)
	if r.leaseId != 0 {
		err := r.saveService(ctx, service, r.leaseId)
		return service, err
	}
	r.lease = etcd3.NewLease(r.client)
	grant, err := r.lease.Grant(ctx, int64(r.keepaliveTTL.Seconds()))
	if err != nil {
		return nil, gerror.Wrapf(err, `etcd grant failed with keepalive ttl "%s"`, r.keepaliveTTL)
	}
	err = r.saveService(ctx, service, grant.ID)
	if err != nil {
		return nil, err
	}
	keepAliceCh, err := r.client.KeepAlive(context.Background(), grant.ID)
	if err != nil {
		return nil, err
	}
	_ = grpool.AddWithRecover(ctx, func(ctx context.Context) {
		r.doKeepAlive(grant.ID, keepAliceCh)
	}, nil)
	r.leaseId = grant.ID
	return service, nil
}

func (r *Registry) saveService(ctx context.Context, service gsvc.Service, leaseId etcd3.LeaseID) error {
	var (
		key   = service.GetKey()
		value = service.GetValue()
	)
	_, err := r.client.Put(ctx, key, value, etcd3.WithLease(leaseId))
	if err != nil {
		return gerror.Wrapf(
			err,
			`etcd put failed with key "%s", value "%s", lease "%d"`,
			key, value, leaseId,
		)
	}
	r.logger.Debugf(
		ctx,
		`etcd put success with key "%s", value "%s", lease "%d"`,
		key, value, leaseId,
	)
	r.service = service
	r.registryFlag = true
	return nil
}

// Deregister off-lines and removes `service` from the Registry.
func (r *Registry) Deregister(ctx context.Context, service gsvc.Service) error {
	_, err := r.client.Delete(ctx, service.GetKey())
	if r.lease != nil {
		_ = r.lease.Close()
	}
	return err
}

// doKeepAlive continuously keeps alive the lease from ETCD.
func (r *Registry) doKeepAlive(leaseID etcd3.LeaseID, keepAliceCh <-chan *etcd3.LeaseKeepAliveResponse) {
	var ctx = context.Background()
	for {
		select {
		case <-r.client.Ctx().Done():
			r.logger.Noticef(ctx, "keepalive done for lease id: %d", leaseID)
			return

		case _, ok := <-keepAliceCh:
			if !ok {
				r.logger.Noticef(ctx, `keepalive exit, lease id: %d`, leaseID)
				// 通道关闭，重新注册
				if r.service != nil {
					for i := 0; i < 10; i++ {
						r.leaseId = 0
						_, err := r.Register(ctx, r.service)
						if err != nil {
							time.Sleep(time.Second)
							continue
						}
						return
					}
					return
				}

			}
		}
	}
}
