package domain

import (
	"net/http"
	"strings"
	"log"
	"github.com/itachizhu/api-gateway-go/repository"
	_ "github.com/itachizhu/api-gateway-go/repository"
	"github.com/itachizhu/api-gateway-go/util"
	"github.com/itachizhu/api-gateway-go/model"
	"io"
	"io/ioutil"
	"bytes"
	"mime/multipart"
	"os"
)

const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodHead    = "HEAD"
	MethodPatch   = "PATCH"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

type ProxyService interface {
	verifyRequest(r *http.Request)
	makeRequestBody(r *http.Request) io.Reader
	Proxy(proxyType string, appName string, r *http.Request) (*http.Response, int32)
}

type BaseProxyService struct {
	ProxyService
	free    bool
	cache   bool
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

func (p *BaseProxyService) Proxy(proxyType string, appName string, r *http.Request) (*http.Response, int32) {
	p.ProxyService.verifyRequest(r)
	p.formatUri(appName, strings.Replace(r.URL.Path, "/mcloud/mag/"+proxyType, "", 1))
	request := p.makeRequest(r)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		util.Panic(3002, "转发请求失败:" + err.Error())
	}
	/*
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		util.Panic(3002, "读取response body失败:" + err.Error())
	}
	*/
	return response, p.service.NeedHeaders
}

func (p *BaseProxyService) makeRequestBody(r *http.Request) io.Reader {
	return nil
}

func (p *BaseProxyService) makeRequest(r *http.Request) *http.Request {
	request, err := http.NewRequest(r.Method, p.service.ServiceUri + formatQueryString(r), p.ProxyService.makeRequestBody(r))
	if err != nil {
		util.Panic(3002, "创建http request失败:" + err.Error())
	}
	request.Header = r.Header

	request.Header.Del("Content-Length")
	request.Header.Del("Connection")
	request.Header.Del("Origin")
	request.Header.Del("Cookie")
	request.Header.Del("User-Agent")
	for key, value := range request.Header {
		log.Printf("%v : %v", key, value)
	}

	return request
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
	log.Printf("BaseProxyService")
}

func (p *TextProxyService) verifyRequest(r *http.Request) {
	//panic("我故意的")
	log.Printf("TextProxyService")
}

func (p *TextProxyService) makeRequestBody(r *http.Request) io.Reader {
	log.Printf("我进来了")
	if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/x-www-form-urlencoded") {
		r.ParseForm()
		if len(r.PostForm) > 0 {
			return strings.NewReader(strings.TrimSpace(r.PostForm.Encode()))
		}
	} else {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		log.Printf("%v", string(body))
		if err == nil && body != nil {
			return bytes.NewReader(body)
		}
	}
	return nil
}

func (p *ImageProxyService) verifyRequest(r *http.Request) {
	log.Printf("ImageProxyService")
	//panic("我故意的")
}

func (p *UpstreamProxyService) makeRequestBody(r *http.Request) io.Reader {
	r.ParseMultipartForm(10<<20)
	if len(r.MultipartForm.File) + len(r.MultipartForm.Value) > 0 {
		buffer := new(bytes.Buffer)
		w := multipart.NewWriter(buffer)
		defer func() {
			err := w.Close()
			if err != nil {
				util.Panic(3002, "关闭multipart流失败：" + err.Error())
			}
		}()

		file, err := os.Open("/Users/itachi/Downloads/apktool.txt")
		if err != nil {
			util.Panic(3002, "用户上传的文件打开失败：" + err.Error())
		}
		defer file.Close()
		part, err := w.CreateFormFile("file", "apktool.txt")
		if err != nil {
			util.Panic(3002, "文件part创建失败：" + err.Error())
		}
		_, err = io.Copy(part, file)
		if err != nil {
			util.Panic(3002, "文件转字节传递异常：" + err.Error())
		}
		/*
		for key, files := range r.MultipartForm.File {
			for _, file := range files {
				part, err := w.CreateFormFile(key, file.Filename)
				if err != nil {
					util.Panic(3002, "文件part创建失败：" + err.Error())
				}
				f, err := file.Open()
				if err != nil {
					util.Panic(3002, "用户上传的文件打开失败：" + err.Error())
				}
				_, err = io.Copy(part, f)
				if err != nil {
					util.Panic(3002, "文件转字节传递异常：" + err.Error())
				}
				f.Close()
			}
		}
		*/

		for key, values := range r.MultipartForm.Value {
			for _, value := range values {
				w.WriteField(key, value)
			}
		}

		return bytes.NewReader(buffer.Bytes())
	}
	return nil
}

func Create(serviceType string) ProxyService {
	free := strings.HasPrefix(serviceType, "Free")
	cache := strings.HasSuffix(serviceType, "WithCache")
	sType := strings.TrimPrefix(strings.TrimPrefix(serviceType, "Free"), "ProxyFor")
	sType = strings.TrimSuffix(sType, "WithCache")

	proxyService := new(BaseProxyService)
	proxyService.free = free
	proxyService.cache = cache

	var p ProxyService

	switch sType {
	case "Text":
		p = &TextProxyService{proxyService}
		break
	case "Image":
		p = &ImageProxyService{proxyService}
		break
	case "Upload":
		p = &UpstreamProxyService{proxyService}
		break
	case "Download":
		p = &DownstreamProxyService{proxyService}
		break
	default:
		util.Panic(3002, "no support proxy type!")
	}

	proxyService.ProxyService = p

	return p
}
