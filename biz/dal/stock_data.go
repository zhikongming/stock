package dal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

func IsStockBasicExist(code string) bool {
	filePath := utils.GetLocalStoragePath(fmt.Sprintf("%s/%s", code, utils.StockBasicFile))
	return IsRegularFileExist(filePath)
}

func CreateStockBasic(code string) error {
	filePath := utils.GetLocalStoragePath(fmt.Sprintf("%s/%s", code, utils.StockBasicFile))
	return CreateRegularFile(filePath)
}

func GetStockBasic(code string) (*model.StockBasicDataCompany, error) {
	filePath := utils.GetLocalStoragePath(fmt.Sprintf("%s/%s", code, utils.StockBasicFile))
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, utils.ErrorRecordNotFound
	}
	stockBasic := &model.StockBasicDataCompany{}
	err = json.Unmarshal(data, stockBasic)
	if err != nil {
		return nil, err
	}
	return stockBasic, nil
}

func SaveStockBasic(code string, stockBasic *model.StockBasicDataCompany) error {
	filePath := utils.GetLocalStoragePath(fmt.Sprintf("%s/%s", code, utils.StockBasicFile))
	data, err := json.Marshal(stockBasic)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
