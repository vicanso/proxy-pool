package main

import (
	"errors"
	"time"

	"github.com/vicanso/proxy-pool/config"
	"github.com/vicanso/proxy-pool/crawler"
)

func main() {
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
	crawler := crawler.Crawler{}
	crawler.Start(crawlerProxyList...)
	go func() {
		for range time.NewTicker(config.GetRedetectInterval()).C {
			crawler.RedetectAvailableProxy()
		}
	}()
	done := make(chan bool)
	<-done
	// crawler := crawler.NewXiciProxy()
	// crawler.OnFetch(func(data *service.Proxy) {
	// 	// fmt.Println(data)
	// })
	// crawler.Start(10 * time.Minutes)
	// proxyList, err := xc.Fetch()
	// fmt.Println(err)
	// fmt.Println(len(proxyList))
}
