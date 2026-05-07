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
	IsParsedPrice bool   `json:"is_parsed_price" gorm:"column:is_parsed_price"`
	BdCompanyCode string `json:"bd_company_code" gorm:"column:bd_company_code"`
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

func GetStockCodeByCodeList(ctx context.Context, list []string) ([]*StockCode, error) {
	db := GetDB()
	var stockCodeList []*StockCode
	err := db.WithContext(ctx).Find(&stockCodeList).Where("company_code in ?", list).Error
	if err != nil {
		return nil, err
	}
	return stockCodeList, nil
}

func GetStockCodeByIndustry(ctx context.Context, industryType int) ([]*StockCode, error) {
	db := GetDB()
	var stockCodeList []*StockCode
	err := db.WithContext(ctx).Where("business_type = ?", industryType).Find(&stockCodeList).Error
	if err != nil {
		return nil, err
	}
	return stockCodeList, nil
}

func GetStockCodeByCode(ctx context.Context, code string) (*StockCode, error) {
	db := GetDB()
	var stockCode StockCode
	err := db.WithContext(ctx).Where("company_code = ? or company_code_hk = ?", code, code).First(&stockCode).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
	return &stockCode, nil
}

func GetStockCodeByCodeOrName(ctx context.Context, codeOrName string) (*StockCode, error) {
	db := GetDB()
	var stockCode StockCode
	param := "%" + codeOrName + "%"
	err := db.WithContext(ctx).Where("company_code like ? or company_code_hk like ? or company_name like ? or company_name_hk like ?", param, param, param, param).First(&stockCode).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
	return &stockCode, nil
}

func GetStockCodeByName(ctx context.Context, name string) (*StockCode, error) {
	db := GetDB()
	var stockCode StockCode
	err := db.WithContext(ctx).Where("company_name = ?", name).First(&stockCode).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
	return &stockCode, nil
}

// GetStockCodeByParsedPrice 根据IsParsedPrice字段获取股票代码
func GetStockCodeByParsedPrice(ctx context.Context, isParsedPrice bool) ([]*StockCode, error) {
	db := GetDB()
	var stockCodeList []*StockCode
	err := db.WithContext(ctx).Where("is_parsed_price = ?", isParsedPrice).Find(&stockCodeList).Error
	if err != nil {
		return nil, err
	}
	return stockCodeList, nil
}

func AddParsedPriceCodeList(ctx context.Context, codeList []string) error {
	db := GetDB()
	return db.WithContext(ctx).Model(&StockCode{}).Where("company_code in ?", codeList).Update("is_parsed_price", true).Error
}

func DeleteParsedPriceCodeList(ctx context.Context, codeList []string) error {
	db := GetDB()
	return db.WithContext(ctx).Model(&StockCode{}).Where("company_code in ?", codeList).Update("is_parsed_price", false).Error
}

func UpdateStockCode(ctx context.Context, stockCode *StockCode) error {
	db := GetDB()
	return db.WithContext(ctx).Save(stockCode).Error
}
