package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const ServerPort = ":3000"
const ServicesURL = "http://localhost" + ServerPort + "/services"

type registry struct { //小写，是私有的，记录当前所有已注册的服务
	registrations []Registration //registration.go文件中的结构
	mutex         *sync.RWMutex  //用来锁上面这个资源防止同时访问
}

func (r *registry) add(reg Registration) error { //registry类的方法，作用是添加reg到slice中
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg) //先添加这个服务到注册表中
	r.mutex.Unlock()
	err := r.sendRequiredServices(reg) //发送所要依赖的请求，查看patch中是否存在并更新provider
	r.notify(patch{                    //add服务的时候通知服务器更新provider
		Added: []patchEntry{
			patchEntry{
				Name: reg.ServiceName,
				URL:  reg.ServiceURL,
			},
		},
	})
	return err
}

func (r registry) notify(fullPatch patch) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for _, reg := range r.registrations { //从已注册的服务里挑选
		go func(reg Registration) {
			for _, reqService := range reg.RequiredServices { //遍历所有需要的服务
				p := patch{Added: []patchEntry{}, Removed: []patchEntry{}} //创建新的patch
				sendUpdate := false
				for _, added := range fullPatch.Added {
					if added.Name == reqService {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}
				for _, removed := range fullPatch.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
				if sendUpdate {
					err := r.sendPatch(p, reg.ServiceUpdateURL) //更新目前拥有的patch服务
					if err != nil {
						log.Println(err)
						return
					}
				}

			}
		}(reg)
	}
}

func (r registry) sendRequiredServices(reg Registration) error {
	r.mutex.RLock() //只读的锁
	defer r.mutex.RUnlock()

	var p patch
	for _, serviceReg := range r.registrations { //循环已经注册的服务
		for _, reqService := range reg.RequiredServices { //循环当前已注册服务所依赖的服务
			if serviceReg.ServiceName == reqService { //如果相同
				p.Added = append(p.Added, patchEntry{ //添加到patch中
					Name: serviceReg.ServiceName,
					URL:  serviceReg.ServiceURL,
				})
			}
		}
	}
	err := r.sendPatch(p, reg.ServiceUpdateURL)
	if err != nil {
		return err
	}
	return nil
}

func (r registry) sendPatch(p patch, url string) error {
	d, err := json.Marshal(p) //转换为json
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}
	return nil
}

func (r *registry) remove(url string) error {
	for i := range reg.registrations {
		if reg.registrations[i].ServiceURL == url {
			r.notify(patch{
				Removed: []patchEntry{
					{
						Name: r.registrations[i].ServiceName,
						URL:  r.registrations[i].ServiceURL,
					},
				},
			})
			r.mutex.Lock()
			reg.registrations = append(reg.registrations[:i], reg.registrations[i+1:]...)
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("Service at URL %s not found", url)
}

func (r *registry) heartbeat(freq time.Duration) {
	for {
		var wg sync.WaitGroup
		for _, reg := range r.registrations {
			wg.Add(1)
			go func(reg Registration) {
				defer wg.Done()
				success := true
				for attemps := 0; attemps < 3; attemps++ {
					res, err := http.Get(reg.HeartbeatURL)
					if err != nil {
						log.Println(err)
					} else if res.StatusCode == http.StatusOK {
						log.Printf("Heartbeat check passed for %v", reg.ServiceName)
						if !success { //这是如果上一次false这次成功了，就把服务重新加回来
							r.add(reg)
						}
						break
					}
					log.Printf("Heartbeat check failed for %v", reg.ServiceName)
					if success {
						success = false
						r.remove(reg.ServiceURL)
					}
					time.Sleep(1 * time.Second)
				}
			}(reg)
			wg.Wait()
			time.Sleep(freq)
		}
	}
} //心跳检测

var once sync.Once //只会运行一次，无论被调用多少次

func SetupRegistryService() {
	once.Do(func() {
		go reg.heartbeat(3 * time.Second)
	})
}

var reg = registry{ //初始化一个registry结构
	registrations: make([]Registration, 0),
	mutex:         new(sync.RWMutex),
}

type RegistryService struct{} //空类，只是用于实现方法，用于注册注册服务的handler

func (s RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) { //实现了这个函数就成为了Handle函数的第二个接口了
	log.Println("Request received")
	switch r.Method {
	case http.MethodPost:
		dec := json.NewDecoder(r.Body) //解析报文结构体
		var r Registration
		err := dec.Decode(&r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Adding service: %v with URL: %s\n", r.ServiceName,
			r.ServiceURL)
		err = reg.add(r) //将本次服务注册到管理中
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		url := string(payload)
		log.Printf("Removing service at URL: %s", url)
		err = reg.remove(url)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
