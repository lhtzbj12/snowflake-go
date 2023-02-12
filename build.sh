#! /bin/sh

rm -rf ./sfgo
go mod tidy
go build --tags netgo
if [ $? -ne 0 ]; then
   echo 'go build 失败'
   exit 1
fi
docker build -t registry.cn-beijing.aliyuncs.com/lhtzbj12/snowflake-go .
if [ $? -ne 0 ]; then
   echo 'docker build 失败'
   exit 1
fi
docker login --username=lhtzbj12@163.com registry.cn-beijing.aliyuncs.com
if [ $? -ne 0 ]; then
   echo 'docker login 失败'
   exit 1
fi
docker push registry.cn-beijing.aliyuncs.com/lhtzbj12/snowflake-go