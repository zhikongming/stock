package model

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/utils"
)

const (
	StableBadDebtRate   = 0.05
	StableMigrationRate = 0.1
	StableAdequacyRate  = 9.0
	MinAdequacyRate     = 7.75
)

func (s *BankReport) GetPreReport(ctx context.Context, base *StockReportBase) error {
	preRp, err := dal.GetStockMonthOnMonthReport(ctx, base.Code, base.Year, base.ReportType.ToInt())
	if err != nil {
		return err
	}
	if preRp == nil {
		return nil
	}
	preReport := &BankReport{}
	err = json.Unmarshal([]byte(preRp.Report), preReport)
	if err != nil {
		return err
	}
	s.preMOMReport = preReport

	preYOYRp, err := dal.GetStockYearOnYearReport(ctx, base.Code, base.Year, base.ReportType.ToInt())
	if err != nil {
		return err
	}
	if preYOYRp == nil {
		return nil
	}
	preYOYReport := &BankReport{}
	err = json.Unmarshal([]byte(preYOYRp.Report), preYOYReport)
	if err != nil {
		return err
	}
	s.preYOYReport = preYOYReport
	return nil
}

func (s *BankReport) GetPreYOYReport() *BankReport {
	return s.preYOYReport
}

func (s *BankReport) GetPreMOMReport() *BankReport {
	return s.preMOMReport
}

func (s *BankReport) GetMsg(ctx context.Context, base *StockReportBase) (*BankMessage, error) {
	err := s.GetPreReport(ctx, base)
	if err != nil {
		return nil, err
	}

	resp := &BankMessage{}
	// 股东分析
	msg, err := s.GetShareholderMsg(ctx, base)
	if err != nil {
		return nil, err
	} else {
		resp.ShareholderMsg = msg
	}
	// 营收分析
	msg, err = s.GetIncomeMsg(ctx, base)
	if err != nil {
		return nil, err
	} else {
		resp.IncomeMsg = msg
	}
	// 资产分析
	msg, err = s.GetAssetMsg(ctx, base)
	if err != nil {
		return nil, err
	} else {
		resp.AssetMsg = msg
	}
	// 不良分析
	msg, err = s.GetBadDebtMsg(ctx, base)
	if err != nil {
		return nil, err
	} else {
		resp.BadDebtMsg = msg
	}
	// 拨备覆盖率分析
	msg, err = s.GetCoverageRateMsg(ctx, base)
	if err != nil {
		return nil, err
	} else {
		resp.CoverageRateMsg = msg
	}
	// 资本充足率分析
	msg, err = s.GetAdequacyRateMsg(ctx, base)
	if err != nil {
		return nil, err
	} else {
		resp.AdequacyRateMsg = msg
	}
	return resp, nil
}

