package main

import (
	"Go_distributed_system/log"
	"Go_distributed_system/registry"
	"Go_distributed_system/service"
	"context"
	"fmt"
	stlog "log"
)

func main() {
	log.Run("./distributed.log") //log文件的名字
	host, port := "localhost", "4000"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	r := registry.Registration{ //服务注册信息
		ServiceName:      registry.LogService,
		ServiceURL:       serviceAddress,
		RequiredServices: make([]registry.ServiceName, 0),
		ServiceUpdateURL: serviceAddress + "/services",
		HeartbeatURL:     serviceAddress + "/heartbeat",
	}
	ctx, err := service.Start( //发送POST请求注册服务
		context.Background(),
		host,
		port,
		r,
		log.RegisterHandlers,
	)

	if err != nil {
		stlog.Fatalln(err)
	}
	<-ctx.Done() //阻塞操作，直到上下文被取消

	fmt.Println("Shutting down log service.")
}
