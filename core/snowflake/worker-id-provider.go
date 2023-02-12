package snowflake

import "sfgo/common/tools"

const (
	PROVIDER_HOSTNAME   = "hostname"
	PROVIDER_ENVIRNMENT = "envirnment"
	PROVIDER_ZOOKEEPER  = "zookeeper"
)

// WOKER_ID_PROVIDER 工作节点ID提供者可以为  hostName envirnment zookeeper
//
// 默认值为 envirnment
//
// 如果WOKER_ID_PROVIDER=hostname，则要求系统hostname格式为 xxxx-1 xxxx-2等，在k8s里采用StatefulSet部署即可
var workerIdProvider = tools.GetEnv("WOKER_ID_PROVIDER", "envirnment")

// 如果WOKER_ID_PROVIDER=envirnment，则需要在系统环境变量里设置下面的环境变量值
var workerIdProviderEnvName = "SNOWFLAKE_WORKER_ID"

// 如果WOKER_ID_PROVIDER=zookeeper，则需要提供Zookeeper的连接字符串
var zkConnString = tools.GetEnv("ZOOKEEPER_CONN_STRING", "localhost:2181")

type WorkerIdProvider interface {
	Init(ip, port, appName string) error
	GetWorkerId() (int64, error)
}

func GetWorkerProvider() WorkerIdProvider {
	var wokerIdProvider WorkerIdProvider
	switch workerIdProvider {
	case PROVIDER_ENVIRNMENT:
		wokerIdProvider = NewEnvWorkerIdProvider(workerIdProviderEnvName)
	case PROVIDER_ZOOKEEPER:
		wokerIdProvider = NewZookeeperWorkerIdProvider(zkConnString)
	default:
		wokerIdProvider = NewHostNameWokerIdProvider()
	}
	return wokerIdProvider
}
