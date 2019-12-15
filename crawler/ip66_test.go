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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/go-axios"
)

func TestIP66Proxy(t *testing.T) {
	assert := assert.New(t)
	ip66 := NewIP66Proxy(time.Minute)
	html := `<html>
		<body>
			<div id="PageList">
				<a>10</a>
				<a></a>
			</div>
			<div id="main">
				<table>
					<tr></tr>
					<tr><td>183.166.71.93</td><td>9999</td><td>安徽省淮南市</td><td>高匿代理</td><td>2019年12月14日14时 验证</td></tr>
				</table>
			</div>
		</body>
	</html>`
	ip66.ins.Mock(&axios.Response{
		Status: 200,
		Data:   []byte(html),
	})
	done := make(chan bool)
	ip66.OnFetch(func(p *Proxy) {
		assert.Equal("183.166.71.93", p.IP)
		assert.Equal("9999", p.Port)
		done <- true
	})
	go ip66.Start()
	<-done
	ip66.Stop()
	assert.Equal(10, ip66.maxPage)
}
