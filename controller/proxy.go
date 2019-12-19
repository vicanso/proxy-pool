// Copyright 2019 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"strconv"

	"github.com/vicanso/elton"
	"github.com/vicanso/proxy-pool/router"
	"github.com/vicanso/proxy-pool/service"
)

type (
	proxyCtrl struct{}
)

func init() {
	ctrl := proxyCtrl{}
	g := router.NewGroup("/proxies")

	g.GET("", ctrl.list)
	g.GET("/one", ctrl.findOne)
}

// list get all available proxy
func (proxyCtrl) list(c *elton.Context) (err error) {
	c.CacheMaxAge("1m")
	// 直接返回所有可用的proxy，暂不考虑分页等处理
	c.Body = map[string]interface{}{
		"proxies": service.GetAvailableProxyList(),
	}
	return
}

// findOne get one available proxy
func (proxyCtrl) findOne(c *elton.Context) (err error) {
	category := c.QueryParam("category")
	speed := -1
	sp := c.QueryParam("speed")
	if sp != "" {
		v, e := strconv.Atoi(sp)
		if e == nil {
			speed = v
		}
	}
	p := service.GetAvailableProxy(category, speed)
	if p == nil {
		c.NoContent()
		return
	}
	c.Body = p
	return
}
