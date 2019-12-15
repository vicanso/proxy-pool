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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyList(t *testing.T) {
	assert := assert.New(t)
	pl := new(ProxyList)

	p := &Proxy{
		IP:       "127.0.0.1",
		Port:     "80",
		Category: "http",
		Speed:    2,
	}
	pl.Add(p)
	pl.Add(p)

	result := pl.FindOne("http", 2)
	assert.Equal(p, result)

	assert.False(pl.Exists(&Proxy{
		IP:       "127.0.0.1",
		Port:     "80",
		Category: "https",
	}))
	assert.True(pl.Exists(&Proxy{
		IP:       "127.0.0.1",
		Port:     "80",
		Category: "http",
	}))
	assert.Equal(1, pl.Size())
	assert.Equal(p, pl.List()[0])
	assert.Equal(p, pl.Reset()[0])
	assert.Equal(0, pl.Size())

	pl.Add(p)
	newList := make([]*Proxy, 0)
	pl.Replace(newList)
	assert.Equal(newList, pl.data)
}
