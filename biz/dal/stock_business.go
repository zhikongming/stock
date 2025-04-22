package dal

import "gorm.io/gorm"

type StockBusiness struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	BusinessName string `json:"business_name" gorm:"column:business_name"`
}

func (StockBusiness) TableName() string {
	return "stock_business"
}

func GetStockBusiness(id uint) (*StockBusiness, error) {
	db := GetDB()
	var stockBusiness StockBusiness
	err := db.Where("id =?", id).First(&stockBusiness).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
	return &stockBusiness, nil
}
