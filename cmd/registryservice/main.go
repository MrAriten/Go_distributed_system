package main

import (
	"Go_distributed_system/registry"
	"context"
	"fmt"
	"log"
	"net/http"
)

func main() {
	registry.SetupRegistryService() //启动心跳检测

	//下面这个handle是这里最重要的，写明了注册功能是如何处理请求的
	http.Handle("/services", &registry.RegistryService{}) //RegistryService{}实现了第二个参数的接口功能//将注册服务函数注册到HTTP处理服务上

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var srv http.Server
	srv.Addr = registry.ServerPort

	go func() {
		log.Println(srv.ListenAndServe()) //阻塞监听有无注册事件，并将error输出到标准输出
		cancel()                          //如果报错则关闭
	}()

	go func() {
		fmt.Println("Registry service started. Press any key to stop.")
		var s string
		fmt.Scanln(&s) //阻塞等待输入
		srv.Shutdown(ctx)
		cancel()
	}()

	<-ctx.Done()
	fmt.Println("Shutting down registry service")
}
