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
	if disableMsg {
		return resp, nil
	}
	msg, err := report.GetMsg(ctx, base)
	if err != nil {
		return nil, err
	}
	resp.Message = msg
	resp.PreMOMReport = report.GetPreMOMReport()
	resp.PreYOYReport = report.GetPreYOYReport()
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
