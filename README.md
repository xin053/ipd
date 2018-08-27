# ipd

批量ip反查服务, 批量获取ip的地理位置信息,包括国家,省份,城市,ISP以及地理位置,支持国内外ip，目前仅支持ipv4.

基于[`gin框架`](https://github.com/gin-gonic/gin)构建, 默认端口`6789`,可在`config.toml`中修改服务端口以及其他配置项

## 从源代码构建

1. 下载包管理器[`dep`](https://github.com/golang/dep)

    ```shell
    go get -u github.com/golang/dep/cmd/dep
    ```

    **确保`dep`在环境变量`PATH`中**
2. 下载[`ipd`](https://github.com/xin053/ipd)源码

    ```shell
    export GOPATH=`pwd`
    go get -d github.com/golang/xin053/ipd
    ```

3. 安装依赖

    ```shell
    dep ensure
    ```

4. 构建`ipd`可执行文件

    ```shell
    cd github.com/golang/xin053/ipd
    go build ipd.go
    ```

    或者

    ```shell
    go build -ldflags "-w -s" ipd.go
    ```

5. 使用

    ```shell
    # 运行
    ./ipd
    # 停止
    kill -2 pid
    ```

## 使用

`ipd`提供四种ip反查的方式:

1. 通过[纯真ip数据库](http://www.cz88.net/fox/ipdat.shtml)(目前更新到`2018-08-25`)
2. 通过[GeoLite2数据库](https://dev.maxmind.com/zh-hans/geoip/geoip2/geolite2-%E5%BC%80%E6%BA%90%E6%95%B0%E6%8D%AE%E5%BA%93/)
3. 通过[ip2region数据库](https://github.com/lionsoul2014/ip2region),启动服务时会自动从github下载最新的数据库文件
4. 通过公开的 REST API方式, 目前支持四种接口:

    1. 淘宝ip查询接口:`http://ip.taobao.com/service/getIpInfo.php?ip=`
    2. 新浪ip查询接口:`http://int.dpool.sina.com.cn/iplookup/iplookup.php?format=json&ip=`
    3. 太平洋ip查询接口:`http://whois.pconline.com.cn/ipJson.jsp?ip=`
    4. 百度ip查询接口:`http://api.map.baidu.com/location/ip?ak=yourAK&ip=`

### API

1. 从公共API获取ip信息

    **接口**

    `POST /v1/api`

    **请求头**

    1. `Authorization` = `thisisaveryimportantkey`(key可在`config.go`中配置)
    2. `Content-Type` = `application/json`

    **Body**

    ```json
    {
        "ip":
            ["111.111.111.111", "8.8.8.8"]
    }
    ```

    **返回**

    ```json
    [
        {
            "ip": "8.8.8.8",
            "country": "美国",
            "region": "",
            "city": "",
            "isp": "Google公共DNS",
            "geo_x": 0,
            "geo_y": 0
        },
        {
            "ip": "111.111.111.111",
            "country": "",
            "region": "",
            "city": "",
            "isp": "日本东京市KDDI通信公司",
            "geo_x": 0,
            "geo_y": 0
        }
    ]
    ```

2. 从纯真ip数据库获取ip信息

    **接口**

    `POST /v1/db`

    **请求头**

    1. `Authorization` = `thisisaveryimportantkey`(key可在`config.go`中配置)
    2. `Content-Type` = `application/json`

    **Body**

    ```json
    {
        "ip":
            ["111.111.111.111", "8.8.8.8"]
    }
    ```

    **返回**

    ```json
    [
        {
            "ip": "8.8.8.8",
            "country": "美国",
            "region": "",
            "city": "",
            "isp": "Google公共DNS",
            "geo_x": 0,
            "geo_y": 0
        },
        {
            "ip": "111.111.111.111",
            "country": "",
            "region": "",
            "city": "",
            "isp": "日本东京市KDDI通信公司",
            "geo_x": 0,
            "geo_y": 0
        }
    ]
    ```

3. 从GeoLite2数据库获取ip信息

    **接口**

    `POST /v1/db2`

    **请求头**

    1. `Authorization` = `thisisaveryimportantkey`(key可在`config.go`中配置)
    2. `Content-Type` = `application/json`

    **Body**

    ```json
    {
        "ip":
            ["111.111.111.111", "8.8.8.8"]
    }
    ```

    **返回**

    ```json
    [
        {
            "ip": "8.8.8.8",
            "country": "美国",
            "region": "",
            "city": "",
            "isp": "Google公共DNS",
            "geo_x": 0,
            "geo_y": 0
        },
        {
            "ip": "111.111.111.111",
            "country": "",
            "region": "",
            "city": "",
            "isp": "日本东京市KDDI通信公司",
            "geo_x": 0,
            "geo_y": 0
        }
    ]
    ```

4. 从ip2region数据库获取ip信息

    **接口**

    `POST /v1/db3`

    **请求头**

    1. `Authorization` = `thisisaveryimportantkey`(key可在`config.go`中配置)
    2. `Content-Type` = `application/json`

    **Body**

    ```json
    {
        "ip":
            ["111.111.111.111", "8.8.8.8"]
    }
    ```

    **返回**

    ```json
    [
        {
            "ip": "8.8.8.8",
            "country": "美国",
            "region": "",
            "city": "",
            "isp": "Google公共DNS",
            "geo_x": 0,
            "geo_y": 0
        },
        {
            "ip": "111.111.111.111",
            "country": "",
            "region": "",
            "city": "",
            "isp": "日本东京市KDDI通信公司",
            "geo_x": 0,
            "geo_y": 0
        }
    ]
    ```

5. 整合以上四种方式获取ip信息,先异步查ip2region数据库(默认的主数据库),查不到的ip再查纯真, 再查GeoLite2数据库,最后通过api查询

    主数据库可在配置文件中配置(`config.toml`中的`request_order`配置)

    **接口**

    `POST /v1/ip`

    **请求头**

    1. `Authorization` = `thisisaveryimportantkey`(key可在`config.go`中配置)
    2. `Content-Type` = `application/json`

    **Body**

    ```json
    {
        "ip":
            ["111.111.111.111", "8.8.8.8"]
    }
    ```

    **返回**

    ```json
    [
        {
            "ip": "8.8.8.8",
            "country": "美国",
            "region": "",
            "city": "",
            "isp": "Google公共DNS",
            "geo_x": 0,
            "geo_y": 0
        },
        {
            "ip": "111.111.111.111",
            "country": "",
            "region": "",
            "city": "",
            "isp": "日本东京市KDDI通信公司",
            "geo_x": 0,
            "geo_y": 0
        }
    ]
    ```

## 项目结构

```
api\
   |api.go                      # api 方式查询 ip 信息主文件
   |baidu.go                    # 百度 ip API 服务解析
   |base.go                     # 通用接口
   |pconline.go                 # 太平洋 ip API 服务解析
   |sina.go                     # 新浪 ip API 服务解析
   |taobao.go                   # 淘宝 ip API 服务解析
config\
      |config.go                # 从 config.toml 读取配置以及其他配置
es\
  |es.go                        # elasticsearch 存储相关
geolite2\
        |geolite2.go            # db2 方式查询 ip 信息主文件
ip2region\
         |ip2region             # db3 方式查询 ip 信息主文件
         |lib                   # 解析 ip2region.db 的库文件
middleware\
          |auth.go              # 简单授权验证中间件
          |cors.go              # cors 跨域中间件
          |json_logger.go       # 日志服务中间件
          |sentry.go            # sentry 服务
qqwry\
     |qqwry.go                  # db 方式查询 ip 信息主文件
server\
     |server.go                 # /v1/ip 接口主文件
utils\
     |utils_test.go             # 工具包测试
     |utils.go                  # 工具包
config.toml                     # ipd 服务使用的配置文件
GeoLite2-City.mmdb              # geolite2 数据库二进制文件
Gopkg.lock                      # dep 包管理器 lock 文件
Gopkg.toml                      # dep 包管理器 toml 文件
ip2region.db                    # ip2region 数据库二进制文件
ipd.go                          # ipd 服务 main 包
qqwry.dat                       # 纯真 ip 数据库二进制文件
README.md                       # README
```

## 构建自己的ip信息数据库

ipd服务支持将查询过的ip信息添加到elasticsearch数据库,作后续其他的使用.默认启用elasticsearch存储

1. 安装[elasticsearch](https://www.elastic.co/products/elasticsearch),建议安装最新版
2. 修改`config.toml`中的`elasticsearch`段的`url`, 如果不想使用存储ip功能,将`elasticsearch`段注释即可

## 其他事项

1. 程序默认使用sentry服务, 修改`config.go`中的`dsn`以使用自己的sentry,也可以注释掉`config.toml`中的`sentry`段来禁用sentry
2. 更多设置请查看`config.toml`的备注说明

## TODO

- [ ] 添加es接口，直接从es查询ip数据
- [ ] 增加精准查询接口(多种查询方式同时 goroutine，获取接口分析取最优)
- [x] 添加缓存机制