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
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/vicanso/go-axios"
	"github.com/vicanso/proxy-pool/log"
	"go.uber.org/zap"
)

const (
	StatusRunning = iota
	StatusStop
)

const (
	detectRunning = iota + 1
	detectStop
)

const (
	defaultUserAgent     = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36"
	defaulttProxyTimeout = 10 * time.Second
	defaultDetectURL     = "https://www.baidu.com"
	defaultDetectTimeout = 3 * time.Second
)

var (
	speedDevides = []time.Duration{750 * time.Millisecond, 1500 * time.Millisecond}

	logger = log.Default()
)

type (
	// Crawler crawler
	Crawler struct {
		sync.Mutex
		HTTPDetectURL              string
		HTTPSDetectURL             string
		newProxyList               ProxyList
		avaliableProxyList         ProxyList
		newProxyDetectStatus       int32
		availableProxyDetectStatus int32
	}
	// baseProxyCrawler base proxy crawler
	// nolint
	baseProxyCrawler struct {
		// 每次抓取代理信息间隔（需要注意不同的网站对访问频率有不同的限制，不要设置太短）
		interval time.Duration
		// axios http实例
		ins *axios.Instance
		// 获取到IP的回调函数
		fetchListener FetchListener
		// 当前页
		currentPage int
		// 最大页数
		maxPage int
		// 状态，运行中或停止
		status int32
	}
	// FetchListener fetch listener
	FetchListener func(*Proxy)
	// ProxyCrawler proxy crawler
	ProxyCrawler interface {
		// OnFetch set fetch listener
		OnFetch(FetchListener)
		// Start start the crawler
		Start()
		// Stop stop the crawler
		Stop()
	}
)

// OnFetch set fetch listener
func (bp *baseProxyCrawler) OnFetch(fn FetchListener) {
	bp.fetchListener = fn
}

// Stop stop the crawler
func (bp *baseProxyCrawler) Stop() {
	atomic.StoreInt32(&bp.status, StatusStop)
}

// fetchPage fetch html content of the current page
func (bp *baseProxyCrawler) fetchPage(name, urlTemplate string) (doc *goquery.Document, err error) {
	ins := bp.ins
	// 至最后一页则重置页码
	if bp.maxPage != 0 && bp.currentPage == bp.maxPage {
		bp.currentPage = 0
		bp.maxPage = 0
	}
	bp.currentPage++
	resp, err := ins.Get(fmt.Sprintf(urlTemplate, bp.currentPage))
	// 对于抓取失败，则直接退出
	if err != nil ||
		resp.Status != http.StatusOK ||
		len(resp.Data) == 0 {
		logger.Error(
			name+" get proxy list fail",
			zap.Int("page", bp.currentPage),
			zap.Error(err),
		)
		return
	}
	return goquery.NewDocumentFromReader(bytes.NewReader(resp.Data))
}

// NewProxyClient create a new http client with proxy
func NewProxyClient(p *Proxy) *http.Client {
	proxyURL, _ := url.Parse(fmt.Sprintf("http://%s:%s", p.IP, p.Port))
	if proxyURL == nil {
		return nil
	}
	return &http.Client{
		Transport: &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return proxyURL, nil
			},
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// analyze check the proxy is available and speed
func (c *Crawler) analyze(p *Proxy) (available bool) {
	httpClient := NewProxyClient(p)
	if httpClient == nil {
		return false
	}
	ins := axios.NewInstance(&axios.InstanceConfig{
		Timeout: defaultDetectTimeout,
		Client:  httpClient,
	})
	startedAt := time.Now()
	resp, err := ins.Get(defaultDetectURL)
	if err != nil {
		return false
	}
	if resp.Status >= http.StatusOK && resp.Status < http.StatusBadRequest {
		d := time.Since(startedAt)
		atomic.StoreInt32(&p.Speed, int32(len(speedDevides)))
		// 将当前proxy划分对应的分段
		for index, item := range speedDevides {
			if d < item {
				atomic.StoreInt32(&p.Speed, int32(index))
				break
			}
		}
	}
	return true
}

// addNewProxy add proxy to new proxy list
func (c *Crawler) addNewProxy(p *Proxy) {
	c.newProxyList.Add(p)
}

// detectProxyList detect proxy list
func (c *Crawler) detectProxyList(list []*Proxy) (availableList []*Proxy, unavailableList []*Proxy) {
	availableList = make([]*Proxy, 0)
	unavailableList = make([]*Proxy, 0)
	w := sync.WaitGroup{}
	// 控制最多检测proxy的数量
	chans := make(chan bool, 5)
	for _, item := range list {
		w.Add(1)
		go func(p *Proxy) {
			chans <- true
			avaliable := c.analyze(p)
			atomic.StoreInt64(&p.DetectedAt, time.Now().Unix())
			if avaliable {
				availableList = append(availableList, p)
			} else {
				unavailableList = append(unavailableList, p)
			}
			<-chans
			w.Done()
		}(item)
	}
	w.Wait()
	return
}

// detectNewProxy detect the new proxy is avaliable
func (c *Crawler) detectNewProxy() {
	old := atomic.SwapInt32(&c.newProxyDetectStatus, detectRunning)
	// 如果已经在运行中，则直接退出
	if old == detectRunning {
		return
	}
	proxyList := c.newProxyList.Reset()
	availableList, _ := c.detectProxyList(proxyList)
	c.avaliableProxyList.Add(availableList...)

	atomic.StoreInt32(&c.newProxyDetectStatus, detectStop)
	// 等待1分钟后，重新运行detect new proxy
	time.Sleep(time.Minute)
	c.detectNewProxy()
}

// RedetectAvailableProxy redetect available proxy
func (c *Crawler) RedetectAvailableProxy() {
	old := atomic.SwapInt32(&c.availableProxyDetectStatus, detectRunning)
	// 如果已经在运行中，则直接退出
	if old == detectRunning {
		return
	}
	proxyList := c.avaliableProxyList.List()
	availableList, unavailableList := c.detectProxyList(proxyList)

	// 如果成功，则重置失败次数
	for _, p := range availableList {
		atomic.StoreInt32(&p.Fails, 0)
	}
	// 如果失败，则失败次数+1
	failProxyList := make([]*Proxy, 0)
	for _, p := range unavailableList {
		count := atomic.AddInt32(&p.Fails, 1)
		if count >= 3 {
			failProxyList = append(failProxyList, p)
		}
	}
	// 对于三次检测失败的代理则删除
	c.avaliableProxyList.Remove(failProxyList...)

	atomic.StoreInt32(&c.availableProxyDetectStatus, detectStop)
}

// Start start fetch proxy
func (c *Crawler) Start(crawlers ...ProxyCrawler) {
	for _, item := range crawlers {
		item.OnFetch(c.addNewProxy)
		go item.Start()
	}
	// 首次延时10秒后则执行detect new proxy
	go func() {
		time.Sleep(10 * time.Second)
		c.detectNewProxy()
	}()
}
