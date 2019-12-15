# 应用配置

在具体抓取代理时，可根据需求设定相应的抓取列表以及检测延时等，使用[viper](https://github.com/spf13/viper)，可以方便的通过`yaml`来定义应用配置。在生成go程序，我习惯性采用单一可执行文件，将配置打包至应用程序，对于自动构建发布的流程比较适合。


## 初始化配置

```go
func init() {
	configType := "yml"
	configExt := "." + configType
	data, err := box.Find("default" + configExt)
	if err != nil {
		panic(err)
	}
	viper.SetConfigType(configType)
	v := viper.New()
	v.SetConfigType(configType)
	// 读取默认配置中的所有配置
	err = v.ReadConfig(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	configs := v.AllSettings()
	// 将default中的配置全部以默认配置写入
	for k, v := range configs {
		viper.SetDefault(k, v)
	}

	// 根据当前运行环境配置读取
	// 可根据不同的环境仅调整与default不一致的相关配置
	if env != "" {
		envConfigFile := env + configExt
		data, err = box.Find(envConfigFile)
		if err != nil {
			panic(err)
		}
		// 读取当前运行环境对应的配置
		err = viper.ReadConfig(bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
	}
}
```

## 获取抓取列表配置

```go
type (
	Crawler struct {
		Name     string
		Interval time.Duration
	}
)

// GetCrawlers get crawlers config
func GetCrawlers() []*Crawler {
	crawlers := make([]*Crawler, 0)
	data := viper.GetStringSlice("crawler")
	for _, name := range data {
		interval := viper.GetDuration(name + ".interval")
		// 如果未配置抓取间隔时间，则设置为2分钟
		if interval == 0 {
			interval = 2 * time.Minute
		}
		crawlers = append(crawlers, &Crawler{
			Name:     name,
			Interval: interval,
		})
	}
	return crawlers
}
```