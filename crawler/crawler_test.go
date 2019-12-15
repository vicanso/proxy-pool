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

	"github.com/vicanso/go-axios"

	"github.com/stretchr/testify/assert"
)

func TestBaseProxyCrawler(t *testing.T) {
	assert := assert.New(t)
	bp := new(baseProxyCrawler)

	assert.Nil(bp.fetchListener)
	bp.OnFetch(func(_ *Proxy) {})
	assert.NotNil(bp.fetchListener)

	ins := axios.NewInstance(nil)
	bp.currentPage = 10
	bp.maxPage = 10
	bp.ins = ins
	// empty data
	done := ins.Mock(&axios.Response{
		Status: 200,
		Data:   []byte(""),
	})
	doc, err := bp.fetchPage("", "%d")
	assert.Nil(err)
	assert.Nil(doc)
	assert.Equal(0, bp.maxPage)
	assert.Equal(1, bp.currentPage)
	done()
}

