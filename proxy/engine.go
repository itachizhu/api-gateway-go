package proxy

import (
	"net/http"
	"log"
	"strings"
	"github.com/itachizhu/api-gateway-go/domain"
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
	defer func() {
		if err := recover(); err != nil {
			engine.writer.WriteHeader(http.StatusOK)
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

	path := engine.request.URL.Path
	if !strings.HasPrefix(path, "/mcloud/mag") {
		engine.writer.WriteHeader(http.StatusNotFound)
		engine.writer.Write(default404Body)
		return
	}
	paths := strings.Split(strings.TrimSpace(path), "/")
	if len(paths) < 5 {
		engine.writer.WriteHeader(http.StatusMethodNotAllowed)
		engine.writer.Write(default405Body)
		return
	}
	code, headers, body := domain.Create(paths[3]).Proxy(paths[3], paths[4], engine.request)
	engine.writer.WriteHeader(code)
	for key, value := range headers {
		for _, v := range value {
			if strings.TrimSpace(strings.ToLower(key)) == "content-type" {
				engine.writer.Header().Set(key, v)
			} else {
				engine.writer.Header().Add(key, v)
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
