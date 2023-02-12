package shell_gen

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GenOnlineShell(param GenOnOfflineShellParma) {
	var namespaceId string
	if v1, ok := param.ExtParam["namespaceId"]; ok {
		namespaceId = v1
	}
	var builder strings.Builder
	// 这里不能使用DELETE，即使删除了，由于仍有心跳存在，会重新注册上
	builder.WriteString("curl -X PUT '")
	putUrl := fmt.Sprintf("%s://%s/nacos/v1/ns/instance?serviceName=%s&ip=%s&port=%s&namespaceId=%s&enabled=true'",
		param.Scheme, param.RegCenterHost, param.ServiceName, param.IP, param.Port, namespaceId)
	builder.WriteString(putUrl)
	var filePath = filepath.Join(getCurrentPath(), "offline.sh")
	SaveShell(filePath, builder.String())
}

func GenOfflineShell(param GenOnOfflineShellParma) {
	var namespaceId string
	if v1, ok := param.ExtParam["namespaceId"]; ok {
		namespaceId = v1
	}
	var builder strings.Builder
	// 这里不能使用DELETE，即使删除了，由于仍有心跳存在，会重新注册上
	builder.WriteString("curl -X PUT '")
	putUrl := fmt.Sprintf("%s://%s/nacos/v1/ns/instance?serviceName=%s&ip=%s&port=%s&namespaceId=%s&enabled=false'",
		param.Scheme, param.RegCenterHost, param.ServiceName, param.IP, param.Port, namespaceId)
	builder.WriteString(putUrl)
	var filePath = filepath.Join(getCurrentPath(), "offline.sh")
	SaveShell(filePath, builder.String())
}

func SaveShell(path, content string) {
	var builder strings.Builder
	builder.WriteString("#!/bin/sh -l	\n")
	builder.WriteString(content)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString(builder.String())
	f.Sync()
	fmt.Printf("Save %s success. Content  is\n%s\n", path, builder.String())
}

func getCurrentPath() string {
	if ex, err := os.Executable(); err == nil {
		return filepath.Dir(ex)
	}
	return "./"
}
