package domain

import (
	"net/http"
	"github.com/itachizhu/api-gateway-go/util"
	"io"
	"mime/multipart"
	"bytes"
	"strings"
	"io/ioutil"
)

type ProxyRequestBuilder interface {
	makeRequestHeaders(request *http.Request, contentType string)
	makeRequestBody() (io.Reader, string)
}

type BaseProxyRequestBuilder struct {
	child ProxyRequestBuilder
	request *http.Request
	uri string
}

type TextProxyRequestBuilder struct {
	*BaseProxyRequestBuilder
}

type ImageProxyRequestBuilder struct {
	*BaseProxyRequestBuilder
}

type DownstreamProxyRequestBuilder struct {
	*BaseProxyRequestBuilder
}

type UpstreamProxyRequestBuilder struct {
	*BaseProxyRequestBuilder
}

func (p *BaseProxyRequestBuilder) SetUri(uri string) *BaseProxyRequestBuilder {
	p.uri = uri
	return p
}

func (p *BaseProxyRequestBuilder) SetRequest(request *http.Request) *BaseProxyRequestBuilder {
	p.request = request
	return p
}

func (p *BaseProxyRequestBuilder) Build() *http.Request {
	body, contentType := p.child.makeRequestBody()
	request, err := http.NewRequest(p.request.Method, p.uri, body)
	if err != nil {
		util.Panic(3002, "创建http request失败:"+err.Error())
	}
	p.child.makeRequestHeaders(request, contentType)
	return request
}

func (p *BaseProxyRequestBuilder) makeRequestHeaders(request *http.Request, contentType string) {
	if p.request.Header != nil && len(p.request.Header) > 0 {
		for key, values := range p.request.Header {
			for _, value := range values {
				if !isHopByHopHeader(key) {
					request.Header.Add(key, value)
				}
			}
		}
	}
}

func (p *BaseProxyRequestBuilder) makeRequestBody() (io.Reader, string) {
	//defer p.request.Body.Close()
	body, err := ioutil.ReadAll(p.request.Body)
	if err == nil && body != nil && len(body) > 0 {
		return bytes.NewReader(body), ""
	}
	return nil, ""
}

func (p *TextProxyRequestBuilder) makeRequestBody() (io.Reader, string) {
	if strings.Contains(strings.ToLower(p.request.Header.Get("Content-Type")), "application/x-www-form-urlencoded") {
		p.request.ParseForm()
		if len(p.request.PostForm) > 0 {
			return strings.NewReader(strings.TrimSpace(p.request.PostForm.Encode())), ""
		}
	}
	return p.BaseProxyRequestBuilder.makeRequestBody()
}

func (p *UpstreamProxyRequestBuilder) makeRequestHeaders(request *http.Request, contentType string) {
	p.BaseProxyRequestBuilder.makeRequestHeaders(request, contentType)
	if len(contentType) > 0 {
		request.Header.Set("Content-Type", contentType)
	}
}

func (p *UpstreamProxyRequestBuilder) makeRequestBody() (io.Reader, string) {
	p.request.ParseMultipartForm(10 << 20)
	if len(p.request.MultipartForm.File)+len(p.request.MultipartForm.Value) > 0 {
		buffer := new(bytes.Buffer)
		w := multipart.NewWriter(buffer)
		for key, files := range p.request.MultipartForm.File {
			for _, file := range files {
				part, err := w.CreateFormFile(key, file.Filename)
				if err != nil {
					util.Panic(3002, "文件part创建失败："+err.Error())
				}
				f, err := file.Open()
				if err != nil {
					util.Panic(3002, "用户上传的文件打开失败："+err.Error())
				}
				_, err = io.Copy(part, f)
				if err != nil {
					util.Panic(3002, "文件转字节传递异常："+err.Error())
				}
				f.Close()
			}
		}
		for key, values := range p.request.MultipartForm.Value {
			for _, value := range values {
				w.CreateFormField(key)
				w.WriteField(key, value)
			}
		}
		err := w.Close()
		if err != nil {
			util.Panic(3002, "关闭multipart流失败："+err.Error())
		}
		return buffer, w.FormDataContentType()
	}
	return nil, ""
}

func CreateRequest(serviceType string) *BaseProxyRequestBuilder {
	requestBuilder := new(BaseProxyRequestBuilder)
	var p ProxyRequestBuilder
	switch serviceType {
	case text:
		p = &TextProxyRequestBuilder{requestBuilder}
		break
	case image:
		p = &ImageProxyRequestBuilder{requestBuilder}
		break
	case downstream:
		p = &DownstreamProxyRequestBuilder{requestBuilder}
		break
	case upstream:
		p = &UpstreamProxyRequestBuilder{requestBuilder}
		break
	default:
		util.Panic(3002, "no support proxy type!")
	}
	requestBuilder.child = p
	return requestBuilder
}