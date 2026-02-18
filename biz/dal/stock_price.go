package dal

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type StockPrice struct {
	ID                       uint      `json:"id" gorm:"primaryKey"`
	CompanyCode              string    `json:"company_code" gorm:"column:company_code"`
	Date                     time.Time `json:"date" gorm:"column:date"`
	PriceHigh                float64   `json:"price_high" gorm:"column:price_high"`
	PriceLow                 float64   `json:"price_low" gorm:"column:price_low"`
	PriceOpen                float64   `json:"price_open" gorm:"column:price_open"`
	PriceClose               float64   `json:"price_close" gorm:"column:price_close"`
	Amount                   int64     `json:"amount" gorm:"column:amount"`
	BollingUp                float64   `json:"bolling_up" gorm:"column:bolling_up"`
	BollingDown              float64   `json:"bolling_down" gorm:"column:bolling_down"`
	BollingMid               float64   `json:"bolling_mid" gorm:"column:bolling_mid"`
	Ma5                      float64   `json:"ma5" gorm:"column:ma5"`
	Ma10                     float64   `json:"ma10" gorm:"column:ma10"`
	Ma20                     float64   `json:"ma20" gorm:"column:ma20"`
	Ma30                     float64   `json:"ma30" gorm:"column:ma30"`
	Ma60                     float64   `json:"ma60" gorm:"column:ma60"`
	MacdDif                  float64   `json:"macd_dif" gorm:"column:macd_dif"`
	MacdDea                  float64   `json:"macd_dea" gorm:"column:macd_dea"`
	KdjK                     float64   `json:"kdj_k" gorm:"column:kdj_k"`
	KdjD                     float64   `json:"kdj_d" gorm:"column:kdj_d"`
	KdjJ                     float64   `json:"kdj_j" gorm:"column:kdj_j"`
	UpdateTime               time.Time `json:"update_time" gorm:"column:update_time"`
	MainInflowAmount         int64     `json:"main_inflow_amount" gorm:"column:main_inflow_amount"`
	ExtremeLargeInflowAmount int64     `json:"extreme_large_inflow_amount" gorm:"column:extreme_large_inflow_amount"`
	LargeInflowAmount        int64     `json:"large_inflow_amount" gorm:"column:large_inflow_amount"`
	MediumInflowAmount       int64     `json:"medium_inflow_amount" gorm:"column:medium_inflow_amount"`
	SmallInflowAmount        int64     `json:"small_inflow_amount" gorm:"column:small_inflow_amount"`
}

func (StockPrice) TableName() string {
	return "stock_price"
}

func (s *StockPrice) GetMacdValue() float64 {
	return s.MacdDif - s.MacdDea
}

func (s *StockPrice) IsFundInflowUpdated() bool {
	// 检查是否有资金流入数据更新
	return s.MainInflowAmount != 0 || s.ExtremeLargeInflowAmount != 0 || s.LargeInflowAmount != 0 || s.MediumInflowAmount != 0 || s.SmallInflowAmount != 0
}

func GetStockPriceByDate(ctx context.Context, code string, dateStart string, dateEnd string, limit int) ([]*StockPrice, error) {
	var stockPrice []*StockPrice
	db := GetDB()
	db = db.WithContext(ctx).Where("company_code = ?", code)
	if dateStart != "" {
		db = db.Where("date >= ?", dateStart)
	}
	if dateEnd != "" {
		db = db.Where("date <= ?", dateEnd)
	}
	if limit > 0 {
		db = db.Limit(limit)
	}
	db = db.Order("id desc")
	err := db.Find(&stockPrice).Error
	if err != nil {
		return nil, err
	}
	return stockPrice, nil
}

func GetStockPriceByCodeAndDate(ctx context.Context, code string, date string) (*StockPrice, error) {
	var stockPrice StockPrice
	db := GetDB()
	err := db.WithContext(ctx).Where("company_code =?", code).Where("date =?", date).First(&stockPrice).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
	return &stockPrice, nil
}

func GetLastStockPrice(ctx context.Context, code string) (*StockPrice, error) {
	var stockPrice StockPrice
	db := GetDB()
	err := db.WithContext(ctx).Where("company_code =?", code).Order("id desc").Limit(1).First(&stockPrice).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
	return &stockPrice, nil
}

func GetLastNStockPrice(ctx context.Context, code string, date string, limit int) ([]*StockPrice, error) {
	var stockPriceList []*StockPrice
	db := GetDB()
	db = db.WithContext(ctx).Where("company_code =?", code)
	if date != "" {
		db = db.Where("date <= ?", date)
	}
	err := db.Order("id desc").Limit(limit).Find(&stockPriceList).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
	return stockPriceList, nil
}

func CreateStockPrice(ctx context.Context, stockPrice *StockPrice) error {
	db := GetDB()
	err := db.WithContext(ctx).Create(stockPrice).Error
	if err != nil {
		return err
	}
	return nil
}

func UpdateStockPrice(ctx context.Context, stockPrice *StockPrice) error {
	db := GetDB()
	return db.WithContext(ctx).Save(stockPrice).Error
}
