# mgconfig Go语言统一配置管理

最新版本号 v1.1.12

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

+ MySQL （Gorm v1/v2)
+ PostgreSQL (Gorm v1)
+ MS SQL Server (Gorm v1)
+ MongoDB (Mgo v2)
+ Redis (go-redis)
+ CouchBase
+ HBase (gohbase)
+ Hive (gohive)
+ InfluxDB

## 支持的消息队列

+ RabbitMQ (jazz)
+ Kafka (sarama)

## 支持的搜索引擎

+ ElasticSearch (olivere/elastic)

## 安装
```shell script
go get -u github.com/maczh/mgconfig
```

## 使用方法

### 本地配置文件

+ 默认文件名为`application.yml`，可自定义名称，配置内容如下
```yaml
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
    mid: "-"                            #配置文件中间名
    used: nacos,mysql,mongodb,redis     #当前应用启用的配置,MySQL代表使用GORM v2版本的MySQL，小写mysql代表GORM v1版本
    prefix:                             #配置文件名前缀定义
      mysql: mysql                      #mysql对应的配置文件名前缀，如当前配置中对应的配置文件名为 mysql-go-test.yml
      mongodb: mongodb
      redis: redis
      rabbitmq: rabbitmq
      nacos: nacos
      pgsql: pgsql
      mssql: mssql
      consul: consul
      elasticsearch: elasticsearch
      hbase: hbase
      hive: hive
      couchdb: couchdb
      influxdb: influxdb
```
+ mysql配置范例 mysql-test.yml
```yaml
go:
  data:
    mysql: user:pwd@tcp(xxx.xxx.xxx.xxx:3306)/dbname?charset=utf8&parseTime=True&loc=Local
    mysql_debug: true   #打开调试模式
    mysql_pool:     #连接池设置,若无此项则使用单一长连接
      max: 200      #实际最大连接数
      total: 1000   #最大并发数,不填默认为最大连接数5倍
      timeout: 30   #空闲连接超时，秒，默认60秒
      life: 5       #连接生命周期，分钟，默认60分钟
```
+ mysql多库连接配置范例 mysql-multidb-test.yml
```yaml
go:
  data:
    mysql: 
      multidb: true
      dbNames: test1,test2
      test1: user1:pwd1@tcp(xxx.xxx.xxx.xxx:3306)/dbname1?charset=utf8&parseTime=True&loc=Local
      test2: user2:pwd2@tcp(xxx.xxx.xxx.xxx:3306)/dbname2?charset=utf8&parseTime=True&loc=Local
    mysql_debug: true   #打开调试模式
    mysql_pool:     #连接池设置,若无此项则使用单一长连接
      max: 200      #实际最大连接数
      total: 1000   #最大并发数,不填默认为最大连接数5倍
      timeout: 30   #空闲连接超时，秒，默认60秒
      life: 5       #连接生命周期，分钟，默认60分钟
```


+ mongodb配置范例 mongodb-test.yml
```yaml
go:
  data:
    mongodb:
      uri: mongodb://user:pwd@xxx.xxx.xxx.xxx:port/dbname #当使用复制集时 mongodb://user:pwd@192.168..3.5:27017,192.168.3.6:27017/dbname?replicaSet=replsetname
      db: dbname
      debug: true   #打开调试模式
    mongo_pool:     #连接池设置,若无此项则使用单一长连接
      max: 20       #最大连接数
```

+ mongodb多库连接配置范例 mongodb-multidb-test.yml
```yaml
go:
  data:
    mongodb:
      multidb: true
      dbNames: test1,test2
      test1:
          uri: mongodb://user1:pwd1@xxx.xxx.xxx.xxx:port/dbname1 #当使用复制集时 mongodb://user:pwd@192.168..3.5:27017,192.168.3.6:27017/dbname?replicaSet=replsetname
          db: dbname1
      test2:
        uri: mongodb://user2:pwd2@xxx.xxx.xxx.xxx:port/dbname2 #当使用复制集时 mongodb://user:pwd@192.168..3.5:27017,192.168.3.6:27017/dbname?replicaSet=replsetname
        db: dbname2
      debug: true   #打开调试模式
    mongo_pool:     #连接池设置,若无此项则使用单一长连接
      max: 20       #最大连接数
```


+ redis配置范例 redis-test.yml
```yaml
go:
  data:
    redis:
      host: xxx.xxx.xxx.xxx
      port: 6379
      password: password
      database: 1
      timeout: 1000
    redis_pool:
      min: 3        #最小空闲连接数,默认2
      max: 200      #连接池大小，最小默认10
      idle: 10      #空闲超时，分钟,默认5分钟
      timeout: 300  #连接超时，秒，默认60秒
```

