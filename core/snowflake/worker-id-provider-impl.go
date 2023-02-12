package snowflake

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sfgo/common/tools"
	"sfgo/common/valiutil"
	"strconv"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

const rootNodePathTemplate = "/snowflake-go/worker-id-provider/%s"

// PayloadData 保存的的负载
type PayloadData struct {
	IP        string `json:"ip"`
	Port      string `json:"port"`
	Timestamp int64  `json:"timestamp"`
}

// marshalPayloadData 创建负载数据
func marshalPayloadData(ip, port string, timestamp int64) []byte {
	data := PayloadData{
		IP:        ip,
		Port:      port,
		Timestamp: timestamp,
	}
	v, _ := json.Marshal(data)
	return v
}

// unmarshalPayloadData 解析负载数据
func unmarshalPayloadData(value []byte) (*PayloadData, error) {
	var data PayloadData
	err := json.Unmarshal(value, &data)
	if err == nil {
		return &data, nil
	} else {
		return nil, fmt.Errorf("unmarshal playload data fialed. %s", value)
	}
}

// ZookeeperWorkerIdProvider 基于Zookeeper实现
type ZookeeperWorkerIdProvider struct {
	connStr             string
	ip                  string
	port                string
	rootNodePath        string
	workerIdNodeName    string
	workerIdNodePathPre string
	workerId            int64
}

// NewZookeeperWorkerIdProvider 创建ZookeeperWorkerIdProvider
func NewZookeeperWorkerIdProvider(connStr string) *ZookeeperWorkerIdProvider {
	if connStr == "" {
		panic("zookeeper connection string can't be empty")
	}
	return &ZookeeperWorkerIdProvider{
		connStr: connStr,
	}
}

func (zwp *ZookeeperWorkerIdProvider) getConn() *zk.Conn {
	var hosts = strings.Split(zwp.connStr, ",")
	conn, _, err := zk.Connect(hosts, time.Second*5)
	if err != nil {
		fmt.Println(err)
		return nil
	} else {
		return conn
	}
}

// Init 初始化
func (zwp *ZookeeperWorkerIdProvider) Init(ip, port, appName string) error {
	zwp.ip = ip
	zwp.port = port
	// 设置根节点名称
	zwp.rootNodePath = fmt.Sprintf(rootNodePathTemplate, appName)
	// 设置workerId节点名称
	zwp.workerIdNodeName = ip + ":" + port
	// 补全workerId节点路径前辍
	zwp.workerIdNodePathPre = zwp.rootNodePath + "/" + zwp.workerIdNodeName + "-"
	// 给默认值
	zwp.workerId = -1
	conn := zwp.getConn()
	defer conn.Close()
	// 处理根节点
	err := dealRootNode(conn, zwp.rootNodePath)
	if err != nil {
		return err
	}
	// 处理workerId节点
	// 找查已存在的节点
	children, _, err := conn.Children(zwp.rootNodePath)
	if err != nil {
		return fmt.Errorf("get children failed. reason: %s", err.Error())
	}
	var existWorkerId int64 = -1
	existNodePath, exist := "", false
	for _, child := range children {
		nodeKey := strings.Split(child, "-")
		if nodeKey[0] == zwp.workerIdNodeName {
			existNodePath = zwp.rootNodePath + "/" + child
			value, err := strconv.ParseInt(nodeKey[1], 10, 64)
			if err != nil {
				return errors.New("node name unrecognizable")
			}
			existWorkerId = value
			exist = true
			break
		}
	}
	// 找到当前节点
	if exist {
		// 获取节点的数据，判断时间戳
		data, stat, err := conn.Get(existNodePath)
		if err != nil {
			return err
		}
		payloadData, err := unmarshalPayloadData(data)
		curTimestamp := time.Now().UnixMilli()
		if err != nil || payloadData.Timestamp < curTimestamp {
			_, err := conn.Set(existNodePath, marshalPayloadData(ip, port, curTimestamp), stat.Version)
			if err != nil {
				return err
			}
		}
		if payloadData.Timestamp > curTimestamp {
			return fmt.Errorf("init timestamp check error,forever node timestamp gt this node time")
		}
		log.Printf("get workerId via exists workerId node. workerId: %d, path: %s", existWorkerId, existNodePath)
		zwp.workerId = existWorkerId
	} else {
		//控制访问权限模式
		var acl = zk.WorldACL(zk.PermAll)
		// 不存在，则创建
		p, err := conn.Create(zwp.workerIdNodePathPre, marshalPayloadData(ip, port, time.Now().UnixMilli()), 2, acl)
		if err != nil {
			return err
		}
		log.Printf("create workerId node success. path: %s", p)
		// 获取序号
		seq := strings.TrimPrefix(p, zwp.workerIdNodePathPre)
		seqValue, err := strconv.ParseInt(seq, 10, 64)
		if err != nil {
			return errors.New("node name's format unrecognizable")
		}
		log.Printf("get workerId via new workerId node. workerId: %d", seqValue)
		zwp.workerId = seqValue
	}
	return nil
}

