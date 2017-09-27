package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"
)

func main() {
	path := "/hello"
	pathHash := sha256.Sum256([]byte(path))
	log.Printf("====path hash: %v", hex.EncodeToString([]byte(pathHash[:])))

	fileName := ""
	_fileName := "xxxxxx; filename=itachi.jpg"
	index := strings.Index(_fileName, "filename=")
	if index > 0 {
		fileName = _fileName[index + len("filename="):]
	} else {
		index = strings.Index(_fileName, "filename*=")
		if index > 0 {
			fileName = _fileName[index + len("filename*="):]
		}
	}
	log.Printf("===file name: %v", fileName)
}
