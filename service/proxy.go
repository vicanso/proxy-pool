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

package service

import (
	"errors"
	"time"

	"github.com/vicanso/proxy-pool/config"
	"github.com/vicanso/proxy-pool/crawler"
)

var (
	defaultCrawler = new(crawler.Crawler)
)

func init() {
	crawlerProxyList := make([]crawler.ProxyCrawler, 0)
	for _, item := range config.GetCrawlers() {
		interval := item.Interval
		var c crawler.ProxyCrawler
		switch item.Name {
		case crawler.ProxyIP66:
			c = crawler.NewIP66Proxy(interval)
		case crawler.ProxyKuai:
			c = crawler.NewKuaiProxy(interval)
		default:
			c = crawler.NewXiciProxy(interval)
		}
		crawlerProxyList = append(crawlerProxyList, c)
	}
	if len(crawlerProxyList) == 0 {
		panic(errors.New("no proxy crawler"))
	}
	defaultCrawler.Start(crawlerProxyList...)
	go func() {
		detectConfig := config.GetDetect()
		for range time.NewTicker(detectConfig.Interval).C {
			defaultCrawler.RedetectAvailableProxy()
		}
	}()
}

// GetAvailableProxyList get available proxy lsit
func GetAvailableProxyList() []*crawler.Proxy {
	return defaultCrawler.GetAvailableProxyList()
}

// GetAvailableProxy get available proxy
func GetAvailableProxy(category string, speed int) *crawler.Proxy {
	return defaultCrawler.GetAvailableProxy(category, int32(speed))
}
