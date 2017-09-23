package util

import "encoding/json"

func Panic(code int32, message string) {
	m := map[string]interface{}{
		"errorCode": code,
		"errorMessage": message,
	}
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	panic(data)
}