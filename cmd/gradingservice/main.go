package main

import (
	"Go_distributed_system/grades"
	"Go_distributed_system/log"
	"Go_distributed_system/registry"
	"Go_distributed_system/service"
	"context"
	"fmt"
	stlog "log"
)

func main() {
	host, port := "localhost", "6000"
	serviceAddress := fmt.Sprintf("http://%v:%v", host, port)

	r := registry.Registration{
		ServiceName:      registry.GradingService,
		ServiceURL:       serviceAddress,
		RequiredServices: []registry.ServiceName{registry.LogService},
		ServiceUpdateURL: serviceAddress + "/services",
		HeartbeatURL:     serviceAddress + "/heartbeat", //用于通知service的状态，否则宕机的时候不知道状况
	}
	ctx, err := service.Start(context.Background(),
		host,
		port,
		r,
		grades.RegisterHandlers)
	if err != nil {
		stlog.Fatal(err)
	}

	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		fmt.Printf("Logging service found at: %s\n", logProvider)
		log.SetClientLogger(logProvider, r.ServiceName)
	}

	<-ctx.Done() //阻塞等待
	fmt.Println("Shutting down grading service")
}
