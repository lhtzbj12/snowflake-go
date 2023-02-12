package main

import (
	"log"
	"sfgo/common/tools"
	"sfgo/discovery"
	"sfgo/web"
)

var port = tools.GetEnv("SERVER_PORT", "8074")

func main() {
	log.Println("server start.")
	discovery.AutoRegister(port)
	web.Run("", port)
}
