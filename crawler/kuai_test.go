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

	"github.com/vicanso/go-axios"

	"github.com/stretchr/testify/assert"
)

func TestKuaiProxy(t *testing.T) {
	assert := assert.New(t)
	kuai := NewKuaiProxy(time.Minute)
	html := `<html>
		<body>
			<div id="listnav">
				<a>10</a>
			</div>
			<div id="list">
				<table>
					<tbody>
					<tr>
						<td data-title="IP">171.13.103.213</td>
						<td data-title="PORT">9999</td>
						<td data-title="匿名度">高匿名</td>
						<td data-title="类型">HTTP</td>
						<td data-title="位置">河南省鹤壁市  电信</td>
						<td data-title="响应速度">2秒</td>
						<td data-title="最后验证时间">2019-12-14 15:31:01</td>
					</tr>
					</tbody>
				</table>
			</div>
		</body>
	</html>`
	kuai.ins.Mock(&axios.Response{
		Status: 200,
		Data:   []byte(html),
	})
	done := make(chan bool)
	kuai.OnFetch(func(p *Proxy) {
		assert.Equal("171.13.103.213", p.IP)
		assert.Equal("9999", p.Port)
		assert.Equal("http", p.Category)
		done <- true
	})
	go kuai.Start()
	<-done
	kuai.Stop()
	assert.Equal(10, kuai.maxPage)
}
