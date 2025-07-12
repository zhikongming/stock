package report

import (
	"context"
	"encoding/json"
	"fmt"

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
