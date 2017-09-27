package domain

import (
	"github.com/itachizhu/api-gateway-go/model"
	"log"
	"strconv"
	"strings"
	"encoding/json"
	"crypto/sha256"
	"encoding/hex"
	"encoding/base64"
	"path/filepath"
	"time"
	"os"
	"io/ioutil"
)

const (
	defaultTimeout = 1 * 60
	codeKey        = "code"
	headersKey     = "headers"
	bodyKey        = "body"
	cachePrefix    = "api-gateway:"
)

type Cache interface {
	cacheBody(m map[string]interface{})
	bodyFromCache(m map[string]interface{}) []byte
}

type BaseCache struct {
	child   Cache
	code    int
	headers map[string][]string
	body    []byte
	service *model.Service
}

type TextCache struct {
	*BaseCache
}

type ByteCache struct {
	*BaseCache
}

type FileCache struct {
	*BaseCache
}

func NewCache(serviceType string, code int, headers map[string][]string, body []byte, service *model.Service) *BaseCache {
	cache := &BaseCache{
		code:    code,
		headers: headers,
		body:    body,
		service: service,
	}
	var c Cache
	switch serviceType {
	case text:
		c = &TextCache{cache}
		break
	case image:
		c = &ByteCache{cache}
		break
	case downstream:
		c = &FileCache{cache}
		break
	default:
		log.Printf("no support proxy type!")
		return nil
	}
	cache.child = c
	return cache
}

func (p *BaseCache) Cache() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("缓存不成功，过程中产生了错误: %v", err)
		}
	}()
	//log.Printf("测试一下！")
	key := p.cacheKey()
	m := make(map[string]interface{})
	m[codeKey] = p.code
	m[headersKey] = p.headers
	//m[body] = string(p.body)
	p.child.cacheBody(m)
	data, err := json.Marshal(m)
	if err != nil {
		log.Printf("不好意思，json序列化数据出错了: %v", err)
		return
	}
	err = redisCache.Set(key, string(data), defaultTimeout)
	if err != nil {
		log.Printf("不好意思，缓存数据出错了: %v", err)
		return
	}
	err = p.cachePrefix(key)
	if err != nil {
		log.Printf("不好意思，缓存key值出错了: %v", err)
		return
	}
	log.Printf("应用[%v]的路径为[%v]缓存成功！", p.service.AppName, p.service.ServiceUri)
}

func (p *BaseCache) FromCache() (code int, headers map[string][]string, body []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("获取缓存错误: %v", err)
			code = 0
			headers = nil
			body = nil
			return
		}
	}()
	cacheData := get(p.cacheKey())
	if len(cacheData) == 0 {
		return 0, nil, nil
	}
	var data interface{}
	err := json.Unmarshal([]byte(cacheData), &data)
	if err != nil {
		log.Printf("获取缓存错误: %v", err)
		return 0, nil, nil
	}
	m := data.(map[string]interface{})
	code = int(m[codeKey].(float64))
	_headers := m[headersKey].(map[string]interface{})
	headers = make(map[string][]string)
	for key, values := range _headers {
		headerValues := make([]string, 0)
		for _, value := range values.([]interface{}) {
			headerValues = append(headerValues, value.(string))
		}
		headers[key] = headerValues
	}
	body = p.child.bodyFromCache(m)
	log.Printf("此次请求从缓存中获取，路径为: %v", p.service.ServiceUri)
	return
}

func (p *BaseCache) cacheKey() string {
	path := strings.Replace(p.service.ServiceUri, p.service.Url, "", 1)
	pathHash := sha256.Sum256([]byte(path))
	return strconv.FormatInt(p.service.Id, 10) + ":" + hex.EncodeToString([]byte(pathHash[:]))
}

func (p *BaseCache) cachePrefix(key string) error {
	prefix := cachePrefix + strconv.FormatInt(p.service.Id, 10)
	keysString := get(prefix)
	if len(keysString) == 0 {
		keysString = "[]"
	}
	m := make([]string, 0)
	err := json.Unmarshal([]byte(keysString), m)
	if err != nil {
		m = make([]string, 0)
	}
	if !contains(m, key) {
		m = append(m, key)
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return redisCache.Set(prefix, data, 0)
}

func (p *BaseCache) cacheBody(m map[string]interface{}) {
}

func (p *BaseCache) bodyFromCache(m map[string]interface{}) []byte {
	return nil
}

func (p *TextCache) cacheBody(m map[string]interface{}) {
	m[bodyKey] = string(p.body)
}

func (p *TextCache) bodyFromCache(m map[string]interface{}) []byte {
	return []byte(m[bodyKey].(string))
}

func (p *ByteCache) cacheBody(m map[string]interface{}) {
	m[bodyKey] = base64.StdEncoding.EncodeToString(p.body)
}

func (p *ByteCache) bodyFromCache(m map[string]interface{}) []byte {
	value, err := base64.StdEncoding.DecodeString(m[bodyKey].(string))
	if err != nil {
		log.Printf("读取缓存错误: %v", err)
		return nil
	}
	return value
}

func (p *FileCache) cacheBody(m map[string]interface{}) {
	filePath := "/Users/itachi/go/src/github.com/itachizhu/api-gateway-go/tmp/"
	filePath = filepath.FromSlash(filePath)
	fileName := "attachment"
	if len(p.headers) > 0 {
		values, ok := p.headers["Content-Disposition"]
		if ok && len(values) > 0 {
			_fileName := values[0]
			index := strings.Index(_fileName, "filename=")
			if index > 0 {
				fileName = _fileName[index+len("filename=")+1: len(_fileName)-1]
			} else {
				index = strings.Index(_fileName, "filename*=")
				if index > 0 {
					fileName = _fileName[index+len("filename*="):]
				}
			}
		}
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			log.Printf("目录不存在，创建目录错误: %v", err)
			panic(err)
		}
	}
	filePath += p.service.AppName + "-" + strconv.FormatInt(time.Now().UnixNano()/1000, 10) + "-" + fileName
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		err = os.Remove(filePath)
		if err != nil {
			log.Printf("文件已存在，删除文件错误: %v", err)
			panic(err)
		}
	}
	err := ioutil.WriteFile(filePath, p.body, os.ModePerm)
	if err != nil {
		log.Printf("写入文件数据错误: %v", err)
		panic(err)
	}
	m[bodyKey] = filePath
}

func (p *FileCache) bodyFromCache(m map[string]interface{}) []byte {
	return nil
}
