# 代理IP抓取

代理IP地址主要通过抓取提供免费代理IP的网站，如`西刺`。通过获取其网页数据，筛选出代理服务器的信息，如：IP、端口、支持的代理类型以及是否匿名代理。

根据需求先定义一个Proxy的结构来保存代理信息，一个ProxyCrawler的interface，便于后面扩展从不同的网站上抓取代理信息。


```go
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
```

ProxyCrawler的OnFetch函数用于设定获取proxy成功时的回调，各抓取服务在抓取到相应的代理信息之后，则触发此回调，为什么使用这样的实现形式呢？因为各网站都会对访问频率有限制，因此无法快速的获取所有代理信息，只能抓取一页数据之后，延时再去抓取的方式，使用回调则可以在获取到一个代理信息时，则立即触发回调。

下面开始编写对`西刺`的数据抓取，主要获取其高匿代理的数据，地址为：`https://www.xicidaili.com/nn/`，其数据是以html的形式返回，展示形式为表格，因此抓取的方法也比较简单，HTTP请求获取数据，解析html，从表格中获取相应的代理信息。

首先将抓取IP地址的功能提取公共模块，`baseProxyCrawler`:

```go
// baseProxyCrawler base proxy crawler
// nolint
type baseProxyCrawler struct {
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
		logger.Error("get proxy list fail",
			zap.String("name", name),
			zap.Int("page", bp.currentPage),
			zap.Error(err),
		)
		return
	}
	logger.Info("get proxy list success",
		zap.String("name", name),
		zap.Int("page", bp.currentPage),
	)
	return goquery.NewDocumentFromReader(bytes.NewReader(resp.Data))
}
```

下面是具体抓取`西刺`代理的免费IP地址的代码如下：

```go
type (
	// xiciProxy xici proxy
	xiciProxy struct {
		baseProxyCrawler
	}
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
	if err != nil {
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
```

至此抓取免费代理IP地址已经完成，后续增加了对`66ip`以及`快代理`的抓取，如果还有其它网站可供抓取，只需参考实现抓取的流程则可。

# 代理IP连通性测试

在实际使用时发现大部分抓取到的IP地址都不可用，因此增加检测模块，定时对新抓取的代理IP检测（使用此IP做代理去访问baidu），访问成功的则记录至可用代理列表中。

## 代理列表

代理列表用于保存代理地址，提供判断是否存在、添加、重置等方法，主要用于保存新抓取的代理地址以及可用代理地址，部分代码如下：

```go
type (
	// ProxyList proxy list
	ProxyList struct {
		sync.RWMutex
		data []*Proxy
	}
)

// Exists test whether or not the proxy exists
func (pl *ProxyList) Exists(p *Proxy) bool {
	pl.RLock()
	defer pl.RUnlock()
	found := false
	for _, item := range pl.data {
		if item.IP == p.IP &&
			item.Port == p.Port &&
			item.Category == p.Category {
			found = true
			break
		}
	}
	return found
}

// Add add proxy to list
func (pl *ProxyList) Add(p *Proxy) {
	if pl.Exists(p) {
		return
	}
	pl.Lock()
	defer pl.Unlock()
	if len(pl.data) == 0 {
		pl.data = make([]*Proxy, 0, 100)
	}
	pl.data = append(pl.data, p)
}

```

## 代理IP检测

对代理IP检测主要判断其可用性以及连接速度（默认为3秒如果无响应则不可用），检测逻辑比较简单，创建新的Transport，指定其使用的Proxy，如果成功返回则计算处理时间，并添加至可用代理列表。

```go
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
			IdleConnTimeout:       10 * time.Second,
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
	// 多次检测，只要一次成功则认为成功
	for i := 0; i < detectConfig.MaxTimes; i++ {
		ins := axios.NewInstance(&axios.InstanceConfig{
			Timeout: detectConfig.Timeout,
			Client:  httpClient,
		})
		startedAt := time.Now()
		resp, err := ins.Get(detectConfig.URL)
		if err != nil {
			continue
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
			available = true
			break
		}
	}
	return
}
```

对于新抓取代理IP的检测使用的是定时检测的方式，在每次对当前新增代理列表检测完成之后，等待1分钟后再进去下一次检测。

```go
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
```

而对于可用代理IP列表，也需要定期去检测是否可还是可用，在多次检测均失败则认为此代理也不可用，从可用代理列表中删除，逻辑如下：

```go
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
```
