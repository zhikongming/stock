package dal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

func IsStockDailyExist(code string) bool {
	filePath := utils.GetLocalStoragePath(fmt.Sprintf("%s/%s", code, utils.StockDailyFile))
	return IsRegularFileExist(filePath)
}

func CreateStockDaily(code string) error {
	filePath := utils.GetLocalStoragePath(fmt.Sprintf("%s/%s", code, utils.StockDailyFile))
	return CreateRegularFile(filePath)
}

func GetStockDaily(code string) ([]*model.LocalStockDailyData, error) {
	filePath := utils.GetLocalStoragePath(fmt.Sprintf("%s/%s", code, utils.StockDailyFile))
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, utils.ErrorRecordNotFound
	}
	var stockDaily []*model.LocalStockDailyData
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	err = decoder.Decode(&stockDaily)
	if err != nil {
		return nil, err
	}
	return stockDaily, nil
}

func SaveStockDaily(code string, stockDaily []*model.LocalStockDailyData) error {
	filePath := utils.GetLocalStoragePath(fmt.Sprintf("%s/%s", code, utils.StockDailyFile))
	data, err := json.Marshal(stockDaily)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
