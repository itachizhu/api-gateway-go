package domain

import (
	"net/http"
	"encoding/json"
	"strings"
	"compress/gzip"
	"io/ioutil"
	"github.com/itachizhu/api-gateway-go/util"
	"io"
	"log"
	"net/url"
	"strconv"
)

type ProxyResponseBuilder interface {
	Build() (code int, headers map[string][]string, body []byte)
	makeResponseHeaders() map[string][]string
	makeResponseBody() []byte
}

type BaseProxyResponseBuilder struct {
	child ProxyResponseBuilder
	needHeaders bool
	fileName string
	response *http.Response
}

type TextProxyResponseBuilder struct {
	*BaseProxyResponseBuilder
}

type ImageProxyResponseBuilder struct {
	*BaseProxyResponseBuilder
}

type DownstreamProxyResponseBuilder struct {
	*BaseProxyResponseBuilder
}

type UpstreamProxyResponseBuilder struct {
	*BaseProxyResponseBuilder
}

func (p *BaseProxyResponseBuilder) SetResponse(response *http.Response) *BaseProxyResponseBuilder {
	p.response = response
	return p
}

func (p *BaseProxyResponseBuilder) SetNeedHeaders(needHeaders bool) *BaseProxyResponseBuilder {
	p.needHeaders = needHeaders
	return p
}

func (p *BaseProxyResponseBuilder) SetFileName(fileName string) *BaseProxyResponseBuilder {
	p.fileName = fileName
	return p
}

func (p *BaseProxyResponseBuilder) Build() (code int, headers map[string][]string, body []byte) {
	defer func() {
		if p.response.Body != nil {
			p.response.Body.Close()
			p.response.Body = nil
		}
	}()
	if p.response.StatusCode >= http.StatusBadRequest {
		return p.makeErrorResponse()
	}
	return p.response.StatusCode, p.child.makeResponseHeaders(), p.child.makeResponseBody()
}

func (p *BaseProxyResponseBuilder) makeResponseBody() []byte {
	var reader io.ReadCloser
	var err error
	defer func() {
		if reader != nil {
			reader.Close()
			reader = nil
		}
	}()
	if p.response.Header.Get("Content-Encoding") == "gzip" || strings.Contains(p.response.Header.Get("Content-Type"), "gzip") {
		reader, err = gzip.NewReader(p.response.Body)
		if err != nil {
			log.Printf("创建gzip reader失败: %v", err)
			reader = p.response.Body
		}
	} else {
		reader = p.response.Body
	}
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		util.Panic(3002, "读取response body失败:"+err.Error())
	}
	return body
}

func (p *BaseProxyResponseBuilder) makeResponseHeaders() map[string][]string {
	m := make(map[string][]string)
	if p.needHeaders {
		for key, value := range p.response.Header {
			m[key] = value
		}
	} else {
		for key, value := range p.response.Header {
			if strings.TrimSpace(strings.ToLower(key)) == "content-type" {
				m[key] = value
				break
			}
		}
	}
	return m
}

func (p *BaseProxyResponseBuilder) makeErrorResponse() (code int, headers map[string][]string, body []byte) {
	m := map[string]interface{}{
		"errorCode": 3002,
		"errorMessage": "业务系统服务异常。HttpStatusCode=" + strconv.Itoa(p.response.StatusCode),
	}
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return http.StatusOK, map[string][]string {
		"Content-Type": {"application/json; charset=UTF-8"},
	}, data
}

func (p *DownstreamProxyResponseBuilder) makeResponseHeaders() map[string][]string {
	m := p.BaseProxyResponseBuilder.makeResponseHeaders()
	value, ok := m["Content-Type"]
	if len(p.fileName) == 0 {
		p.fileName = "attachment"
	}
	if !ok || !strings.Contains(strings.ToLower(value[0]), "application/octet-stream") {
		m["Content-Type"]= []string{"application/octet-stream"}
	}
	value, ok = m["Content-Disposition"]
	if !ok || !strings.Contains(strings.ToLower(value[0]), "form-data;") {
		m["Content-Disposition"]= []string{"form-data; name=\"attachment\"; filename=\"" + url.QueryEscape(p.fileName) + "\""}
	}
	return m
}

func CreateResponse(serviceType string) *BaseProxyResponseBuilder {
	responseBuilder := new(BaseProxyResponseBuilder)
	var p ProxyResponseBuilder
	switch serviceType {
	case text:
		p = &TextProxyResponseBuilder{responseBuilder}
		break
	case image:
		p = &ImageProxyResponseBuilder{responseBuilder}
		break
	case downstream:
		p = &DownstreamProxyResponseBuilder{responseBuilder}
		break
	case upstream:
		p = &UpstreamProxyResponseBuilder{responseBuilder}
		break
	default:
		util.Panic(3002, "no support proxy type!")
	}
	responseBuilder.child = p
	return responseBuilder
}