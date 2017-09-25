package main

import (
	"mime/multipart"
	"os"
	"io"
	"bytes"
	"net/http"
	"io/ioutil"
	"log"
)

func main() {
	buffer := new(bytes.Buffer)
	w := multipart.NewWriter(buffer)
	file, err := os.Open("/Users/itachi/Downloads/apktool.txt")
	if err != nil {
		log.Printf("1 %v", err)
		return
	}
	defer file.Close()
	part, err := w.CreateFormFile("file", "apktool.txt")
	if err != nil {
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Printf("2 %v", err)
		return
	}
	log.Printf("%v", buffer.Len())
	err = w.Close()
	if err != nil {
		log.Printf("%v", err)
		return
	}
	request, err := http.NewRequest("POST", "http://localhost:9070/upload", buffer)
	if err != nil {
		log.Printf("3 %v", err)
		return
	}
	request.Header.Set("Content-Type", w.FormDataContentType())
	var client http.Client
	response, err := client.Do(request)
	if err != nil {
		log.Printf("4 %v", err)
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	log.Printf("5 %v", string(body))
}
