package dal

import (
	"context"

	"gorm.io/gorm"
)

type StockCode struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	CompanyCode   string `json:"company_code" gorm:"column:company_code"`
	CompanyCodeHK string `json:"company_code_hk" gorm:"column:company_code_hk"`
	CompanyName   string `json:"company_name" gorm:"column:company_name"`
	CompanyNameHK string `json:"company_name_hk" gorm:"column:company_name_hk"`
	ClassiName    string `json:"classi_name" gorm:"column:classi_name"`
	BusinessType  int    `json:"business_type" gorm:"column:business_type"`
	ListedDate    string `json:"listed_date" gorm:"column:listed_date"`
}

func (StockCode) TableName() string {
	return "stock_code"
}

func IsStockCodeExist(ctx context.Context, code string) (bool, error) {
	db := GetDB()
	var stockCode StockCode
	err := db.WithContext(ctx).Where("company_code = ? or company_code_hk = ?", code, code).First(&stockCode).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func CreateStockCode(ctx context.Context, stockCode *StockCode) error {
	db := GetDB()
	return db.WithContext(ctx).Create(stockCode).Error
}

func GetAllStockCode(ctx context.Context) ([]*StockCode, error) {
	db := GetDB()
	var stockCodeList []*StockCode
	err := db.WithContext(ctx).Find(&stockCodeList).Error
	if err != nil {
		return nil, err
	}
	return stockCodeList, nil
}
