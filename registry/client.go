package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
)

// 主要功能是向服务器发送注册请求
// 并设置好该服务的必要handler
func RegisterService(r Registration) error {
	heartbeatURL, err := url.Parse(r.HeartbeatURL)
	if err != nil {
		return err
	}
	http.HandleFunc(heartbeatURL.Path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	serviceUpdateURL, err := url.Parse(r.ServiceUpdateURL)
	if err != nil {
		return err
	}
	http.Handle(serviceUpdateURL.Path, &serviceUpdateHanlder{})

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err = enc.Encode(r) //对注册信息进行json解码，有错误则返回
	if err != nil {
		return err
	}

	res, err := http.Post(ServicesURL, "application/json", buf) //没有错误则发送注册请求
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK { //如果发送不成功
		return fmt.Errorf("Failed to register service. Registry service "+
			"responded with code %v", res.StatusCode)
	}

	return nil
}

type serviceUpdateHanlder struct{}

func (suh serviceUpdateHanlder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	dec := json.NewDecoder(r.Body)
	var p patch
	err := dec.Decode(&p)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Updated received %v\n", p)
	prov.Update(p)
}

func ShutdownService(url string) error { //用于删除注册
	req, err := http.NewRequest(http.MethodDelete, ServicesURL,
		bytes.NewBuffer([]byte(url))) //创建一个DELETE请求，并将其发送到ServicesURL指定的URL，请求体中包含了url字符串
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "text/plain")
	res, err := http.DefaultClient.Do(req) //执行HTTP请求，并将响应结果保存在res变量中。http.DefaultClient是一个默认的HTTP客户端，
	// 它提供了发送HTTP请求的功能。Do()函数接受一个req参数，即前面创建的请求对象，并返回响应结果和可能的错误。
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to deregister service. Registry "+
			"service responded with code %v", res.StatusCode)
	}
	return nil
}

type providers struct { //向外提供服务
	services map[ServiceName][]string
	mutex    *sync.RWMutex
}

func (p *providers) Update(pat patch) { //更新服务的url
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, patchEntry := range pat.Added {
		if _, ok := p.services[patchEntry.Name]; !ok { //查看这个服务是否存在
			p.services[patchEntry.Name] = make([]string, 0)
		}
		p.services[patchEntry.Name] = append(p.services[patchEntry.Name],
			patchEntry.URL)
	}

	for _, patchEntry := range pat.Removed {
		if providerURLs, ok := p.services[patchEntry.Name]; ok {
			for i := range providerURLs {
				if providerURLs[i] == patchEntry.URL {
					p.services[patchEntry.Name] = append(providerURLs[:i],
						providerURLs[i+1:]...)
				}
			}
		}
	}
}

func (p providers) get(name ServiceName) (string, error) { //获取某个服务的url
	providers, ok := p.services[name]
	if !ok {
		return "", fmt.Errorf("No providers available for service %v", name)
	}
	idx := int(rand.Float32() * float32(len(providers)))
	return providers[idx], nil
}

func GetProvider(name ServiceName) (string, error) {
	return prov.get(name)
}

var prov = providers{
	services: make(map[ServiceName][]string),
	mutex:    new(sync.RWMutex),
}
