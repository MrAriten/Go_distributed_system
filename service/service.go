package service

import (
	"Go_distributed_system/registry"
	"context"
	"fmt"
	"log"
	"net/http"
)

// Go语言的context库是Go标准库中的一部分，
// 用于在并发和协程之间传递上下文信息、控制协程的取消操作以及处理超时。
// context库提供了Context类型，该类型用于创建和传递上下文。
// 这个函数实现的功能是对每个service的handler进行初始化，并对服务器发送注册请求
func Start(ctx context.Context, host, port string,
	reg registry.Registration,
	registerHandlersFunc func()) (context.Context, error) {

	registerHandlersFunc()                               //这里是log函数的服务注册，用于注册HTTP的handler
	ctx = startService(ctx, reg.ServiceName, host, port) //服务开始
	err := registry.RegisterService(reg)                 //进行注册，发送POST请求
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

// cancel函数是context库中的一个重要函数，用于取消上下文及其关联的协程或操作。
// 当调用cancel函数时，会触发上下文的取消信号，通知相关的协程或操作停止执行并进行清理操作。
// 取消上下文的主要作用是在协程或操作不再需要执行时，通过通知相关的协程进行取消，以释放资源、停止计算或回滚操作。
// 取消操作可以避免长时间运行的协程或操作导致资源泄漏、长时间占用系统资源或执行不必要的工作。
func startService(ctx context.Context, serviceName registry.ServiceName,
	host, port string) context.Context {

	ctx, cancel := context.WithCancel(ctx) //创建一个新的上下文（context.Context）以及一个取消函数（cancel function）

	var srv http.Server
	srv.Addr = ":" + port

	go func() {
		log.Println(srv.ListenAndServe()) //这里也会阻塞等待，如果出错了便会进行到下面
		err := registry.ShutdownService(fmt.Sprintf("http://%s:%s", host, port))
		if err != nil {
			log.Println(err)
		}
		cancel() //取消上下文及其关联的协程操作
	}()

	go func() {
		fmt.Printf("%v started. Press any key to stop. \n", serviceName)
		var s string
		fmt.Scanln(&s) //阻塞等待输入，这里的功能就是输入任何键就会导致服务关停
		err := registry.ShutdownService(fmt.Sprintf("http://%s:%s", host, port))
		if err != nil {
			log.Println(err)
		}
		srv.Shutdown(ctx)
		cancel()
	}()

	return ctx
}
