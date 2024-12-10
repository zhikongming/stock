package utils

import (
	"encoding/json"
	"fmt"
)

func PrintJsonString(s interface{}) {
	data, _ := json.Marshal(s)
	fmt.Printf("%s\n", string(data))
}
