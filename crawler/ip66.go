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
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/vicanso/go-axios"
)

type (
	// ip66Proxy ip66 proxy
	ip66Proxy struct {
		baseProxyCrawler
	}
)

const (
	ProxyIP66 = "ip66"
)

// NewIP66Proxy create a new ip66 proxy crawler
func NewIP66Proxy(interval time.Duration) *ip66Proxy {
	header := make(http.Header)
	header.Set("User-Agent", defaultUserAgent)
	ins := axios.NewInstance(&axios.InstanceConfig{
		BaseURL: "http://www.66ip.cn",
		Headers: header,
		Timeout: defaulttProxyTimeout,
	})
	ip66 := new(ip66Proxy)
	ip66.interval = interval
	ip66.ins = ins
	return ip66
}

// Start start the crawler
func (ip66 *ip66Proxy) Start() {
	ip66.status = StatusRunning
	for {
		if ip66.status != StatusRunning {
			return
		}
		_ = ip66.fetch()
		time.Sleep(ip66.interval)
	}
}

func (ip66 *ip66Proxy) fetch() (err error) {
	doc, err := ip66.fetchPage("xici", "/%d")
	if err != nil || doc == nil {
		return
	}
	// 仅在首次获取
	if ip66.maxPage == 0 {
		pages := doc.Find("#PageList a")
		value := pages.Eq(pages.Length() - 2).Text()
		max, _ := strconv.Atoi(value)
		if max == 0 {
			max = 1
		}
		ip66.maxPage = max
	}
	doc.Find("#main table tr").Each(func(i int, s *goquery.Selection) {
		// 表头忽略
		if i == 0 {
			return
		}
		tdList := s.Find("td")
		ip := tdList.Eq(0).Text()
		port := tdList.Eq(1).Text()
		if ip == "" || port == "" || ip66.fetchListener == nil {
			return
		}
		fn := ip66.fetchListener
		fn(&Proxy{
			IP:        ip,
			Port:      port,
			Anonymous: true,
			Category:  "http",
		})
	})
	return
}
