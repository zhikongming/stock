package dal

import (
	"context"

	"gorm.io/gorm"
)

type StockReport struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	CompanyCode  string `gorm:"column:company_code"`
	Year         int    `gorm:"column:year"`
	ReportType   int    `gorm:"column:report_type"`
	Report       string `gorm:"column:report"`
	Measurement  string `json:"measurement"`
	IndustryType int    `json:"industry_type"`
	Comment      string `json:"comment"`
}

type StockReportSorter []*StockReport

func (s StockReportSorter) Len() int {
	return len(s)
}

func (s StockReportSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s StockReportSorter) Less(i, j int) bool {
	if s[i].Year < s[j].Year {
		return true
	} else if s[i].Year > s[j].Year {
		return false
	} else {
		return s[i].ReportType < s[j].ReportType
	}
}

func (StockReport) TableName() string {
	return "stock_report"
}

func GetStockReport(ctx context.Context, companyCode string, year int, reportType int) (*StockReport, error) {
	var report StockReport
	err := db.WithContext(ctx).Where("company_code = ? AND year = ? AND report_type = ?", companyCode, year, reportType).First(&report).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &report, nil
}

// 获取同比财报数据
func GetStockYearOnYearReport(ctx context.Context, companyCode string, year int, reportType int) (*StockReport, error) {
	year -= 1
	return GetStockReport(ctx, companyCode, year, reportType)
}

// 获取环比财报数据
func GetStockMonthOnMonthReport(ctx context.Context, companyCode string, year int, reportType int) (*StockReport, error) {
	if reportType == 1 {
		year -= 1
		reportType = 4
	} else {
		reportType -= 1
	}
	return GetStockReport(ctx, companyCode, year, reportType)
}

func UpdateStockReport(ctx context.Context, report *StockReport) error {
	err := db.WithContext(ctx).Model(&StockReport{}).Where("company_code =? AND year =? AND report_type =?", report.CompanyCode, report.Year, report.ReportType).Updates(
		map[string]interface{}{
			"report":      report.Report,
			"comment":     report.Comment,
			"measurement": report.Measurement,
		},
	).Error
	if err != nil {
		return err
	}
	return nil
}

func CreateStockReport(ctx context.Context, report *StockReport) error {
	err := db.WithContext(ctx).Create(report).Error
	if err != nil {
		return err
	}
	return nil
}

func GetAllReports(ctx context.Context, companyCode string) ([]*StockReport, error) {
	var reports []*StockReport
	err := db.WithContext(ctx).Where("company_code =?", companyCode).Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func GetReportsByIndustry(ctx context.Context, industryType int) ([]*StockReport, error) {
	stockList, err := GetStockCodeByIndustry(ctx, industryType)
	if err != nil {
		return nil, err
	}
	stockCodeList := make([]string, 0)
	for _, stock := range stockList {
		stockCodeList = append(stockCodeList, stock.CompanyCode)
	}
	var reports []*StockReport
	err = db.WithContext(ctx).Where("company_code in ?", stockCodeList).Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}
