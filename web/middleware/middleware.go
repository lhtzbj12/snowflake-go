package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func GlobalMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		fmt.Println("全局中间件开始执行了")
		// 设置变量到Context的key中，可以通过Get()取
		c.Set("request", "***我是全局中间件***")
		// 执行函数
		c.Next()
		v, exsist := c.Get("request")
		fmt.Printf("request value: %s %v\n", v, exsist)
		status := c.Writer.Status()
		fmt.Println("全局中间件执行完毕", status)
		t2 := time.Since(t)
		fmt.Println("time:", t2)
	}
}

// 定义中间
func TimeUsedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		fmt.Println("局部中间件开始执行了")
		// 设置变量到Context的key中，可以通过Get()取
		c.Set("request", "***我是局部中间件***")
		// 执行函数
		c.Next()
		v, exsist := c.Get("request")
		fmt.Printf("request value: %s %v\n", v, exsist)
		// 中间件执行完后续的一些事情
		status := c.Writer.Status()
		fmt.Println("局部中间件执行完毕", status)
		t2 := time.Since(t)
		fmt.Println("time:", t2)
	}
}
