package web

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sfgo/web/handler/actuator"
	"sfgo/web/handler/id"
	"syscall"
	"time"

	// https://github.com/chenjiandongx/ginprom
	"github.com/chenjiandongx/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func routerInit(r *gin.Engine) {
	r.GET("/health", actuator.Health)
	// 监控
	groupActuator := r.Group("/actuator")
	{
		groupActuator.GET("/health", actuator.Health)
		groupActuator.GET("/metrics", ginprom.PromHandler(promhttp.Handler()))
	}
	// id组
	groupId := r.Group("/id")
	{
		groupId.GET("/get", id.GetOne)
		// 获取多个id
		groupId.GET("/batch", id.GetBatch)
	}
}

func Run(ip, port string) {
	// Web服务初始化
	r := gin.Default()
	// 增加promethues指标导出中间件
	r.Use(ginprom.PromMiddleware(nil))
	//路由初始化
	routerInit(r)
	// 启动
	srv := &http.Server{
		Addr:    ip + ":" + port,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	<-ctx.Done()
	log.Println("Server exiting.")
}
