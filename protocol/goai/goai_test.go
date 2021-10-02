// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package goai_test

import (
	"context"
	"github.com/gogf/gf/protocol/goai"
	"github.com/gogf/gf/test/gtest"
	"github.com/gogf/gf/util/gmeta"
	"testing"
)

func Test_Basic(t *testing.T) {
	type CommonReq struct {
		AppId      int64  `json:"appId" v:"required" in:"cookie" description:"应用Id"`
		ResourceId string `json:"resourceId" in:"query" description:"资源Id"`
	}
	type SetSpecInfo struct {
		StorageType string   `v:"required|in:CLOUD_PREMIUM,CLOUD_SSD,CLOUD_HSSD" description:"StorageType"`
		Shards      int32    `description:"shards 分片数"`
		Params      []string `description:"默认参数(json 串-ClickHouseParams)"`
	}
	type CreateResourceReq struct {
		CommonReq
		gmeta.Meta `path:"/CreateResourceReq" method:"POST" tags:"default"`
		Name       string                  `description:"实例名称"`
		Product    string                  `description:"业务类型"`
		Region     string                  `v:"required" description:"区域"`
		SetMap     map[string]*SetSpecInfo `v:"required" description:"配置Map"`
		SetSlice   []SetSpecInfo           `v:"required" description:"配置Slice"`
	}

	gtest.C(t, func(t *gtest.T) {
		var (
			err error
			oai = goai.New()
			req = new(CreateResourceReq)
		)
		err = oai.Add(goai.AddInput{
			Object: req,
		})
		t.AssertNil(err)
		// Schema asserts.
		t.Assert(len(oai.Components.Schemas), 2)
		t.Assert(oai.Components.Schemas[`CreateResourceReq`].Value.Type, goai.TypeObject)
		t.Assert(len(oai.Components.Schemas[`CreateResourceReq`].Value.Properties), 7)
		t.Assert(oai.Components.Schemas[`CreateResourceReq`].Value.Properties[`appId`].Value.Type, goai.TypeNumber)
		t.Assert(oai.Components.Schemas[`CreateResourceReq`].Value.Properties[`resourceId`].Value.Type, goai.TypeString)

		t.Assert(len(oai.Components.Schemas[`SetSpecInfo`].Value.Properties), 3)
		t.Assert(oai.Components.Schemas[`SetSpecInfo`].Value.Properties[`Params`].Value.Type, goai.TypeArray)
	})
}

func TestOpenApiV3_Add(t *testing.T) {
	type CommonReq struct {
		AppId      int64  `json:"appId" v:"required" in:"path" description:"应用Id"`
		ResourceId string `json:"resourceId" in:"query" description:"资源Id"`
	}
	type SetSpecInfo struct {
		StorageType string   `v:"required|in:CLOUD_PREMIUM,CLOUD_SSD,CLOUD_HSSD" description:"StorageType"`
		Shards      int32    `description:"shards 分片数"`
		Params      []string `description:"默认参数(json 串-ClickHouseParams)"`
	}
	type CreateResourceReq struct {
		CommonReq
		gmeta.Meta `path:"/CreateResourceReq" method:"POST" tags:"default"`
		Name       string                  `description:"实例名称"`
		Product    string                  `description:"业务类型"`
		Region     string                  `v:"required" description:"区域"`
		SetMap     map[string]*SetSpecInfo `v:"required" description:"配置Map"`
		SetSlice   []SetSpecInfo           `v:"required" description:"配置Slice"`
	}

	type CreateResourceRes struct {
		gmeta.Meta `description:"Demo Response Struct"`
		FlowId     int64 `description:"创建实例流程id"`
	}

	f := func(ctx context.Context, req *CreateResourceReq) (res *CreateResourceRes, err error) {
		return
	}

	gtest.C(t, func(t *gtest.T) {
		var (
			err error
			oai = goai.New()
		)
		err = oai.Add(goai.AddInput{
			Path:   "/test1/{appId}",
			Method: goai.HttpMethodPut,
			Object: f,
		})
		t.AssertNil(err)

		err = oai.Add(goai.AddInput{
			Path:   "/test2/{appId}",
			Method: goai.HttpMethodPost,
			Object: f,
		})
		t.AssertNil(err)
		// Schema asserts.
		t.Assert(len(oai.Components.Schemas), 3)
		t.Assert(oai.Components.Schemas[`CreateResourceReq`].Value.Type, goai.TypeObject)
		t.Assert(len(oai.Components.Schemas[`CreateResourceReq`].Value.Properties), 7)
		t.Assert(oai.Components.Schemas[`CreateResourceReq`].Value.Properties[`appId`].Value.Type, goai.TypeNumber)
		t.Assert(oai.Components.Schemas[`CreateResourceReq`].Value.Properties[`resourceId`].Value.Type, goai.TypeString)

		t.Assert(len(oai.Components.Schemas[`SetSpecInfo`].Value.Properties), 3)
		t.Assert(oai.Components.Schemas[`SetSpecInfo`].Value.Properties[`Params`].Value.Type, goai.TypeArray)

		// Paths.
		t.Assert(len(oai.Paths), 2)
		t.AssertNE(oai.Paths[`/test1/{appId}`].Put, nil)
		t.Assert(len(oai.Paths[`/test1/{appId}`].Put.Tags), 1)
		t.Assert(len(oai.Paths[`/test1/{appId}`].Put.Parameters), 2)
		t.AssertNE(oai.Paths[`/test2/{appId}`].Post, nil)
		t.Assert(len(oai.Paths[`/test2/{appId}`].Post.Tags), 1)
		t.Assert(len(oai.Paths[`/test2/{appId}`].Post.Parameters), 2)
	})
}
