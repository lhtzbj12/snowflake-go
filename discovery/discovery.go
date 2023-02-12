package discovery

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sfgo/common/httputil"
	"sfgo/common/netutil"
	"sfgo/common/tools"
	"sfgo/discovery/shell_gen"
	"syscall"
	"time"

	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// 是否启用服务发现
var enabled = tools.GetEnv("DISCOVERY_ENABLED", "true")

// 往注册中心注册时的命名空间，默认 public
var namespace = tools.GetEnv("DISCOVERY_NAMESPACE", "public")

// 注册中心hosts，多个用 , 分隔
var srvAddr = tools.GetEnv("DISCOVERY_SRV_ADDR", "localhost:8848")

// 微服务ip，默认自动获取，多个IP用 , 分隔，将随机取1个
var microSrvHost = tools.GetEnv("DISCOVERY_MICROSRV_HOST", "")

// 微服务端口，默认为-1，为-1时，将使用 SERVER_PORT
var microSrvHostPort = tools.GetEnv("DISCOVERY_MICROSRV_PORT", "-1")

// 微服务名称，默认 demo-microsrv
var microSrvName = tools.GetEnv("DISCOVERY_MICROSRV_NAME", "idgen-microsrv")

// 微服务健康检查IP，可访问时，才会向Nacos注册
var microSrvHealthCheckHost = tools.GetEnv("DISCOVERY_MICROSRV_HEALTH_HOST", "")

// 微服务健康检查Port，可访问时，才会向Nacos注册
var microSrvHealthCheckPort = tools.GetEnv("DISCOVERY_MICROSRV_HEALTH_PORT", "")

// 微服务健康检查地址，可访问时，才会向Nacos注册
var microSrvHealthCheckUrl = tools.GetEnv("DISCOVERY_MICROSRV_HEALTH_URL", "/health")

// 日志级别
var logLevel = tools.GetEnv("DISCOVERY_LOG_LEVEL", "warn")

type NacosHost struct {
	IpAddr string
	Port   uint64
}

func chkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func paramInit(serverPort string) {
	// 如果微服务host无效，则自动获取
	if microSrvHost == "" {
		microSrvHost = netutil.GetFirstNonLoopbackIP()
	} else {
		// 如果微服务host有效，且有多个，则随机取一个
		hosts := strings.Split(microSrvHost, ",")
		hostsLen := len(hosts)
		rand.Seed(time.Now().UnixNano())
		if hostsLen > 1 {
			microSrvHost = hosts[rand.Intn(hostsLen)]
		}
	}
	if microSrvHostPort == "" || microSrvHostPort == "-1" {
		microSrvHostPort = serverPort
	}
	if microSrvHealthCheckHost == "" {
		microSrvHealthCheckHost = microSrvHost
	}
	if microSrvHealthCheckPort == "" || microSrvHealthCheckPort == "-1" {
		microSrvHealthCheckPort = microSrvHostPort
	}
}

func getNacosHost() []constant.ServerConfig {
	serverConfigs := make([]constant.ServerConfig, 0)
	srvs := strings.Split(srvAddr, ",")
	for _, srv := range srvs {
		ipPort := strings.Split(srv, ":")
		port, _ := strconv.Atoi(ipPort[1])
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr:      ipPort[0],
			ContextPath: "/nacos",
			Port:        uint64(port),
			Scheme:      "http",
		})
	}
	return serverConfigs
}

func healthCheck() {
	// 如果健康检查的地址是k8s里的 nodeIP:nodePort，集群里只要有可用Pod，将立即返回200，这可能造成误解，即当前被探测的应用并没有准备好，就注册了
	// 健康检查地址不为空，因此，要一直等到地址返回200，才进行注册
	if microSrvHealthCheckUrl != "" {
		fullUrl := "http://" + microSrvHealthCheckHost + ":" + microSrvHealthCheckPort + microSrvHealthCheckUrl
		for {
			resp, err := httputil.HttpGet(fullUrl, 5)
			if err != nil {
				log.Printf("health check failed: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Printf("health check failed, http status code %v", resp.StatusCode)
				time.Sleep(time.Second * 5)
				continue
			}
			break
		}

	}
}

func Register() {
	// 健康检查
	healthCheck()
	clientConfig := constant.ClientConfig{
		NamespaceId:         namespace, //we can create multiple clients with different namespaceId to support multiple namespace.When namespace is public, fill in the blank string here.
		TimeoutMs:           10000,
		NotLoadCacheAtStart: true,
		LogLevel:            logLevel,
	}
	serverConfigs := getNacosHost()
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	chkError(err)
	port, err := strconv.Atoi(microSrvHostPort)
	chkError(err)
	registerInstanceParam := vo.RegisterInstanceParam{
		Ip:          microSrvHost,
		Port:        uint64(port),
		ServiceName: microSrvName,
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"preserved.register.source": "microSrvName"},
		ClusterName: "DEFAULT",       // default value is DEFAULT
		GroupName:   "DEFAULT_GROUP", // default value is DEFAULT_GROUP
	}
	success, err := namingClient.RegisterInstance(registerInstanceParam)
	chkError(err)
	if success {
		log.Printf("nacos registry, %v DEFAULT_GROUP %v %v %v register finished", namespace, microSrvName, microSrvHost, microSrvHostPort)
	} else {
		log.Println("register failed")
	}
	// 生成上下线脚本
	param := shell_gen.GenOnOfflineShellParma{
		ServiceName:   microSrvName,
		IP:            microSrvHost,
		Port:          microSrvHostPort,
		Scheme:        "http",
		RegCenterHost: srvAddr,
		ExtParam:      map[string]string{"namespaceId": namespace},
	}
	// 生成上线脚本
	shell_gen.GenOnlineShell(param)
	// 生成下线脚本
	shell_gen.GenOfflineShell(param)
	// 注册关机事件
	DeregisterWhenShutdown(namingClient, registerInstanceParam)
}

// DeregisterWhenShutdown 服务停止时从注册中心反注册
func DeregisterWhenShutdown(namingClient naming_client.INamingClient, regParam vo.RegisterInstanceParam) {
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	success, err := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          regParam.Ip,
		Port:        regParam.Port,
		Cluster:     regParam.ClusterName,
		ServiceName: regParam.ServiceName,
		GroupName:   regParam.GroupName,
		Ephemeral:   regParam.Ephemeral,
	})
	chkError(err)
	if success {
		log.Printf("nacos registry, %v DEFAULT_GROUP %v %v %v deregister finished", namespace, microSrvName, microSrvHost, microSrvHostPort)
	} else {
		log.Println("deregister failed")
	}
}

func AutoRegister(serverPort string) {
	if enabled != "true" {
		log.Println("discovery is disabled")
		return
	}
	log.Println("discovery is enabled")
	paramInit(serverPort)
	go Register()
}