func (s *BankReport) GetShareholderMsg(ctx context.Context, base *StockReportBase) (string, error) {
	if s.Shareholder == nil {
		return "", nil
	}
	// 股东分析，依赖前一个季度的数据
	if s.preMOMReport == nil {
		return "", nil
	}
	// 生成股东人数解析数据
	template := fmt.Sprintf("<p>%v股东变化：</p>", base.GetYearMsg())
	index := 1
	// 分析前十大股东里面，新增和退出情况
	currentShareholderMap := make(map[string]*TopShareholder)
	for _, shareholder := range s.Shareholder.TopShareholderList {
		currentShareholderMap[shareholder.Name] = shareholder
	}
	preShareholderMap := make(map[string]*TopShareholder)
	for _, shareholder := range s.preMOMReport.Shareholder.TopShareholderList {
		preShareholderMap[shareholder.Name] = shareholder
	}
	newShareholderList := make([]*TopShareholder, 0)
	deleteShareholderList := make([]*TopShareholder, 0)
	for _, shareholder := range s.Shareholder.TopShareholderList {
		_, ok := preShareholderMap[shareholder.Name]
		if !ok {
			newShareholderList = append(newShareholderList, shareholder)
		}
	}
	for _, shareholder := range s.preMOMReport.Shareholder.TopShareholderList {
		_, ok := currentShareholderMap[shareholder.Name]
		if !ok {
			deleteShareholderList = append(deleteShareholderList, shareholder)
		}
	}
	if len(newShareholderList) > 0 || len(deleteShareholderList) > 0 {
		template += fmt.Sprintf("%d. ", index)
		template += "新增前十大股东："
		for _, shareholder := range newShareholderList {
			template += fmt.Sprintf("%v(%v);", shareholder.Name, shareholder.Share)
		}
		template += "退出前十大股东："
		for _, shareholder := range deleteShareholderList {
			template += fmt.Sprintf("%v(%v);", shareholder.Name, shareholder.Share)
		}
		index++
	}
	// 分析前十大股东里面，股东数量变化情况
	diffShareholderList := make([]*TopShareholder, 0)
	for name, currentShareholder := range currentShareholderMap {
		preShareholder, ok := preShareholderMap[name]
		if !ok {
			continue
		}
		if currentShareholder.Share != preShareholder.Share {
			diffShareholderList = append(diffShareholderList, &TopShareholder{
				Name:  name,
				Share: currentShareholder.Share - preShareholder.Share,
				Ratio: utils.Float64KeepDecimal(currentShareholder.Ratio-preShareholder.Ratio, 2),
			})
		}
	}
	if len(diffShareholderList) > 0 {
		template += fmt.Sprintf("<p>%d. ", index)
		template += "前十大股东数量变化："
		for _, shareholder := range diffShareholderList {
			if shareholder.Share > 0 {
				template += fmt.Sprintf("%v 增持股数 %d, 增持比例 %v%%; ", shareholder.Name, shareholder.Share, shareholder.Ratio)
			} else {
				template += fmt.Sprintf("%v 减持股数 %d, 减持比例 %v%%; ", shareholder.Name, -shareholder.Share, -shareholder.Ratio)
			}
		}
		index++
	}
	template += "其他大股东持股不变.</p>"

	if s.Shareholder.ShareholderNumber > s.preMOMReport.Shareholder.ShareholderNumber {
		template += fmt.Sprintf("<p>%d. ", index)
		template += fmt.Sprintf("持股股东数环比由 %v 户增加到 %v 户, 增加了 %v 户, 说明股东人数增加, 大量散户冲进来不利于后续上涨, 会有获利盘会在股价上升中卖出.</p>", s.preMOMReport.Shareholder.ShareholderNumber, s.Shareholder.ShareholderNumber, s.Shareholder.ShareholderNumber-s.preMOMReport.Shareholder.ShareholderNumber)
	} else {
		template += fmt.Sprintf("<p>%d. ", index)
		template += fmt.Sprintf("持股股东数环比由 %v 户减少到 %v 户, 减少了 %v 户, 说明股东人数减少, 筹码进一步被机构持有, 有利于后续股价上涨.</p>", s.preMOMReport.Shareholder.ShareholderNumber, s.Shareholder.ShareholderNumber, s.preMOMReport.Shareholder.ShareholderNumber-s.Shareholder.ShareholderNumber)
	}

	return template, nil
}

func GetDiffMessage(current, pre float64) string {
	if current > pre {
		return "增加"
	} else {
		return "减少"
	}
}

func GetDiffRatio(current, pre float64) float64 {
	return utils.Float64KeepDecimal(math.Abs(current-pre)/pre*100, 2)
}

func GetBpsRatio(current, pre float64) float64 {
	return utils.Float64KeepDecimal(math.Abs(current-pre)*100, 2)
}

func GetAbsValue(val float64) float64 {
	return utils.Float64KeepDecimal(math.Abs(val), 2)
}

