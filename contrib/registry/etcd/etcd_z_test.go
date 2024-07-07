// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package etcd_test

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"

	"github.com/gogf/gf/contrib/registry/etcd/v2"
	"github.com/gogf/gf/v2/net/gsvc"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/test/gtest"
	"github.com/gogf/gf/v2/util/guid"
)

func TestRegistry(t *testing.T) {
	var (
		ctx      = gctx.GetInitCtx()
		registry = etcd.New(`127.0.0.1:2379@root:123`)
	)
	svc := &gsvc.LocalService{
		Name:      guid.S(),
		Endpoints: gsvc.NewEndpoints("127.0.0.1:8888"),
		Metadata: map[string]interface{}{
			"protocol": "https",
		},
	}
	gtest.C(t, func(t *gtest.T) {
		registered, err := registry.Register(ctx, svc)
		t.AssertNil(err)
		t.Assert(registered.GetName(), svc.GetName())
	})

	// Search by name.
	gtest.C(t, func(t *gtest.T) {
		result, err := registry.Search(ctx, gsvc.SearchInput{
			Name: svc.Name,
		})
		t.AssertNil(err)
		t.Assert(len(result), 1)
		t.Assert(result[0].GetName(), svc.Name)
	})

	// Search by prefix.
	gtest.C(t, func(t *gtest.T) {
		result, err := registry.Search(ctx, gsvc.SearchInput{
			Prefix: svc.GetPrefix(),
		})
		t.AssertNil(err)
		t.Assert(len(result), 1)
		t.Assert(result[0].GetName(), svc.Name)
	})

	// Search by metadata.
	gtest.C(t, func(t *gtest.T) {
		result, err := registry.Search(ctx, gsvc.SearchInput{
			Name: svc.GetName(),
			Metadata: map[string]interface{}{
				"protocol": "https",
			},
		})
		t.AssertNil(err)
		t.Assert(len(result), 1)
		t.Assert(result[0].GetName(), svc.Name)
	})
	gtest.C(t, func(t *gtest.T) {
		result, err := registry.Search(ctx, gsvc.SearchInput{
			Name: svc.GetName(),
			Metadata: map[string]interface{}{
				"protocol": "grpc",
			},
		})
		t.AssertNil(err)
		t.Assert(len(result), 0)
	})

	gtest.C(t, func(t *gtest.T) {
		err := registry.Deregister(ctx, svc)
		t.AssertNil(err)
	})
}

func TestWatch(t *testing.T) {
	var (
		ctx      = gctx.GetInitCtx()
		registry = etcd.New(`127.0.0.1:2379@root:123`)
	)

	svc1 := &gsvc.LocalService{
		Name:      guid.S(),
		Endpoints: gsvc.NewEndpoints("127.0.0.1:8888"),
		Metadata: map[string]interface{}{
			"protocol": "https",
		},
	}
	gtest.C(t, func(t *gtest.T) {
		registered, err := registry.Register(ctx, svc1)
		t.AssertNil(err)
		t.Assert(registered.GetName(), svc1.GetName())
	})

	gtest.C(t, func(t *gtest.T) {
		watcher, err := registry.Watch(ctx, svc1.GetPrefix())
		t.AssertNil(err)

		// Register another service.
		svc2 := &gsvc.LocalService{
			Name:      svc1.Name,
			Endpoints: gsvc.NewEndpoints("127.0.0.1:9999"),
		}
		registered, err := registry.Register(ctx, svc2)
		t.AssertNil(err)
		t.Assert(registered.GetName(), svc2.GetName())

		// Watch and retrieve the service changes:
		// svc1 and svc2 is the same service name, which has 2 endpoints.
		proceedResult, err := watcher.Proceed()
		t.AssertNil(err)
		t.Assert(len(proceedResult), 1)
		t.Assert(
			proceedResult[0].GetEndpoints(),
			gsvc.Endpoints{svc1.GetEndpoints()[0], svc2.GetEndpoints()[0]},
		)

		// Watch and retrieve the service changes:
		// left only svc1, which means this service has only 1 endpoint.
		err = registry.Deregister(ctx, svc2)
		t.AssertNil(err)
		proceedResult, err = watcher.Proceed()
		t.AssertNil(err)
		t.Assert(
			proceedResult[0].GetEndpoints(),
			gsvc.Endpoints{svc1.GetEndpoints()[0]},
		)
		t.AssertNil(watcher.Close())
	})

	gtest.C(t, func(t *gtest.T) {
		err := registry.Deregister(ctx, svc1)
		t.AssertNil(err)
	})
}

func TestLease(t *testing.T) {
	// 创建 etcd 客户端
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"10.8.5.21:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("Error creating etcd client:", err)
		return
	}
	defer cli.Close()
	var lease clientv3.Lease
	// 创建租约
	lease = clientv3.NewLease(cli)
	grant, err := lease.Grant(context.TODO(), 10)
	if err != nil {
		fmt.Println("Error creating lease:", err)
		return
	}

	// 将键与租约关联
	key := "foo"
	value := "bar"
	_, err = cli.Put(context.TODO(), key, value, clientv3.WithLease(grant.ID))
	if err != nil {
		fmt.Println("Error putting key with lease:", err)
		return
	}

	// 定期续约租约
	ch, kaerr := cli.KeepAlive(context.TODO(), grant.ID)
	if kaerr != nil {
		fmt.Println("Error setting up keep-alive:", kaerr)
		return
	}

	// 处理续约响应
	go func() {
		for {
			ka := <-ch
			if ka == nil {
				fmt.Println("Lease keep-alive channel closed")
				return
			}
			fmt.Println("Lease keep-alive response received:", ka)
		}
	}()

	// 模拟一些操作
	time.Sleep(30 * time.Second)

	// 更新键值并续约
	newValue := "baz"
	_, err = cli.Put(context.TODO(), key, newValue, clientv3.WithLease(grant.ID))
	if err != nil {
		fmt.Println("Error updating key with lease:", err)
		return
	}
	fmt.Println("Key updated with new value and lease renewed")

	// 保持程序运行以观察续约效果
	select {}
}
