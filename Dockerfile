FROM registry.cn-beijing.aliyuncs.com/lhtzbj12/alpine-tzsh:3.15.0
EXPOSE 8074

WORKDIR /app
COPY ./sfgo /app/

#启动命令以空格为分隔符拆成数组
CMD ./sfgo
