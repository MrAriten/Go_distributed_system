package registry

type Registration struct {
	ServiceName      ServiceName
	ServiceURL       string
	RequiredServices []ServiceName //该服务所依赖的其他服务
	ServiceUpdateURL string        //最新服务的地址，是和registry沟通来获得的
	HeartbeatURL     string
}

type ServiceName string

const (
	LogService     = ServiceName("LogService")
	GradingService = ServiceName("GradingService")
	PortalService  = ServiceName("Portald")
)

type patchEntry struct { //更新通知包
	Name ServiceName
	URL  string
}

type patch struct { //同时更新的所有条目
	Added   []patchEntry
	Removed []patchEntry
}
