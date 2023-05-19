package main

import (
	"Go_distributed_system/log"
	"Go_distributed_system/portal"
	"Go_distributed_system/registry"
	"Go_distributed_system/service"
	"context"
	"fmt"
	stlog "log"
)

func main() {
	err := portal.ImportTemplates()
	if err != nil {
		stlog.Fatal(err)
	}
	host, port := "localhost", "5000"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	r := registry.Registration{
		ServiceName: registry.PortalService,
		ServiceURL:  serviceAddress,
		RequiredServices: []registry.ServiceName{
			registry.LogService,
			registry.GradingService,
		},
		ServiceUpdateURL: serviceAddress + "/services",
		HeartbeatURL:     serviceAddress + "/heartbeat",
	}

	ctx, err := service.Start(context.Background(),
		host,
		port,
		r,
		portal.RegisterHandlers)
	if err != nil {
		stlog.Fatal(err)
	}
	if logProvider, err := registry.GetProvider(registry.LogService); err != nil {
		log.SetClientLogger(logProvider, r.ServiceName)
	}
	<-ctx.Done()
	fmt.Println("Shutting down portal.")
}