func (s *BankReport) GetIncomeMsg(ctx context.Context, base *StockReportBase) (string, error) {
	if s.Income == nil {
		return "", nil
	}
	// 分析营业收入同比信息
	if s.preYOYReport == nil {
		return "", nil
	}

	if s.preMOMReport == nil {
		return "", nil
	}
	msg := fmt.Sprintf("<p>营业收入 %v %v同比%v %v%%. ", s.Income.TotalIncome, base.GetMeasurement(), GetDiffMessage(s.Income.TotalIncome, s.preYOYReport.Income.TotalIncome), GetDiffRatio(s.Income.TotalIncome, s.preYOYReport.Income.TotalIncome))
	msg += fmt.Sprintf("其中净利息收入 %v %v，同比%v %v%%; ", s.Income.InterestIncome.TotalIncome, base.GetMeasurement(), GetDiffMessage(s.Income.InterestIncome.TotalIncome, s.preYOYReport.Income.InterestIncome.TotalIncome), GetDiffRatio(s.Income.InterestIncome.TotalIncome, s.preYOYReport.Income.InterestIncome.TotalIncome))
	msg += fmt.Sprintf("手续费收入 %v %v，同比%v %v%%; ", s.Income.ServiceIncome, base.GetMeasurement(), GetDiffMessage(s.Income.ServiceIncome, s.preYOYReport.Income.ServiceIncome), GetDiffRatio(s.Income.ServiceIncome, s.preYOYReport.Income.ServiceIncome))
	msg += fmt.Sprintf("其他非息收入 %v %v，同比%v %v%%. </p>", s.Income.NonInterestIncome, base.GetMeasurement(), GetDiffMessage(s.Income.NonInterestIncome, s.preYOYReport.Income.NonInterestIncome), GetDiffRatio(s.Income.NonInterestIncome, s.preYOYReport.Income.NonInterestIncome))
	// 分析净息差
	msg += fmt.Sprintf("<p>第%v单季净息差 %v%%，环比%v %v bps, 同比%v %v bps。</p>",
		base.GetYearMsg(),
		s.Income.InterestIncome.InterestRatePeriod,
		GetDiffMessage(s.Income.InterestIncome.InterestRatePeriod, s.preMOMReport.Income.InterestIncome.InterestRatePeriod),
		GetBpsRatio(s.Income.InterestIncome.InterestRatePeriod, s.preMOMReport.Income.InterestIncome.InterestRatePeriod),
		GetDiffMessage(s.Income.InterestIncome.InterestRatePeriod, s.preYOYReport.Income.InterestIncome.InterestRatePeriod),
		GetBpsRatio(s.Income.InterestIncome.InterestRatePeriod, s.preYOYReport.Income.InterestIncome.InterestRatePeriod))
	loanRateDiff := math.Abs(s.Income.InterestIncome.LoanRatePeriod - s.preMOMReport.Income.InterestIncome.LoanRatePeriod)
	depositRateDiff := math.Abs(s.Income.InterestIncome.DepositRatePeriod - s.preMOMReport.Income.InterestIncome.DepositRatePeriod)
	reason := "存款成本率"
	if loanRateDiff > depositRateDiff {
		reason = "贷款收益率"
	}
	diffMsg := GetDiffMessage(s.Income.InterestIncome.InterestRate, s.preMOMReport.Income.InterestIncome.InterestRate)
	msg += fmt.Sprintf("<p>通过分析单季资产收益率和负债成本, 净息差%v主要是由%v%v造成的. ",
		diffMsg, reason, diffMsg)
	loanDiffMsg := GetDiffMessage(s.Income.InterestIncome.LoanRatePeriod, s.preMOMReport.Income.InterestIncome.LoanRatePeriod)
	msg += fmt.Sprintf("首先看资产端，贷款收益率环比%v %v bps, 主要是受到存量按揭全部重定价和LPR影响.", loanDiffMsg, GetBpsRatio(s.Income.InterestIncome.LoanRatePeriod, s.preMOMReport.Income.InterestIncome.LoanRatePeriod))
	depositDiffMsg := GetDiffMessage(s.Income.InterestIncome.DepositRatePeriod, s.preMOMReport.Income.InterestIncome.DepositRatePeriod)
	msg += fmt.Sprintf("负债端，存款成本率环比%v %v bps, 和银行自行调整存款收益率有关.</p>", depositDiffMsg, GetBpsRatio(s.Income.InterestIncome.DepositRatePeriod, s.preMOMReport.Income.InterestIncome.DepositRatePeriod))
	// 分析
	return msg, nil
}

