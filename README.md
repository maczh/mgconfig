# mgconfig Go语言统一配置管理

## 配置文件格式

仅支持.yml格式

## 支持的统一配置中心

+ Nacos
+ Spring Cloud Config
+ Consul

## 支持的微服务发现与注册中心

+ Nacos
+ Consul

## 支持的数据库

+ MySQL （Gorm v1)
+ PostgreSQL (Gorm v1)
+ MS SQL Server (Gorm v1)
+ MongoDB (Mgo v2)
+ Redis (go-redis)
+ CouchBase
+ SSDB
+ HBase (gohbase)
+ Hive (gohive)
+ InfluxDB

## 支持的消息队列

+ RabbitMQ (jazz)

## 支持的搜索引擎

+ ElasticSearch (olivere/elastic)
+ 阿里云 OpenSearch

## 安装
```
go get -u github.com/maczh/mgconfig
```

## 使用方法

### 本地配置文件

+ 默认文件名为`application.yml`，可自定义名称，配置内容如下
```
go:
  application:
    name: myapp         #应用名称,用于自动注册微服务时的服务名
    port: 8080          #端口号
    ip: xxx.xxx.xxx.xxx  #微服务注册时登记的本地IP，不配可自动获取，如需指定外网IP或Docker之外的IP时配置
  discovery: nacos                      #微服务的服务发现与注册中心类型 nacos,consul,默认是 nacos
  config:                               #统一配置服务器相关
    server: http://192.168.1.5:8848/    #配置服务器地址
    server_type: nacos                  #配置服务器类型 nacos,consul,springconfig
    env: test                           #配置环境 一般常用test/prod/dev等，跟相应配置文件匹配
    type: .yml                          #文件格式，目前仅支持yaml
    mid: -go-                           #配置文件中间名
    used: nacos,mysql,mongodb,redis     #当前应用启用的配置
    prefix:                             #配置文件名前缀定义
      mysql: mysql                      #mysql对应的配置文件名前缀，如当前配置中对应的配置文件名为 mysql-go-test.yml
      mongodb: mongodb
      redis: redis
      ssdb: ssdb
      rabbitmq: rabbitmq
      nacos: nacos
      pgsql: pgsql
      mssql: mssql
      consul: consul
      elasticsearch: elasticsearch
      opensearch: opensearch
      hbase: hbase
      hive: hive
      couchdb: couchdb
      influxdb: influxdb
```
+ mysql配置范例 mysql-go-test.yml
```
go:
  data:
    mysql: user:pwd@tcp(xxx.xxx.xxx.xxx:3306)/dbname?charset=utf8&parseTime=True&loc=Local
    mysql_pool:     #连接池设置,若无此项则使用单一长连接
      max: 20     #最大连接数
      min: 5      #最小连接数
      idle: 10    #空闲连接数
      timeout: 300  #空闲超时少数，超时后自动断开空闲连接  
```
+ mongodb配置范例 mongodb-go-test.yml
```
go:
  data:
    mongodb:
      uri: mongodb://user:pwd@xxx.xxx.xxx.xxx:port/dbname
      db: dbname
    mongo_pool:     #连接池设置,若无此项则使用单一长连接
      max: 20     #最大连接数
      min: 5      #最小连接数
      idle: 10    #空闲连接数
      timeout: 300  #空闲超时少数，超时后自动断开空闲连接  
```
+ redis配置范例 redis-go-test.yml
```
go:
  data:
    redis:
      database: 1
      host: xxx.xxx.xxx.xxx
      password: pwd
      pool:
        max-active: 10
        max-idle: 10
        max-wait: -1
        min-idle: 1
      port: 6379
      timeout: 1000
```
+ ssdb配置范例 ssdb-go-test.yml
```
go:
  data:
    ssdb:
      host: xxx.xxx.xxx.xxx
      password: pwd
      port: 8888
      timeout: 3000
```
+ nacos配置范例 nacos-go-test.yml
```
go:
  nacos:
    clusterName: DEFAULT
    port: 8848
    server: xxx.xxx.xxx.xxx
    weight: 1
```
+ rabbitmq配置范例 rabbitmq-go-test.yml
```
go:
  rabbitmq:
    uri: amqp://user:password@xxx.xxx.xxx.xxx:5672/vhost
    exchange: ex1
```
如果vhost中有/，那在vhost之前加上%2f

### 在应用中加载配置

* 在main中加载配置，加载之后所有go.config.used配置中指定的数据库均已经从配置服务器获取相应配置并自动创建连接，数据库公共对象即已经可用
* 加载配置之后，如果used包含nacos或consul，则自动在相应的服务发现与注册中心自动注册本微服务
```
    import "github.com/maczh/mgconfig"

    config_file := "/path/to/myapp.yml"
    mgconfig.InitConfig(config_file)
```

### 程序退出时关闭所有数据库连接和注销服务

```
    mgconfig.SafeExit()
```

### 在应用中使用MySQL单连接范例

```
func GetUserById(id uint) (*pojo.User) {
	user := new(pojo.User)
	mgconfig.Mysql.Table("user_info").Where("id = ?",id).First(&user)
	logs.Debug("查詢結果:",user)
	return user
}
```

###在应用中使用MySQL连接池范例

```
func GetUserById(id uint) (*pojo.User) {
	user := new(pojo.User)
    //从连接池中获取连接
    mysql := mgconfig.GetMysqlConnection()
	mysql.Table("user_info").Where("id = ?",id).First(&user)
	logs.Debug("查詢結果:",user)
    //归还连接到连接池
    mgconfig.ReturnMysqlConnection(mysql)
	return user
}
```

### 修改默认数据库检查时间，默认为5分钟一次

在`conf.go`文件头部修改常量

```
const AUTO_CHECK_MINUTES = 5	//自动检查连接间隔时间，单位为分钟
```

### 在应用中读取主配置文件内容

```
    //读取字符串配置
    mgconfig.GetConfigString("path.img.temp")
    //读取整数配置
    mgconfig.GetConfigInt("online.connection.max")
```