package utils

import (
	"encoding/json"
	"fmt"
)

func PrintJsonString(s interface{}) {
	data, _ := json.Marshal(s)
	fmt.Printf("%s\n", string(data))
}

func ToJsonString(s interface{}) string {
	data, _ := json.Marshal(s)
	return string(data)
}
