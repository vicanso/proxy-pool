# proxy-pool

虽然网上各类网站提供了一堆的免费代理地址，但是其可用性比较差，更新不及时，过多不可用的地址，以及延时较大等问题都干扰实际使用的效果。对于代理地址，期望是越多越好，但是对于代理质量有着更高的要求，宁缺勿滥，因此`proxy-pool`不再将抓取到的代理地址保存至数据库，而调整为定期从免费代理网站下抓取代理地址，使用该地址去测试其可用性（默认配置为访问baidu），测试可用则添加至可用代理地址列表中，如此循环一直抓取新的地址，一直校验。对于已校验可用的代理地址，也定期重新校验是否可用，默认校验间隔为30分钟。

注意：网页部分有增加百度统计，部署时可先删除。

<p align="center">
  <img src="https://raw.githubusercontent.com/vicanso/proxy-pool/master/assets/proxy-pool.jpg">
</p>

## 常用配置

对于有特别需求，可以调整默认的配置，主要的配置如下：

抓取代理网站列表配置（暂时只实现了三个网站的抓取）：

```yml
crawler:
- xici
- ip66
- kuai
```

由于各网站对访问IP频率限制的不同，可根据实际使用中调整各网站的抓取间隔，如设置`xici`的抓取延时为10分钟（如果不配置则为默认值2分钟）：

```yml
xici:
  interval: 10m
```

默认的检测方式是通过代理地址去访问`baidu`，可根据应用场景调整相应的配置：

```yml
detect:
  # 检测时间（定时对现可用的代理地址重新检测）
  interval: 30m
  # 检测地址
  url: https://www.baidu.com/
  # 检测超时
  timeout: 3s
  # 最大次数
  maxTimes: 3
```

## 程序设计

- [config](./doc/config.md)
- [crawler](./doc/crawler.md)
