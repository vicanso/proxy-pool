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

package crawler

import (
	"math/rand"
	"sync"
	"time"
)

type (
	// Proxy proxy server
	Proxy struct {
		DetectedAt int64
		IP         string
		Port       string
		Category   string
		Anonymous  bool
		Speed      int32
		Fails      int32
	}
	// ProxyList proxy list
	ProxyList struct {
		sync.RWMutex
		data []*Proxy
	}
)

func (pl *ProxyList) indexOf(p *Proxy) int {
	index := -1
	for i, item := range pl.data {
		if item.IP == p.IP &&
			item.Port == p.Port &&
			item.Category == p.Category {
			index = i
			break
		}
	}
	return index
}

// Exists test whether or not the proxy exists
func (pl *ProxyList) Exists(p *Proxy) bool {
	pl.RLock()
	defer pl.RUnlock()
	return pl.indexOf(p) != -1
}

// Add add proxy to list
func (pl *ProxyList) Add(list ...*Proxy) {
	if len(list) == 0 {
		return
	}
	pl.Lock()
	defer pl.Unlock()
	if len(pl.data) == 0 {
		pl.data = make([]*Proxy, 0, 100)
	}
	for _, p := range list {
		if pl.indexOf(p) != -1 {
			continue
		}
		pl.data = append(pl.data, p)
	}
}

// Remove remove proxy from list
func (pl *ProxyList) Remove(list ...*Proxy) {
	if len(list) == 0 {
		return
	}
	pl.Lock()
	defer pl.Unlock()
	if len(pl.data) == 0 {
		return
	}
	for _, p := range list {
		index := pl.indexOf(p)
		if index != -1 {
			pl.data = append(pl.data[:index], pl.data[index+1:]...)
		}
	}
}

// List get proxy list
func (pl *ProxyList) List() []*Proxy {
	pl.RLock()
	defer pl.RUnlock()
	return pl.data[:]
}

// Reset reset proxy list
func (pl *ProxyList) Reset() []*Proxy {
	pl.Lock()
	defer pl.Unlock()
	oldProxyList := pl.data[:]
	pl.data = nil
	return oldProxyList
}

// Size get the size of proxy list
func (pl *ProxyList) Size() int {
	pl.RLock()
	defer pl.RUnlock()
	return len(pl.data)
}

// Replace replace the proxy list
func (pl *ProxyList) Replace(list []*Proxy) {
	pl.Lock()
	defer pl.Unlock()
	pl.data = list
}

// FindOne find one proxy
func (pl *ProxyList) FindOne(category string, speed int32) (p *Proxy) {
	pl.RLock()
	defer pl.RUnlock()
	list := pl.data
	// 指定了速度或者代理类型
	if speed >= 0 || category != "" {
		list = make([]*Proxy, 0, 10)
		for _, item := range pl.data {
			if speed >= 0 && item.Speed != speed {
				continue
			}
			if category != "" && item.Category != category {
				continue
			}
			list = append(list, item)
		}
	}
	size := len(list)
	if size == 0 {
		return
	}
	rand.Seed(time.Now().UnixNano())
	return list[rand.Intn(size)]
}
