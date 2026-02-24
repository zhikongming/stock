package dal

import (
	"context"

	"gorm.io/gorm"
)

type StockIndustry struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Code string `json:"code" gorm:"column:code"`
	Name string `json:"name" gorm:"column:name"`
}

func (StockIndustry) TableName() string {
	return "stock_industry"
}

func GetAllStockIndustry(ctx context.Context) ([]*StockIndustry, error) {
	var data []*StockIndustry
	if err := db.WithContext(ctx).Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func GetStockIndustry(ctx context.Context, code string) (*StockIndustry, error) {
	var data *StockIndustry
	if err := db.WithContext(ctx).Where("code = ?", code).First(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func GetStockIndustryByName(ctx context.Context, name string) (*StockIndustry, error) {
	var data *StockIndustry
	if err := db.WithContext(ctx).Where("name = ?", name).First(&data).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
	return data, nil
}

func AddStockIndustry(ctx context.Context, industry *StockIndustry) error {
	if err := db.WithContext(ctx).Create(industry).Error; err != nil {
		return err
	}
	return nil
}

func DeleteStockIndustry(ctx context.Context, industry *StockIndustry) error {
	return db.WithContext(ctx).Delete(industry).Error
}

type StockIndustryRelation struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	CompanyCode  string `json:"company_code" gorm:"column:company_code"`
	IndustryCode string `json:"industry_code" gorm:"column:industry_code"`
}

func (StockIndustryRelation) TableName() string {
	return "stock_industry_relation"
}

func AddStockIndustryRelation(ctx context.Context, relation *StockIndustryRelation) error {
	return db.WithContext(ctx).Create(relation).Error
}

func GetStockIndustryRelation(ctx context.Context, industryCode string) ([]*StockIndustryRelation, error) {
	var data []*StockIndustryRelation
	if err := db.WithContext(ctx).Where("industry_code = ?", industryCode).Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func GetStockIndustryRelationByCompanyCode(ctx context.Context, companyCode string) (*StockIndustryRelation, error) {
	var data *StockIndustryRelation
	if err := db.WithContext(ctx).Where("company_code = ?", companyCode).First(&data).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
	return data, nil
}

func GetAllStockIndustryRelation(ctx context.Context) ([]*StockIndustryRelation, error) {
	var data []*StockIndustryRelation
	if err := db.WithContext(ctx).Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func DeleteStockIndustryRelation(ctx context.Context, relation *StockIndustryRelation) error {
	return db.WithContext(ctx).Delete(relation).Error
}
