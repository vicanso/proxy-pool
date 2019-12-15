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
	// kuaiProxy kuai proxy
	kuaiProxy struct {
		baseProxyCrawler
	}
)

const (
	ProxyKuai = "kuai"
)

// NewKuaiProxy create a new kuai proxy crawler
func NewKuaiProxy(interval time.Duration) *kuaiProxy {
	header := make(http.Header)
	header.Set("User-Agent", defaultUserAgent)
	ins := axios.NewInstance(&axios.InstanceConfig{
		BaseURL: "https://www.kuaidaili.com/free/inha",
		Headers: header,
		Timeout: defaulttProxyTimeout,
	})
	kuaiProxy := new(kuaiProxy)
	kuaiProxy.interval = interval
	kuaiProxy.ins = ins
	return kuaiProxy
}

// Start start the crawler
func (kuai *kuaiProxy) Start() {
	kuai.status = StatusRunning
	for {
		if kuai.status != StatusRunning {
			return
		}
		// 获取proxy信息
		_ = kuai.fetch()
		time.Sleep(kuai.interval)
	}
}

// Fetch fetch proxy list from kuai dai li
func (kuai *kuaiProxy) fetch() (err error) {
	doc, err := kuai.fetchPage("kuai daili", "/%d/")
	if err != nil || doc == nil {
		return
	}
	if kuai.maxPage == 0 {
		pages := doc.Find("#listnav a")
		value := pages.Last().Text()
		max, _ := strconv.Atoi(value)
		if max == 0 {
			max = 1
		}
		kuai.maxPage = max
	}
	doc.Find("#list tbody tr").Each(func(i int, s *goquery.Selection) {
		tdList := s.Find("td")
		ip := tdList.Eq(0).Text()
		port := tdList.Eq(1).Text()
		category := strings.ToLower(tdList.Eq(3).Text())
		if ip == "" ||
			port == "" ||
			category == "" ||
			kuai.fetchListener == nil {
			return
		}
		kuai.fetchListener(&Proxy{
			IP:        ip,
			Port:      port,
			Anonymous: true,
			Category:  category,
		})
	})
	return
}
