package dal

import (
	"os"

	"github.com/zhikongming/stock/utils"
)

func IsStockCodeExist(code string) bool {
	filePath := utils.GetLocalStoragePath(code)
	return IsDirExist(filePath)
}

func CreateStockCode(code string) error {
	filePath := utils.GetLocalStoragePath(code)
	return os.Mkdir(filePath, 0777)
}
