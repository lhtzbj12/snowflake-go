package snowflake

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrInitExpected = errors.New("IdGenerator: must be initialized before using")
)

// singleton
var sl *Snowflake
var once sync.Once

// IdGenerator 基本雪花算法的id生成器
type IdGenerator struct {
	ip               string
	port             string
	appName          string
	workerIdProvider WorkerIdProvider
	initFlag         bool
}

// NewIdGenerator 创建IdGenerator
//
// ip 当前应用监听的ip
//
// port 当前应用监听的port
//
// appName 应用名称，用于区分不同应用
//
// WorkerIdProvider 需要提供WorkerIdProvider
func NewIdGenerator(ip, port, appName string) *IdGenerator {
	workerIdProvider := GetWorkerProvider()
	return &IdGenerator{
		ip:               ip,
		port:             port,
		appName:          appName,
		workerIdProvider: workerIdProvider,
		initFlag:         false,
	}
}

// Init 获取workerId，进行初始化
func (sig *IdGenerator) Init() {
	// create holder
	workerIdHolder := newWorkerIdHolder(sig.ip, sig.port, sig.appName, sig.workerIdProvider)
	workerId, err := workerIdHolder.GetWorkerId()
	if err != nil {
		panic("workerId is wrong." + err.Error())
	}
	if workerId < 0 || workerId > MaxWorkerId() {
		panic(fmt.Sprintf("workerId must between 0 and %d", MaxWorkerId()))
	}
	once.Do(func() {
		sl = NewSnowflake(workerId)
	})
	sig.initFlag = true
}

func (sig *IdGenerator) GetId() (int64, error) {
	if !sig.initFlag {
		return 0, ErrInitExpected
	}
	return sl.GetId()
}

func (sig *IdGenerator) GetIds(n int) ([]int64, error) {
	if !sig.initFlag {
		return nil, ErrInitExpected
	}
	result := make([]int64, 0, n)
	if n <= 1 {
		n = 1
	}
	for i := 0; i < n; i++ {
		v, err := sl.GetId()
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}
