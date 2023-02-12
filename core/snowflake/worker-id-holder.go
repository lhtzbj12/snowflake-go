package snowflake

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sfgo/common/fileutil"
	"strconv"
)

/*
 * WorkerId 保持器。即，只要成功从WorkerIdProvider获取一次ID，就将id保存至本地文件，下次启动时，如果从WorkerIdProvider获取失败，则会读取本地文件
 */

var propPath = filepath.Join(os.TempDir(), "snowflake-go", "%s", "conf", "%s", "workerId.properties")

type WorkerIdHolder struct {
	ip               string
	port             string
	appName          string
	localPropPath    string
	workerIdProvider WorkerIdProvider
}

// NewWorkerIdHolder 创建WorkerId保持器
//
// ip 当前应用监听的ip
//
// port 当前应用监听的port
//
// appName 应用名称，用于区分不同应用
func newWorkerIdHolder(ip, port, appName string, workerIdProvider WorkerIdProvider) *WorkerIdHolder {
	return &WorkerIdHolder{
		ip:               ip,
		port:             port,
		appName:          appName,
		localPropPath:    fmt.Sprintf(propPath, appName, port),
		workerIdProvider: workerIdProvider,
	}
}

func (wih *WorkerIdHolder) GetWorkerId() (int64, error) {
	// 从workerIdProvider获取失败
	err := wih.workerIdProvider.Init(wih.ip, wih.port, wih.appName)
	if err != nil {
		log.Printf("workerIdProvider init failed. %s", err.Error())
		return wih.getWorkerIdLocal()
	}
	workerId, err := wih.workerIdProvider.GetWorkerId()
	if err == nil {
		// 获取成功，则保存到本地
		wih.saveWorkerIdLocal(strconv.FormatInt(workerId, 10))
		return workerId, nil
	} else {
		// 获取失败，则尝试从本地获取
		log.Println("get workerId from workerIdProvider failed, try local...")
		return wih.getWorkerIdLocal()
	}
}

func (wih *WorkerIdHolder) saveWorkerIdLocal(workerId string) {
	os.MkdirAll(filepath.Dir(wih.localPropPath), 0666)
	_, err := os.Create(wih.localPropPath)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.WriteFile(wih.localPropPath, []byte(workerId), 0666)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("save workerId %s to local file %s", workerId, wih.localPropPath)
}

func (wih *WorkerIdHolder) getWorkerIdLocal() (int64, error) {
	if !fileutil.Exists(wih.localPropPath) {
		return 0, fmt.Errorf("the prop file doesn't exists, %s", wih.localPropPath)
	}
	b, err := os.ReadFile(wih.localPropPath)
	if err != nil {
		log.Fatalln(err)
	}
	if len(b) != 0 {
		id, err := strconv.ParseInt(string(b), 10, 64)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("get workerId via local file %s\n", wih.localPropPath)
		return id, nil
	} else {
		return 0, errors.New("the worker id in local file is empty")
	}
}
