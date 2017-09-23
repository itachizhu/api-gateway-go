package proxy

import (
	"net/http"
	"log"
	"strings"
	"github.com/itachizhu/api-gateway-go/domain"
	"io/ioutil"
	"github.com/itachizhu/api-gateway-go/util"
)

var default404Body = []byte("404 page not found")
var default405Body = []byte("405 method not allowed")

type Engine struct {
	request *http.Request
	writer  http.ResponseWriter
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	engine.request = req
	engine.writer = w

	engine.handleHTTPRequest()
}

func New() *Engine {
	engine := &Engine{}
	return engine
}

func (engine *Engine) handleHTTPRequest() {
	path := engine.request.URL.Path

	if !strings.HasPrefix(path, "/mcloud/mag") {
		engine.writer.WriteHeader(404)
		engine.writer.Write(default404Body)
		return
	}

	paths := strings.Split(strings.TrimSpace(path), "/")

	if len(paths) < 5 {
		engine.writer.WriteHeader(405)
		engine.writer.Write(default405Body)
		return
	}

	defer func() {
		if err := recover(); err != nil {
			engine.writer.WriteHeader(200)
			engine.writer.Header().Set("Content-Type", "application/json;charset=UTF-8")

			switch err.(type) {
			case []byte:
				engine.writer.Write(err.([]byte))
				break
			case string:
				engine.writer.Write([]byte("{\"errorCode\":3002,\"errorMessage\":\"" + err.(string) + "\"}"))
				break
			case error:
				engine.writer.Write([]byte("{\"errorCode\":3002,\"errorMessage\":\"" + err.(error).Error() + "\"}"))
				break
			default:
				engine.writer.Write([]byte("{\"errorCode\":3002,\"errorMessage\":\"未知的错误，系统异常!\"}"))
			}
		}
	}()
	response, needHeaders := domain.Create(paths[3]).Proxy(paths[3], paths[4], engine.request)
	if response.StatusCode >= 400 {
		engine.writer.WriteHeader(200)
		engine.writer.Header().Set("Content-Type", "application/json;charset=UTF-8")
		engine.writer.Write([]byte("{\"errorCode\":3002,\"errorMessage\":\"业务系统服务异常。HttpStatusCode=" + string(response.StatusCode) + "\"}"))
		return
	}
	defer func() {
		if response.Body != nil {
			response.Body.Close()
			response.Body = nil
		}
	}()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		util.Panic(3002, "读取response body失败:"+err.Error())
	}
	engine.writer.WriteHeader(response.StatusCode)
	if needHeaders == 1 {
		for key, value := range response.Header {
			//log.Printf("%v : %v", key, value)
			for _, v := range value {
				engine.writer.Header().Add(key, v)
			}
		}
	} else {
		for key, value := range response.Header {
			//log.Printf("%v : %v", key, value)
			if strings.TrimSpace(strings.ToLower(key)) == "content-type" {
				for _, v := range value {
					engine.writer.Header().Set(key, v)
				}
				break
			}
		}
	}
	engine.writer.Write(body)
}

func (engine *Engine) Run(addr ...string) (err error) {
	defer func() {
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
		}
	}()

	address := resolveAddress(addr)
	log.Printf("[api-gateway] Listening and serving HTTP on %s\n", address)
	err = http.ListenAndServe(address, engine)
	return
}