+ redis多库连接配置范例 redis-multidb-test.yml
```yaml
go:
  data:
    redis:
      multidb: true
      dbNames: test1,test2
      test1:
          host: xxx.xxx.xxx.xxx
          port: 6379
          password: password
          database: 1
      test2:
        host: xxx.xxx.xxx.xxx
        port: 6379
        password: password
        database: 2
    redis_pool:
      min: 3        #最小空闲连接数,默认2
      max: 200      #连接池大小，最小默认10
      idle: 10      #空闲超时，分钟,默认5分钟
      timeout: 300  #连接超时，秒，默认60秒
```

+ nacos配置范例 nacos-test.yml
```yaml
go:
  nacos:
    server: xxx.xxx.xxx,xxx.xxx.xxx,xxx.xxx.xxx   #nacos集群，多台IP
    port: 8848,8848,8848   #nacos集群，对应IP多个端口
    clusterName: DEFAULT
    weight: 1
    lan: true   #以内网地址注册，否则以公网地址注册
```
+ rabbitmq配置范例 rabbitmq-go-test.yml
```yaml
go:
  rabbitmq:
    uri: amqp://user:password@xxx.xxx.xxx.xxx:5672/vhost
    exchange: ex1
```
如果vhost中有/，那在vhost之前加上%2f

### 在应用中加载配置

* 在main中加载配置，加载之后所有go.config.used配置中指定的数据库均已经从配置服务器获取相应配置并自动创建连接，数据库公共对象即已经可用
* 加载配置之后，如果used包含nacos或consul，则自动在相应的服务发现与注册中心自动注册本微服务
```go
    import "github.com/maczh/mgconfig"

    config_file := "/path/to/myapp.yml"
    mgconfig.InitConfig(config_file)
```

### 程序退出时关闭所有数据库连接和注销服务

```go
    mgconfig.SafeExit()
```

### 在应用中使用MySQL单连接范例(Gorm v2),连接池
- 单库连接
```go
func GetUserById(id uint) (*pojo.User) {
    user := new(pojo.User)
    //从连接池中获取连接
    mysql,err := mgconf.GetMySQLConnection()
    if err != nil {
    logs.Error("MySQL connection error: {}",err.Error())
    return nil
    }
    mysql.Table("user_info").Where("id = ?",id).First(&user)
    logs.Debug("查詢結果:{}",user)
    //归还连接到连接池(无需显式归还)
     return user
}
```

- 多库连接时
```go
func GetUserById(id uint, dbName string) (*pojo.User) {
    user := new(pojo.User)
    //从连接池中获取连接
    mysql,err := mgconf.GetMySQLConnection(dbName)
    if err != nil {
    logs.Error("MySQL connection error: {}",err.Error())
    return nil
    }
    mysql.Table("user_info").Where("id = ?",id).First(&user)
    logs.Debug("查詢結果:{}",user)
    //归还连接到连接池(无需显式归还)
     return user
}
```

### v1.1.12版本更新说明
+ 增加kafka支持

### v1.1.10版本更新说明
+ 新增RabbitMQ创建死信队列函数

### v1.1.9版本更新说明
+ MySQL支持多库连接
+ mongodb支持多库连接
+ redis支持多库连接

### v1.0.9版本更新说明
+ mysql/MySQL改用database/sql自带的连接池，且必须使用连接池配置，不需要归还连接，但必须使用获取连接函数，要处理获取连接返回的error
+ redis改用go-redis自带连接池，同样必须采用获取连接函数，且需要处理error，也无需归还连接
+ mongodb改用mgo.v2自带连接池，必须使用获取连接函数，且需要处理error，***需要归还连接***，支持replicaSet，自动读写分离
+ nacos支持nacos集群与单机，且增加了采用内网地址/公网地址注册的配置
+ 以上4项更新，配置文件格式也有修改，详见上面配置说明

### 修改默认数据库检查时间，默认为5分钟一次

在`conf.go`文件头部修改常量

```go
const AUTO_CHECK_MINUTES = 5	//自动检查连接间隔时间，单位为分钟
```

### 在应用中读取主配置文件内容

```go
    //读取字符串配置
    mgconfig.GetConfigString("path.img.temp")
    //读取整数配置
    mgconfig.GetConfigInt("online.connection.max")
```
