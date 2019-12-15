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

func TestXiciProxy(t *testing.T) {
	assert := assert.New(t)
	xici := NewXiciProxy(time.Minute)
	html := `<html>
		<body>
			<div class="pagination">
				<a>10</a>
				<a></a>
			</div>
			<table id="ip_list">
				<tbody>
				<tr></tr>
				<tr class="odd">
					<td class="country"><img src="//fs.xicidaili.com/images/flag/cn.png" alt="Cn"></td>
					<td>183.154.49.8</td>
					<td>9999</td>
					<td>
					<a href="/2019-12-14/zhejiang">浙江金华</a>
					</td>
					<td class="country">高匿</td>
					<td>HTTP</td>
					<td class="country">
					<div title="6.259秒" class="bar">
					<div class="bar_inner slow" style="width:68%">

					</div>
					</div>
					</td>
					<td class="country">
					<div title="1.251秒" class="bar">
					<div class="bar_inner medium" style="width:80%">

					</div>
					</div>
					</td>

					<td>1分钟</td>
					<td>19-12-14 16:21</td>
				</tr>
				</tbody>
			</table>
		</body>
	</html>`
	xici.ins.Mock(&axios.Response{
		Status: 200,
		Data:   []byte(html),
	})
	done := make(chan bool)
	xici.OnFetch(func(p *Proxy) {
		assert.Equal("183.154.49.8", p.IP)
		assert.Equal("9999", p.Port)
		assert.Equal("http", p.Category)
		done <- true
	})
	go xici.Start()
	<-done
	xici.Stop()
	assert.Equal(10, xici.maxPage)
}
