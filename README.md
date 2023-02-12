# Snowflake-Go

#### 介绍
Snowflake-Go使用Golang实现了雪花算法（参考了美团Leaf项目，该项目采用JAVA实现），其解决了时钟回拨问题，基于Gin封装成Rest微服务，使用Nacos作为服务发现。雪花算法需要提供datacenterId和workerId，本项目直接简化成workerId，workerId为范围为0~1023，多副本部署时，需要保证各副本的workerId唯一，否则可能导致id重复（虽然概率很低），支持Zookeeper、环境变量、HostName等多种分配workerId的方式。

#### 安装教程

1. 在Docker运行

   直接通过下面的命令启动一个容器，运行Snowflake-Go

   ```bash
   # 使用环境变量方式提供workerId
   docker run --env "SNOWFLAKE_WORKER_ID=1" --env "DISCOVERY_ENABLED=false" -p 8074:8074 -d registry.cn-beijing.aliyuncs.com/lhtzbj12/snowflake-go
   ```

   使用curl命令请求一下接口

   ```bash
   # 获取1个id
   curl http://localhost:8074/id/get
   # 获取多个id
   curl http://localhost:8074/id/batch?count=100
   ```

   ```bash
   # 使用Zookeeper方式提供workerId
   docker run --env "WOKER_ID_PROVIDER=zookeeper" --env "ZOOKEEPER_CONN_STRING=localhost:2181" --env "DISCOVERY_ENABLED=false" -p 8074:8074 -d registry.cn-beijing.aliyuncs.com/lhtzbj12/snowflake-go
   ```

   ```bash
   # 使用HostName方式提供workerId
   docker run --env "WOKER_ID_PROVIDER=hostname" --hostname "id-gen-1" --env "DISCOVERY_ENABLED=false" -p 8074:8074 -d registry.cn-beijing.aliyuncs.com/lhtzbj12/snowflake-go
   ```

2. Docker-Compose部署

   使用Docker-Compose部署时，需要提供docker-compose.yaml文件，可参考本项目下的文件docker-compose.yaml

   ```bash
   docker-compose -f docker-compose.yaml up
   ```

   

3. 部署至K8s
   在K8s集群里进行部署时，需要提供配置文件，可参考本项目下的文件k8s-statefulset.yaml。采用SatefulSet部署，Pod的HostName自动分配为XXX-0、XXX-1等，数字尾辍即可作为workerId。

   ```bash
   kubectl apply -f k8s-statefulset.yaml
   ```

   

#### 环境变量说明

| 变量名称                      | 默认值         | 说明                                                         |
| ----------------------------- | -------------- | ------------------------------------------------------------ |
| SERVER_PORT                   | 8074           | Gin服务启动后监听的端口                                      |
| DISCOVERY_MICROSRV_NAME       | id-generator   | 微服务名称，用于服务发现、Zookeeper里创建节点等              |
| DISCOVERY_ENABLED             | true           | 是否启用服务发现，即是否注册到Nacos（注册中心）里，提供微服务 |
| DISCOVERY_SRV_ADDR            | localhost:8848 | Nacos服务地址                                                |
| DISCOVERY_NAMESPACE           | public         | Nacos中的命名空间                                            |
| DISCOVERY_MICROSRV_HOST       |                | 应用启动时，往注册中心注册时，使用的IP                       |
| DISCOVERY_MICROSRV_PORT       | -1             | 应用启动时，往注册中心注册时，使用的端口，-1时将取 SERVER_PORT |
| WOKER_ID_PROVIDER             | envirnment     | 工作节点ID分配方式，值可以为  hostname  envirnment   zookeeper，如果为hostname，则要求服务器hostName类似 XXXX-1，XXXX-2等，后面的数字就是workerId，建议在k8s里使用StatefulSet部署 |
| SNOWFLAKE_WORKER_ID           |                | 如果WOKER_ID_PROVIDER值为envirnment，可通过本环境变量设置work |
| ZOOKEEPER_CONN_STRING         | localhost:2181 | 如果WOKER_ID_PROVIDER值为zookeeper，可通过本环境变量设置Zookeeper连接字符串 |
| DISCOVERY_MICROSRV_HEALTH_URL | /health        | 健康检查地址，检查通过，才会往注册中心发出注册的请求         |


#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request