// GetWorkerId 获取id
func (zwp *ZookeeperWorkerIdProvider) GetWorkerId() (int64, error) {
	if zwp.workerId < 0 {
		return 0, fmt.Errorf("worker id is wrong. Please check the provider")
	}
	return zwp.workerId, nil
}

func dealRootNode(conn *zk.Conn, rootNodePath string) error {
	exists, _, err := conn.Exists(rootNodePath)
	if err != nil {
		return err
	}
	//控制访问权限模式
	var acl = zk.WorldACL(zk.PermAll)
	// 不存在，则创建节点
	//flags有4种取值：
	//0:永久，除非手动删除
	//zk.FlagEphemeral = 1:短暂，session断开则改节点也被删除
	//zk.FlagSequence  = 2:会自动在节点后面添加序号
	//3:Ephemeral和Sequence，即，短暂且自动添加序号
	if !exists {
		errCreate := createNodeAll(conn, rootNodePath, 0, acl)
		if errCreate != nil {
			return err
		}
		return nil
	} else {
		return nil
	}
}

// 递归创建，存在则跳过
//
// flags有4种取值：
//
// 0:永久，除非手动删除
//
// zk.FlagEphemeral = 1:短暂，session断开则改节点也被删除
//
// zk.FlagSequence  = 2:会自动在节点后面添加序号
//
// 3:Ephemeral和Sequence，即，短暂且自动添加序号
func createNodeAll(conn *zk.Conn, path string, flags int32, acl []zk.ACL) error {
	paths := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(paths) < 1 {
		return fmt.Errorf("path is error, %s", path)
	}
	curPath := ""
	// 是否进行是否存在的检测
	checkExists := true
	nodeExists := false
	var err error
	for i := 0; i < len(paths); i++ {
		curPath += "/" + paths[i]
		if checkExists {
			nodeExists, _, err = conn.Exists(curPath)
			if err != nil {
				return err
			}
		}
		// 如果不存在，则创建
		if !nodeExists {
			// 某父节点一旦不存在，之后，就不进行是否存在的检测
			checkExists = false
			_, errCreate := conn.Create(curPath, nil, flags, acl)
			if errCreate != nil {
				return errCreate
			}
		}
	}
	return nil
}

// EnvWorkerIdProvider 基于环境变量实现
type EnvWorkerIdProvider struct {
	envName  string
	workerId int64
}

// NewEnvWorkerIdProvider 创建EnvWorkerIdProvider
//
// envName 环境变量名称
func NewEnvWorkerIdProvider(envName string) *EnvWorkerIdProvider {
	id := tools.GetEnv(envName, "")
	if id == "" {
		panic(fmt.Sprintf("environment variable %s doesn't exist", envName))
	}
	workerId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("workerId is wrong. environment variable value is %s", id))
	}
	return &EnvWorkerIdProvider{
		envName:  envName,
		workerId: workerId,
	}
}

func (rwp *EnvWorkerIdProvider) Init(ip, port, appName string) error {
	return nil
}

func (rwp *EnvWorkerIdProvider) GetWorkerId() (int64, error) {
	log.Printf("get workerId via environment. evnName: %s workerId: %d", rwp.envName, rwp.workerId)
	return rwp.workerId, nil
}

// HostNameWokerIdProvider 基于hostname实现
//
// 用于k8s里，采用statefulset部署时，获取hostname的序号
type HostNameWokerIdProvider struct {
	hostName string
	workerId int64
}

// NewHostNameWokerIdProvider 创建HostNameWokerIdProvider
func NewHostNameWokerIdProvider() *HostNameWokerIdProvider {
	hostName, err := os.Hostname()
	if err != nil {
		panic(fmt.Sprintf("hostName is wrong. err is %s", err.Error()))
	}
	if hostName == "" {
		panic("hostName is null")
	}
	if !valiutil.Regexp(`.+-\d+`, hostName) {
		panic(fmt.Sprintf(`os hostName is %s. hostname must match .+-\d+ , e.g. id-server-1 order-server-2`, hostName))
	}
	id := hostName[strings.LastIndex(hostName, "-")+1:]
	workerId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("workerId is wrong. hostname is %s", hostName))
	}
	return &HostNameWokerIdProvider{
		hostName: hostName,
		workerId: workerId,
	}
}

func (hwp *HostNameWokerIdProvider) Init(ip, port, appName string) error {
	return nil
}

func (hwp *HostNameWokerIdProvider) GetWorkerId() (int64, error) {
	log.Printf("get workerId via hostname. hostname: %s workerId: %d", hwp.hostName, hwp.workerId)
	return hwp.workerId, nil
}
