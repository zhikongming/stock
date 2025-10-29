package report

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
)

func AddStockReport(ctx context.Context, req *model.AddStockReportReq) error {
	switch req.IndustryType {
	case model.IndustryTypeBank:
		return AddBankReport(ctx, req)
	default:
		return fmt.Errorf("industry type not supported")
	}
}

func AddBankReport(ctx context.Context, req *model.AddStockReportReq) error {
	// 反序列化
	report := &model.BankReport{}
	data, err := json.Marshal(req.Report)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, report)
	if err != nil {
		return err
	}
	base := &model.StockReportBase{
		Code:        req.Code,
		Year:        req.Year,
		ReportType:  req.ReportType,
		Measurement: req.Measurement,
	}
	err = report.Parse(base)
	if err != nil {
		return err
	}

	// 保存到数据库
	rp, err := dal.GetStockReport(ctx, req.Code, req.Year, int(req.ReportType))
	if err != nil {
		return err
	}

	if rp != nil {
		rp.Report = report.ToString()
		rp.Comment = req.Comment
		err = dal.UpdateStockReport(ctx, rp)
		if err != nil {
			return err
		}
		return nil
	} else {
		err = dal.CreateStockReport(ctx, &dal.StockReport{
			CompanyCode:  req.Code,
			Year:         req.Year,
			ReportType:   int(req.ReportType),
			Report:       report.ToString(),
			Measurement:  string(req.Measurement),
			IndustryType: int(req.IndustryType),
			Comment:      req.Comment,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func GetStockReport(ctx context.Context, req *model.GetStockReportReq) (*model.GetStockReportResp, error) {
	rp, err := dal.GetStockReport(ctx, req.Code, req.Year, int(req.ReportType))
	if err != nil {
		return nil, err
	}
	if rp == nil {
		return nil, fmt.Errorf("report not found")
	}

	switch rp.IndustryType {
	case int(model.IndustryTypeBank):
		return GetBankReport(ctx, rp, req.DisableMsg)
	default:
		return nil, fmt.Errorf("industry type not supported")
	}
}

func GetBankReport(ctx context.Context, rp *dal.StockReport, disableMsg bool) (*model.GetStockReportResp, error) {
	var resp *model.GetStockReportResp
	report := &model.BankReport{}
	err := json.Unmarshal([]byte(rp.Report), report)
	if err != nil {
		return nil, err
	}
	resp = &model.GetStockReportResp{
		Report: report,
	}

	base := &model.StockReportBase{
		Code:        rp.CompanyCode,
		Year:        rp.Year,
		ReportType:  model.ReportType(rp.ReportType),
		Measurement: model.MeasurementType(rp.Measurement),
	}
	resp.Measurement = base.GetMeasurement()
	resp.Comment = rp.Comment
	resp.PreMOMReport = report.GetPreMOMReport()
	resp.PreYOYReport = report.GetPreYOYReport()
	if disableMsg {
		return resp, nil
	}
	msg, err := report.GetMsg(ctx, base)
	if err != nil {
		return nil, err
	}
	resp.Message = msg
	return resp, nil
}

func GetBankTrackData(ctx context.Context, req *model.GetBankTrackDataReq) (*model.GetBankTrackDataResp, error) {
	resp := &model.GetBankTrackDataResp{
		DateList:   make([]*model.ReportTime, 0),
		ReportList: make([]*model.BankTrackData, 0),
	}
	// 从数据库获取数据
	reports, err := dal.GetAllReports(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	sort.Sort(dal.StockReportSorter(reports))
	// 解析数据
	reportList := make([]*model.BankReport, 0)
	for _, rp := range reports {
		report := &model.BankReport{}
		err = json.Unmarshal([]byte(rp.Report), report)
		if err != nil {
			return nil, err
		}
		reportList = append(reportList, report)
	}
	// 计算数据
	for idx := 0; idx < len(reportList); idx++ {
		resp.DateList = append(resp.DateList, &model.ReportTime{
			Year:       reports[idx].Year,
			ReportType: reports[idx].ReportType,
		})
		resp.ReportList = append(resp.ReportList, &model.BankTrackData{
			ShareholderNumber:  reportList[idx].Shareholder.ShareholderNumber,
			InterestRate:       reportList[idx].Income.InterestIncome.InterestRate,
			InterestRatePeriod: reportList[idx].Income.InterestIncome.InterestRatePeriod,
			ImpairmentLoss:     reportList[idx].ImpairmentLoss,
			TotalBalance:       reportList[idx].BadDebtAsset.TotalBalance,
			TotalRate:          reportList[idx].BadDebtAsset.TotalRate,
			NewBalance:         reportList[idx].BadDebtAsset.NewBalance,
			NewRate:            reportList[idx].BadDebtAsset.NewRate,
			CoverageRate:       reportList[idx].BadDebtAsset.CoverageRate,
			AdequacyRate:       reportList[idx].AdequacyRate,
			ROE:                reportList[idx].ROE,
			ROA:                reportList[idx].ROA,
			RORWA:              reportList[idx].RORWA,
		})
		if resp.Measurement == "" {
			base := &model.StockReportBase{
				Measurement: model.MeasurementType(reports[idx].Measurement),
			}
			resp.Measurement = base.GetMeasurement()
		}
	}
	return resp, nil
}

func GetIndustryTrackData(ctx context.Context, req *model.GetIndustryTrackDataReq) (*model.GetIndustryTrackDataResp, error) {
	resp := &model.GetIndustryTrackDataResp{
		DateList:   make([]*model.ReportTime, 0),
		ReportList: make(map[string][]*model.BankTrackData, 0),
	}
	// 从数据库获取数据
	allReports, err := dal.GetReportsByIndustry(ctx, int(req.IndustryType))
	if err != nil {
		return nil, err
	}
	reportMap := make(map[string][]*dal.StockReport)
	for _, rp := range allReports {
		if _, ok := reportMap[rp.CompanyCode]; !ok {
			reportMap[rp.CompanyCode] = make([]*dal.StockReport, 0)
		}
		reportMap[rp.CompanyCode] = append(reportMap[rp.CompanyCode], rp)
	}
	// 获取所有的代码信息
	companyMap := make(map[string]*dal.StockCode)
	stockCodeList, err := dal.GetAllStockCode(ctx)
	if err != nil {
		return nil, err
	}
	for _, code := range stockCodeList {
		companyMap[code.CompanyCode] = code
	}

	dateList := make([]*model.ReportTime, 0)
	dateMap := make(map[string]struct{})
	for _, reports := range reportMap {
		for idx := 0; idx < len(reports); idx++ {
			dateKey := fmt.Sprintf("%d_%d", reports[idx].Year, reports[idx].ReportType)
			if _, ok := dateMap[dateKey]; !ok {
				dateMap[dateKey] = struct{}{}
				dateList = append(dateList, &model.ReportTime{
					Year:       reports[idx].Year,
					ReportType: reports[idx].ReportType,
				})
			}
		}
	}
	dateSorter := model.ReportTimeSorter(dateList)
	sort.Sort(dateSorter)
	resp.DateList = dateList

	for code, reports := range reportMap {
		stockName := companyMap[code].CompanyName
		sort.Sort(dal.StockReportSorter(reports))
		// 解析数据
		reportList := make([]*model.BankReport, 0)
		for _, rp := range reports {
			report := &model.BankReport{}
			err = json.Unmarshal([]byte(rp.Report), report)
			if err != nil {
				return nil, err
			}
			reportList = append(reportList, report)
		}
		// 计算数据

		for idx := 0; idx < len(reportList); idx++ {
			if _, ok := resp.ReportList[stockName]; !ok {
				resp.ReportList[stockName] = make([]*model.BankTrackData, len(dateList))
			}
			targetIdx := dateSorter.GetIndex(reports[idx].Year, reports[idx].ReportType)
			resp.ReportList[stockName][targetIdx] = &model.BankTrackData{
				ShareholderNumber:  reportList[idx].Shareholder.ShareholderNumber,
				InterestRate:       reportList[idx].Income.InterestIncome.InterestRate,
				InterestRatePeriod: reportList[idx].Income.InterestIncome.InterestRatePeriod,
				ImpairmentLoss:     reportList[idx].ImpairmentLoss,
				TotalBalance:       reportList[idx].BadDebtAsset.TotalBalance,
				TotalRate:          reportList[idx].BadDebtAsset.TotalRate,
				NewBalance:         reportList[idx].BadDebtAsset.NewBalance,
				NewRate:            reportList[idx].BadDebtAsset.NewRate,
				CoverageRate:       reportList[idx].BadDebtAsset.CoverageRate,
				AdequacyRate:       reportList[idx].AdequacyRate,
				LoanRate:           reportList[idx].Income.InterestIncome.LoanRate,
				LoanRatePeriod:     reportList[idx].Income.InterestIncome.LoanRatePeriod,
				DepositRate:        reportList[idx].Income.InterestIncome.DepositRate,
				DepositRatePeriod:  reportList[idx].Income.InterestIncome.DepositRatePeriod,
				ROE:                reportList[idx].ROE,
				ROA:                reportList[idx].ROA,
				RORWA:              reportList[idx].RORWA,
			}
			if resp.Measurement == "" {
				base := &model.StockReportBase{
					Measurement: model.MeasurementType(reports[idx].Measurement),
				}
				resp.Measurement = base.GetMeasurement()
			}
		}
		for idx, _ := range resp.ReportList[stockName] {
			if resp.ReportList[stockName][idx] == nil {
				resp.ReportList[stockName][idx] = &model.BankTrackData{}
			}
		}
	}
	return resp, nil
}
