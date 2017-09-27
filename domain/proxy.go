package domain

import (
	"net/http"
	"strings"
	"log"
	"github.com/itachizhu/api-gateway-go/repository"
	_ "github.com/itachizhu/api-gateway-go/repository"
	"github.com/itachizhu/api-gateway-go/util"
	"github.com/itachizhu/api-gateway-go/model"
)

const (
	text = "Text"
	image = "Image"
	upstream = "Upload"
	downstream = "Download"
)

type ProxyService interface {
	verifyRequest(r *http.Request)
	Proxy(proxyType string, appName string, r *http.Request) (code int, headers map[string][]string, body []byte)
}

type BaseProxyService struct {
	ProxyService
	free    bool
	cache   bool
	serviceType string
	service *model.Service
}

type TextProxyService struct {
	*BaseProxyService
}

type ImageProxyService struct {
	*BaseProxyService
}

type DownstreamProxyService struct {
	*BaseProxyService
}

type UpstreamProxyService struct {
	*BaseProxyService
}

func (p *BaseProxyService) Proxy(proxyType string, appName string, r *http.Request) (code int, headers map[string][]string, body []byte) {
	p.ProxyService.verifyRequest(r)
	p.formatUri(appName, strings.Replace(r.URL.Path, "/mcloud/mag/"+proxyType, "", 1))
	p.verify(r)
	if p.cache {
		_code, _headers, _body := p.fromCache()
		if _code > 0 && _body != nil {
			return _code, _headers, _body
		}
	}
	request := p.makeRequest(r)
	//defer request.Body.Close()
	defer func() {
		if request != nil && request.Body != nil {
			request.Body.Close()
		}
	}()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		util.Panic(3002, "转发请求失败:"+err.Error())
	}
	code, headers, body = CreateResponse(p.serviceType).SetResponse(response).SetNeedHeaders(1 == p.service.NeedHeaders).Build()
	if p.cache {
		cache := NewCache(p.serviceType, code, headers, body, p.service)
		if cache != nil {
			go cache.Cache()
		}
	}
	return
}

func (p *BaseProxyService) makeRequest(r *http.Request) *http.Request {
	return CreateRequest(p.serviceType).SetRequest(r).SetUri(p.service.ServiceUri+formatQueryString(r)).Build()
}

func formatQueryString(r *http.Request) string {
	if len(r.URL.RawQuery) == 0 {
		return ""
	} else {
		return "?" + r.URL.RawQuery
	}
}

func (p *BaseProxyService) formatUri(appName string, uri string) {
	if len(appName) == 0 {
		util.Panic(1010, "服务应用名称为空。")
	}
	path := uri[strings.Index(uri, appName)+len(appName):]
	p.service = repository.NewServiceRepository().FindService(appName)
	if p.service == nil || p.service.Id < 1 {
		util.Panic(1003, "服务应用未注册。应用名=["+appName+"]")
	}
	p.service.ServiceUri = strings.TrimSuffix(strings.TrimSuffix(p.service.Url, "/")+"/"+strings.TrimPrefix(path, "/"), "/")
}

func (p *BaseProxyService) verifyRequest(r *http.Request) {
	log.Printf("BaseProxyService.verifyRequest")
}

func (p *BaseProxyService) verify(r *http.Request) {
	log.Printf("BaseProxyService.verify")
}

func (p *BaseProxyService) fromCache() (code int, headers map[string][]string, body []byte) {
	log.Printf("BaseProxyService.fromCache")
	cache := NewCache(p.serviceType, 0, nil, nil, p.service)
	if cache != nil {
		return cache.FromCache()
	}
	return 0, nil, nil
}

func (p *ImageProxyService) verifyRequest(r *http.Request) {
	log.Printf("ImageProxyService")
	if strings.ToUpper(r.Method) != strings.ToUpper(http.MethodGet) {
		util.Panic(1015, "目前此类请求不支持此种method。")
	}
}

func (p *DownstreamProxyService) verifyRequest(r *http.Request) {
	log.Printf("DownstreamProxyService")
	if strings.ToUpper(r.Method) != strings.ToUpper(http.MethodGet) {
		util.Panic(1015, "目前此类请求不支持此种method。")
	}
}

func (p *UpstreamProxyService) verifyRequest(r *http.Request) {
	log.Printf("UpstreamProxyService")
	//panic("我故意的")
	if strings.ToUpper(r.Method) != strings.ToUpper(http.MethodPost) && strings.ToUpper(r.Method) != strings.ToUpper(http.MethodPut) {
		util.Panic(1015, "目前此类请求不支持此种method。")
	}
	if !strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "multipart/form-data") {
		util.Panic(1016, "上传文件的请求content-type必须是multipart/form-data。")
	}
}

func Create(serviceType string) ProxyService {
	free := strings.HasPrefix(serviceType, "Free")
	cache := strings.HasSuffix(serviceType, "WithCache")
	sType := strings.TrimPrefix(strings.TrimPrefix(serviceType, "Free"), "ProxyFor")
	sType = strings.TrimSuffix(sType, "WithCache")

	proxyService := new(BaseProxyService)
	proxyService.free = free
	proxyService.cache = cache
	proxyService.serviceType = sType

	var p ProxyService
	switch sType {
	case text:
		p = &TextProxyService{proxyService}
		break
	case image:
		p = &ImageProxyService{proxyService}
		break
	case upstream:
		p = &UpstreamProxyService{proxyService}
		break
	case downstream:
		p = &DownstreamProxyService{proxyService}
		break
	default:
		util.Panic(3002, "no support proxy type!")
	}
	proxyService.ProxyService = p
	return p
}