func (s *BankReport) GetAssetMsg(ctx context.Context, base *StockReportBase) (string, error) {
	if s.Asset == nil {
		return "", nil
	}
	if s.preYOYReport == nil {
		return "", nil
	}
	totalRate := GetDiffRatio(s.Asset.TotalAsset, s.preYOYReport.Asset.TotalAsset)
	loanRate := GetDiffRatio(s.Asset.BankLoan.TotalLoan, s.preYOYReport.Asset.BankLoan.TotalLoan)
	msg := fmt.Sprintf("<p>总资产 %v %v同比增长 %v%%，其中贷款总额%v %v，同比增速 %v%%。</p><p>总负债 %v %v同比增长 %v%%，其中存款 %v%v，同比增长 %v%%。</p>",
		s.Asset.TotalAsset,
		base.GetMeasurement(),
		totalRate,
		s.Asset.BankLoan.TotalLoan,
		base.GetMeasurement(),
		loanRate,
		s.Asset.TotalDebt,
		base.GetMeasurement(),
		GetDiffRatio(s.Asset.TotalDebt, s.preYOYReport.Asset.TotalDebt),
		s.Asset.BankDeposit.TotalDeposit,
		base.GetMeasurement(),
		GetDiffRatio(s.Asset.BankDeposit.TotalDeposit, s.preYOYReport.Asset.BankDeposit.TotalDeposit),
	)
	// 分析贷款
	rateDiffMsg := "低于"
	reasonDiffMsg := "零售信贷需求不足和政府发债增加"
	if loanRate > totalRate {
		rateDiffMsg = "高于"
		reasonDiffMsg = "零售信贷需求旺盛和政府发债减少"
	}
	msg += fmt.Sprintf("<p>财报数据显示，贷款增速%v总资产增速，说明%v。", rateDiffMsg, reasonDiffMsg)
	// 分析存款
	if s.Asset.BankDeposit.CorporateDeposit > 0 {
		corporateDepositRate := GetDiffRatio(s.Asset.BankDeposit.CorporateDeposit, s.preYOYReport.Asset.BankDeposit.CorporateDeposit)
		personDeposit := GetDiffRatio(s.Asset.BankDeposit.PersonDeposit, s.preYOYReport.Asset.BankDeposit.PersonDeposit)
		if corporateDepositRate > 0 {
			msg += fmt.Sprintf("对公存款同比增加 %v%%, 说明对公客户非常牢固。", math.Abs(corporateDepositRate))
		} else {
			msg += fmt.Sprintf("对公存款同比减少 %v%%, 说明对公客户不够牢固，另外存款增速偏低未来会影响银行的扩张速度和负债成本。", math.Abs(corporateDepositRate))
		}
		if personDeposit > 0 {
			msg += fmt.Sprintf("零售存款同比增加 %v%%, 说明负债端优势在强化。", math.Abs(personDeposit))
		} else {
			msg += fmt.Sprintf("零售存款同比减少 %v%%, 说明负债端优势在减弱。", math.Abs(personDeposit))
		}
	}
	msg += "</p>"

	return msg, nil
}

func (s *BankReport) GetBadDebtMsg(ctx context.Context, base *StockReportBase) (string, error) {
	if s.BadDebtAsset == nil {
		return "", nil
	}
	if s.preMOMReport == nil {
		return "", nil
	}
	// 分析不良基础数据
	msg := fmt.Sprintf("<p>根据财报，不良余额 %v %v环比 %v %v%v了 %v %v，",
		s.BadDebtAsset.TotalBalance,
		base.GetMeasurement(),
		s.preMOMReport.BadDebtAsset.TotalBalance,
		base.GetMeasurement(),
		GetDiffMessage(s.BadDebtAsset.TotalBalance, s.preMOMReport.BadDebtAsset.TotalBalance),
		GetAbsValue(s.BadDebtAsset.TotalBalance-s.preMOMReport.BadDebtAsset.TotalBalance),
		base.GetMeasurement(),
	)
	msg += fmt.Sprintf("不良率 %v%%环比 %v%%,%v了 %v 个百分点。</p>",
		s.BadDebtAsset.TotalRate,
		s.preMOMReport.BadDebtAsset.TotalRate,
		GetDiffMessage(s.BadDebtAsset.TotalRate, s.preMOMReport.BadDebtAsset.TotalRate),
		GetAbsValue(s.BadDebtAsset.TotalRate-s.preMOMReport.BadDebtAsset.TotalRate),
	)
	msg += fmt.Sprintf("<p>关注贷款余额 %v %v环比 %v %v%v了 %v %v，",
		s.BadDebtAsset.BadDebt.NoticeBalance,
		base.GetMeasurement(),
		s.preMOMReport.BadDebtAsset.BadDebt.NoticeBalance,
		base.GetMeasurement(),
		GetDiffMessage(s.BadDebtAsset.BadDebt.NoticeBalance, s.preMOMReport.BadDebtAsset.BadDebt.NoticeBalance),
		GetAbsValue(s.BadDebtAsset.BadDebt.NoticeBalance-s.preMOMReport.BadDebtAsset.BadDebt.NoticeBalance),
		base.GetMeasurement(),
	)
	msg += fmt.Sprintf("关注率 %v%% 环比 %v%% %v了 %v 个百分点.</p>",
		s.BadDebtAsset.BadDebt.NoticeRate,
		s.preMOMReport.BadDebtAsset.BadDebt.NoticeRate,
		GetDiffMessage(s.BadDebtAsset.BadDebt.NoticeRate, s.preMOMReport.BadDebtAsset.BadDebt.NoticeRate),
		GetAbsValue(s.BadDebtAsset.BadDebt.NoticeRate-s.preMOMReport.BadDebtAsset.BadDebt.NoticeRate),
	)
	// 分析新生成不良率
	newYOYRateDiff := s.BadDebtAsset.NewRate - s.preYOYReport.BadDebtAsset.NewRate
	newMOMRateDiff := s.BadDebtAsset.NewRate - s.preMOMReport.BadDebtAsset.NewRate
	msg += fmt.Sprintf("<p>根据财报披露的口径，新生成不良率 %v%% 环比%v了 %v 个百分点, 同比%v了 %v 个百分点, ",
		s.BadDebtAsset.NewRate,
		GetDiffMessage(s.BadDebtAsset.NewRate, s.preMOMReport.BadDebtAsset.NewRate),
		GetAbsValue(newMOMRateDiff),
		GetDiffMessage(s.BadDebtAsset.NewRate, s.preYOYReport.BadDebtAsset.NewRate),
		GetAbsValue(newYOYRateDiff),
	)
	if GetAbsValue(newYOYRateDiff-newMOMRateDiff) < StableBadDebtRate {
		msg += "从环比看还是同比看，目前该行的资产质量比较稳定。"
	} else {
		if newMOMRateDiff > 0 {
			msg += "从环比看，新生不良率快速增长，说明银行资产质量下降，需要重点关注。"
		} else {
			msg += "从环比看，新生不良率快速下降，说明银行资产质量上升，需要关注。"
		}
	}
	msg += "</p>"
	// 分析正常迁徙率
	migrationMsg := "较小"
	migrationRateDiff := s.BadDebtAsset.MigrationRate.NormalRate - s.preYOYReport.BadDebtAsset.MigrationRate.NormalRate
	if migrationRateDiff > StableMigrationRate {
		migrationMsg = "较大"
	}
	msg += fmt.Sprintf("<p>从正常贷款迁徙率看，该行的资产质量压力%v。正常贷款的迁徙率 %v%% 同比 %v%% %v %v 个百分点。</p>",
		migrationMsg,
		s.BadDebtAsset.MigrationRate.NormalRate,
		s.preYOYReport.BadDebtAsset.MigrationRate.NormalRate,
		GetDiffMessage(s.BadDebtAsset.MigrationRate.NormalRate, s.preYOYReport.BadDebtAsset.MigrationRate.NormalRate),
		GetAbsValue(migrationRateDiff),
	)
	return msg, nil
}

