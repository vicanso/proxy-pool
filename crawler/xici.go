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
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/vicanso/go-axios"
)

type (
	// xiciProxy xici proxy
	xiciProxy struct {
		baseProxyCrawler
	}
)

const (
	ProxyXiCi = "xici"
)

// NewXiciProxy create a new xici proxy crawler
func NewXiciProxy(interval time.Duration) *xiciProxy {
	header := make(http.Header)
	header.Set("User-Agent", defaultUserAgent)
	ins := axios.NewInstance(&axios.InstanceConfig{
		BaseURL: "https://www.xicidaili.com/nn",
		Headers: header,
		Timeout: defaulttProxyTimeout,
	})
	xiciProxy := new(xiciProxy)
	xiciProxy.interval = interval
	xiciProxy.ins = ins
	return xiciProxy
}

// Start start the crawler
func (xc *xiciProxy) Start() {
	xc.status = StatusRunning
	for {
		if xc.status != StatusRunning {
			return
		}
		// 获取proxy信息
		_ = xc.fetch()
		time.Sleep(xc.interval)
	}
}

// Fetch fetch proxy list from xici
func (xc *xiciProxy) fetch() (err error) {
	doc, err := xc.fetchPage("xici", "/%d")
	if err != nil || doc == nil {
		return
	}
	// 仅在首次获取
	if xc.maxPage == 0 {
		pages := doc.Find(".pagination a")
		value := pages.Eq(pages.Length() - 2).Text()
		max, _ := strconv.Atoi(value)
		if max == 0 {
			max = 1
		}
		xc.maxPage = max
	}
	// 解析表格获取代理列表
	doc.Find("#ip_list tr").Each(func(i int, s *goquery.Selection) {
		// 表头忽略
		if i == 0 {
			return
		}
		tdList := s.Find("td")
		ip := tdList.Eq(1).Text()
		port := tdList.Eq(2).Text()
		anonymous := tdList.Eq(4).Text() == "高匿"
		category := strings.ToLower(tdList.Eq(5).Text())

		if ip == "" ||
			port == "" ||
			category == "" ||
			xc.fetchListener == nil {
			return
		}
		xc.fetchListener(&Proxy{
			IP:        ip,
			Port:      port,
			Anonymous: anonymous,
			Category:  category,
		})
	})
	return
}