func (s *BankReport) GetAdequacyRateMsg(ctx context.Context, base *StockReportBase) (string, error) {
	if s.AdequacyRate == 0.0 {
		return "", nil
	}
	if s.preMOMReport == nil {
		return "", nil
	}
	if s.preYOYReport == nil {
		return "", nil
	}
	// 分析资本充足率
	msg := fmt.Sprintf("核心一级资本充足率 %v%% 环比 %v%% %v了 %v 个百分点，同比 %v%% %v了 %v 个百分点。",
		s.AdequacyRate,
		s.preMOMReport.AdequacyRate,
		GetDiffMessage(s.AdequacyRate, s.preMOMReport.AdequacyRate),
		GetAbsValue(s.AdequacyRate-s.preMOMReport.AdequacyRate),
		s.preYOYReport.AdequacyRate,
		GetDiffMessage(s.AdequacyRate, s.preYOYReport.AdequacyRate),
		GetAbsValue(s.AdequacyRate-s.preYOYReport.AdequacyRate),
	)
	adequacyMsg := ""
	if s.AdequacyRate > StableAdequacyRate {
		adequacyMsg = "银行的资本充足率较高，说明银行短期没有配股或者发行可转债的需求。"
	} else {
		adequacyMsg = "银行的资本充足率较低，说明银行资产不足以支撑业务的扩张，因此要么降低营收增速，要么短期有配股或者发行可转债的需求。"
	}
	msg += fmt.Sprintf("商业银行的核心一级资本充足率的最低要求为%v%%, 该行目前%v", MinAdequacyRate, adequacyMsg)

	return msg, nil
}

func (s *BankReport) GetCoverageRateMsg(ctx context.Context, base *StockReportBase) (string, error) {
	if s.BadDebtAsset.CoverageRate == 0.0 {
		return "", nil
	}
	if s.preMOMReport == nil {
		return "", nil
	}
	if s.preYOYReport == nil {
		return "", nil
	}
	msg := fmt.Sprintf("根据财报，拨备覆盖率 %v%% 环比 %v%% %v了 %v 个百分点。",
		s.BadDebtAsset.CoverageRate,
		s.preMOMReport.BadDebtAsset.CoverageRate,
		GetDiffMessage(s.BadDebtAsset.CoverageRate, s.preMOMReport.BadDebtAsset.CoverageRate),
		GetAbsValue(s.BadDebtAsset.CoverageRate-s.preMOMReport.BadDebtAsset.CoverageRate),
	)
	return msg, nil
}
