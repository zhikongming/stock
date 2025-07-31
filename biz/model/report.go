package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/zhikongming/stock/utils"
)

type ReportType int
type IndustryType int
type MeasurementType string

const (
	ReportType1 ReportType = 1
	ReportType2 ReportType = 2
	ReportType3 ReportType = 3
	ReportType4 ReportType = 4

	IndustryTypeBank   IndustryType = 1
	IndustryTypeCommon IndustryType = 2

	MeasurementTypeMillion        MeasurementType = "million"
	MeasurementTypeHundredMillion MeasurementType = "hundred_million"
)

func (r ReportType) ToInt() int {
	return int(r)
}

type StockReportBase struct {
	Code        string          `json:"code"`
	Year        int             `json:"year"`
	ReportType  ReportType      `json:"report_type"`
	Measurement MeasurementType `json:"measurement"`
	Comment     string          `json:"comment"`
}

type GetStockReportReq struct {
	Code       string     `json:"code" query:"code"`
	Year       int        `json:"year" query:"year"`
	ReportType ReportType `json:"report_type" query:"report_type"`
	DisableMsg bool       `json:"disable_msg" query:"disable_msg"`
}

type BankInterestIncome struct {
	TotalIncome        float64 `json:"total_income"`         // 净利息总收入
	InterestRate       float64 `json:"interest_rate"`        // 合并净息差
	InterestRatePeriod float64 `json:"interest_rate_period"` // 单季度净息差
	LoanRate           float64 `json:"loan_rate"`            // 贷款收益率
	LoanRatePeriod     float64 `json:"loan_rate_period"`     // 单季度贷款收益率
	DepositRate        float64 `json:"deposit_rate"`         // 存款成本率
	DepositRatePeriod  float64 `json:"deposit_rate_period"`  // 单季度存款成本率
}

type BankIncome struct {
	TotalIncome       float64             `json:"total_income"`        // 总营业收入
	InterestIncome    *BankInterestIncome `json:"interest_income"`     // 净利息收入
	ServiceIncome     float64             `json:"service_income"`      // 手续费收入
	NonInterestIncome float64             `json:"non_interest_income"` // 其他非息收入
}

type BankLoan struct {
	TotalLoan     float64 `json:"total_loan"`     // 贷款总额
	CorporateLoan float64 `json:"corporate_loan"` // 对公贷款
	PersonLoan    float64 `json:"person_loan"`    // 零售贷款
}

type BankDeposit struct {
	TotalDeposit     float64 `json:"total_deposit"`     // 存款总额：(合并资产负债表)吸收存款科目，而不是客户存款本金科目，“吸收存款”这个总负债科目，其金额就是“客户存款本金”科目余额与“应付利息”科目余额之和。
	CorporateDeposit float64 `json:"corporate_deposit"` // 对公存款：存款本金科目
	PersonDeposit    float64 `json:"person_deposit"`    // 零售存款：存款本金科目
}

type BankAsset struct {
	TotalAsset  float64      `json:"total_asset"`  // 总资产
	BankLoan    *BankLoan    `json:"bank_loan"`    // 贷款详情, 合并资产负债表-贷款和垫款
	TotalDebt   float64      `json:"total_debt"`   // 总负债
	BankDeposit *BankDeposit `json:"bank_deposit"` // 存款详情
}

type BadDebt struct {
	NormalBalance      float64 `json:"normal_balance"` // 正常类
	NormalRate         float64 `json:"normal_rate"`
	NoticeBalance      float64 `json:"notice_balance"` // 关注类
	NoticeRate         float64 `json:"notice_rate"`
	SubordinateBalance float64 `json:"subordinate_balance"` // 次级
	SubordinateRate    float64 `json:"subordinate_rate"`
	SuspiciousBalance  float64 `json:"suspicious_balance"` // 可疑
	SuspiciousRate     float64 `json:"suspicious_rate"`
	LossBalance        float64 `json:"loss_balance"` // 损失
	LossRate           float64 `json:"loss_rate"`
}

type BankMigrationRate struct {
	NormalRate      float64 `json:"normal_rate"`      // 正常类迁移至关注类
	NoticeRate      float64 `json:"notice_rate"`      // 关注类迁移至次级
	SubordinateRate float64 `json:"subordinate_rate"` // 次级迁移至可疑
	SuspiciousRate  float64 `json:"suspicious_rate"`  // 可疑迁移至损失
}

type BadDebtAsset struct {
	TotalBalance  float64            `json:"total_balance"`  // 不良余额
	TotalRate     float64            `json:"total_rate"`     // 不良率，(次级类贷款余额 + 可疑类贷款余额 + 损失类贷款余额) / 总贷款余额 × 100%
	NewBalance    float64            `json:"new_balance"`    // 新生成不良余额
	NewRate       float64            `json:"new_rate"`       // 新生成不良率
	BadDebt       *BadDebt           `json:"bad_debt"`       // 不良详情
	MigrationRate *BankMigrationRate `json:"migration_rate"` // 迁徙率
	CoverageRate  float64            `json:"coverage_rate"`  // 拨备覆盖率
}

type TopShareholder struct {
	Name  string  `json:"name"`  // 股东名称
	Share int64   `json:"share"` // 持股数量
	Ratio float64 `json:"ratio"` // 持股比例
}

type ShareholderData struct {
	ShareholderNumber  int               `json:"shareholder_number"` // A+H总的股东数量
	TopShareholderList []*TopShareholder `json:"shareholder_list"`   // 前十大股东
}

type BankReport struct {
	Shareholder     *ShareholderData `json:"shareholder"`
	Income          *BankIncome      `json:"income"`
	Expense         float64          `json:"expense"`          // 业务及管理费 + 税金及附加 + 其他业务成本
	ImpairmentLoss  float64          `json:"impairment_loss"`  // 信用减值损失
	RetainedProfits float64          `json:"retained_profits"` // 归属于本行股东的净利润
	Asset           *BankAsset       `json:"asset"`            // 资产负债表
	BadDebtAsset    *BadDebtAsset    `json:"bad_debt_asset"`   // 不良资产 - 贷款质量分析
	AdequacyRate    float64          `json:"adequacy_rate"`    // 核心一级资本充足率
	ROE             float64          `json:"roe"`              // 净资产收益率
	ROA             float64          `json:"roa"`              // 总资产回报率
	RORWA           float64          `json:"rorwa"`            // 风险加权资产收益率
	preYOYReport    *BankReport      `json:"-"`
	preMOMReport    *BankReport      `json:"-"`
}

type AddStockReportReq struct {
	StockReportBase
	IndustryType IndustryType `json:"industry_type"`
	Report       interface{}  `json:"report"`
}

type AddStockReportResp struct {
	ShareholderMsg string `json:"shareholder_msg"`
	IncomeMsg      string `json:"income_msg"`
}

type GetBankTrackDataReq struct {
	Code string `json:"code" query:"code"`
}

type GetIndustryTrackDataReq struct {
	IndustryType IndustryType `json:"industry_type" query:"industry_type"`
}

type ReportTime struct {
	Year       int `json:"year"`
	ReportType int `json:"report_type"`
}

type ReportTimeSorter []*ReportTime

func (s ReportTimeSorter) Len() int {
	return len(s)
}

func (s ReportTimeSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ReportTimeSorter) GetIndex(year int, reportType int) int {
	for idx, rt := range s {
		if rt.Year == year && rt.ReportType == reportType {
			return idx
		}
	}
	return -1
}

func (s ReportTimeSorter) Less(i, j int) bool {
	if s[i].Year < s[j].Year {
		return true
	} else if s[i].Year > s[j].Year {
		return false
	} else {
		return s[i].ReportType < s[j].ReportType
	}
}

type BankTrackData struct {
	ShareholderNumber  int     `json:"shareholder_number"`
	InterestRate       float64 `json:"interest_rate"`        // 合并净息差
	InterestRatePeriod float64 `json:"interest_rate_period"` // 单季度净息差
	ImpairmentLoss     float64 `json:"impairment_loss"`      // 合并信用减值损失
	TotalBalance       float64 `json:"total_balance"`        // 不良余额
	TotalRate          float64 `json:"total_rate"`           // 不良率，(次级类贷款余额 + 可疑类贷款余额 + 损失类贷款余额) / 总贷款余额 × 100%
	NewBalance         float64 `json:"new_balance"`          // 新生成不良余额
	NewRate            float64 `json:"new_rate"`             // 新生成不良率
	CoverageRate       float64 `json:"coverage_rate"`        // 拨备覆盖率
	AdequacyRate       float64 `json:"adequacy_rate"`        // 核心一级资本充足率
	LoanRate           float64 `json:"loan_rate"`            // 贷款收益率
	LoanRatePeriod     float64 `json:"loan_rate_period"`     // 单季度贷款收益率
	DepositRate        float64 `json:"deposit_rate"`         // 存款成本率
	DepositRatePeriod  float64 `json:"deposit_rate_period"`  // 单季度存款成本率
	ROE                float64 `json:"roe"`                  // 净资产收益率
	ROA                float64 `json:"roa"`                  // 总资产回报率
	RORWA              float64 `json:"rorwa"`                // 风险加权资产收益率
}

type GetBankTrackDataResp struct {
	DateList    []*ReportTime    `json:"date_list"`
	ReportList  []*BankTrackData `json:"report_list"`
	Measurement string           `json:"measurement"`
}

type GetIndustryTrackDataResp struct {
	DateList    []*ReportTime               `json:"date_list"`
	ReportList  map[string][]*BankTrackData `json:"report_map"`
	Measurement string                      `json:"measurement"`
}

func (r *BankReport) Parse(base *StockReportBase) error {
	if err := r.Shareholder.Parse(base); err != nil {
		return err
	}
	if err := r.Income.Parse(base); err != nil {
		return err
	}
	if err := r.Asset.Parse(base); err != nil {
		return err
	}
	if err := r.BadDebtAsset.Parse(base); err != nil {
		return err
	}
	return nil
}

func (r *BankReport) ToString() string {
	data, _ := json.Marshal(r)
	return string(data)
}

func (r *StockReportBase) GetMeasurement() string {
	switch r.Measurement {
	case MeasurementTypeMillion:
		return "百万"
	case MeasurementTypeHundredMillion:
		return "亿元"
	default:
		return ""
	}
}

func (s *ShareholderData) Parse(base *StockReportBase) error {
	if s == nil || len(s.TopShareholderList) == 0 {
		return nil
	}
	// 检查百分比是否因为录入超过了100%
	total := 0.0
	for _, item := range s.TopShareholderList {
		total += item.Ratio
	}
	if total > 100 {
		return fmt.Errorf("前十大股东数据录入异常，总持股比例超过100%")
	}
	// 检查是否数据录入异常，比如股数对应的百分比对不上
	totalShare := float64(s.TopShareholderList[0].Share) * 100 / s.TopShareholderList[0].Ratio
	for i := 1; i < len(s.TopShareholderList); i++ {
		tmpRatio := float64(s.TopShareholderList[i].Share) * 100 / totalShare
		if math.Abs(tmpRatio-s.TopShareholderList[i].Ratio) > 1 {
			return fmt.Errorf("前十大股东数据录入异常，股东名称：%v", s.TopShareholderList[i].Name)
		}
		// 对名称进行处理
		s.TopShareholderList[i].Name = strings.Replace(s.TopShareholderList[i].Name, " ", "", -1)
	}
	return nil
}

func (s *BankIncome) Parse(base *StockReportBase) error {
	if s == nil {
		return nil
	}
	if err := s.InterestIncome.Parse(base); err != nil {
		return err
	}
	diff := s.TotalIncome - s.ServiceIncome - s.NonInterestIncome
	if s.InterestIncome != nil {
		diff = diff - s.InterestIncome.TotalIncome
	}
	if math.Abs(diff) > 1 {
		return fmt.Errorf("收入数据录入异常, 总营业收入不等于净利息收入,手续费收入,其他非息收入的和")
	}
	return nil
}

func (s *BankInterestIncome) Parse(base *StockReportBase) error {
	if s == nil {
		return nil
	}
	if s.LoanRate < s.DepositRate {
		return fmt.Errorf("利息数据录入异常, 贷款收益率小于存款成本率")
	}

	if s.InterestRatePeriod == 0.0 {
		s.FillInterestRatePeriod()
	}
	if s.LoanRatePeriod == 0.0 {
		s.FillInterestRatePeriod()
	}
	if s.DepositRatePeriod == 0.0 {
		s.FillInterestRatePeriod()
	}

	return nil
}

func (s *BankInterestIncome) FillInterestRatePeriod() {
	return
}

func (s *BankInterestIncome) FillLoanRatePeriod() {
	return
}

func (s *BankInterestIncome) FillDepositRatePeriod() {
	return
}

func (s *BankAsset) Parse(base *StockReportBase) error {
	if s == nil {
		return nil
	}
	if err := s.BankLoan.Parse(base); err != nil {
		return err
	}
	if err := s.BankDeposit.Parse(base); err != nil {
		return err
	}
	return nil
}

func (s *BankLoan) Parse(base *StockReportBase) error {
	if s == nil {
		return nil
	}
	return nil
}

func (s *BankDeposit) Parse(base *StockReportBase) error {
	if s == nil {
		return nil
	}
	if s.TotalDeposit < s.CorporateDeposit+s.PersonDeposit {
		return fmt.Errorf("存款数据录入异常, 存款总额小于对公存款,零售存款的和")
	}
	return nil
}

func (s *BadDebtAsset) Parse(base *StockReportBase) error {
	if s == nil {
		return nil
	}
	if err := s.BadDebt.Parse(base); err != nil {
		return err
	}
	if err := s.MigrationRate.Parse(base); err != nil {
		return err
	}
	isDataLoss := true
	if s.BadDebt.SubordinateBalance > 0.0 && s.BadDebt.SuspiciousBalance > 0.0 && s.BadDebt.LossBalance > 0.0 {
		isDataLoss = false
	}
	if s.TotalBalance == 0.0 {
		// 可能存在数据缺失，如果数据缺失的话，就不再计算
		if !isDataLoss {
			s.TotalBalance = s.BadDebt.SubordinateBalance + s.BadDebt.SuspiciousBalance + s.BadDebt.LossBalance
			s.TotalBalance = utils.Float64KeepDecimal(s.TotalBalance, 2)
		}
	} else {
		if !isDataLoss {
			if !utils.Float64Equal(s.BadDebt.SubordinateBalance+s.BadDebt.SuspiciousBalance+s.BadDebt.LossBalance, s.TotalBalance, 2) {
				return fmt.Errorf("不良资产数据录入异常, 不良余额不等于次级类贷款余额,可疑类贷款余额,损失类贷款余额的和")
			}
		}
	}
	if s.TotalRate == 0.0 {
		if !isDataLoss {
			s.TotalRate = utils.Float64KeepDecimal(s.TotalBalance*100/(s.TotalBalance+s.BadDebt.NormalBalance+s.BadDebt.NoticeBalance), 2)
		}
	}
	return nil
}

func (s *StockReportBase) GetYearMsg() string {
	msg := fmt.Sprintf("%v年", s.Year)
	switch s.ReportType {
	case ReportType1:
		msg += "一季报"
	case ReportType2:
		msg += "中报"
	case ReportType3:
		msg += "三季报"
	case ReportType4:
		msg += "年报"
	}
	return msg
}

func (s *BadDebt) Parse(base *StockReportBase) error {
	if s == nil {
		return nil
	}
	// 一季度报，三季度报财报数据披露不全面，会导致数据缺失
	if s.NormalBalance == 0.0 || s.NoticeBalance == 0.0 || s.SubordinateBalance == 0.0 || s.SuspiciousBalance == 0.0 || s.LossBalance == 0.0 {
		return nil
	}
	total := s.NormalBalance + s.NoticeBalance + s.SubordinateBalance + s.SuspiciousBalance + s.LossBalance
	if s.NormalRate == 0.0 {
		s.NormalRate = utils.Float64KeepDecimal(s.NormalBalance/total*100, 2)
	} else {
		rate := s.NormalBalance / total * 100
		if math.Abs(rate-s.NormalRate) > 1 {
			return errors.New("不良资产数据录入异常, 正常类不良余额不等于正常类不良率乘以总不良余额的100%")
		}
	}
	if s.NoticeRate == 0.0 {
		s.NoticeRate = utils.Float64KeepDecimal(s.NoticeBalance/total*100, 2)
	} else {
		rate := s.NoticeBalance / total * 100
		if math.Abs(rate-s.NoticeRate) > 1 {
			return errors.New("不良资产数据录入异常, 关注类不良余额不等于关注类不良率乘以总不良余额的100%")
		}
	}
	if s.SubordinateRate == 0.0 {
		s.SubordinateRate = utils.Float64KeepDecimal(s.SubordinateBalance/total*100, 2)
	} else {
		rate := s.SubordinateBalance / total * 100
		if math.Abs(rate-s.SubordinateRate) > 1 {
			return errors.New("不良资产数据录入异常, 次级类不良余额不等于次级类不良率乘以总不良余额的100%")
		}
	}
	if s.SuspiciousRate == 0.0 {
		s.SuspiciousRate = utils.Float64KeepDecimal(s.SuspiciousBalance/total*100, 2)
	} else {
		rate := s.SuspiciousBalance / total * 100
		if math.Abs(rate-s.SuspiciousRate) > 1 {
			return errors.New("不良资产数据录入异常, 可疑类不良余额不等于可疑类不良率乘以总不良余额的100%")
		}
	}
	if s.LossRate == 0.0 {
		s.LossRate = utils.Float64KeepDecimal(s.LossBalance/total*100, 2)
	} else {
		rate := s.LossBalance / total * 100
		if math.Abs(rate-s.LossRate) > 1 {
			return errors.New("不良资产数据录入异常, 损失类不良余额不等于损失类不良率乘以总不良余额的100%")
		}
	}
	return nil
}

func (s *BankMigrationRate) Parse(base *StockReportBase) error {
	if s == nil {
		return nil
	}
	return nil
}

type GetStockReportResp struct {
	Report       interface{} `json:"report"`
	PreYOYReport interface{} `json:"pre_yoy_report"`
	PreMOMReport interface{} `json:"pre_mom_report"`
	Message      interface{} `json:"message"`
	Measurement  string      `json:"measurement"`
	Comment      string      `json:"comment"`
}

type BankMessage struct {
	ShareholderMsg  string `json:"shareholder_msg"`
	IncomeMsg       string `json:"income_msg"`
	AssetMsg        string `json:"asset_msg"`
	BadDebtMsg      string `json:"bad_debt_msg"`
	CoverageRateMsg string `json:"coverage_rate_msg"`
	AdequacyRateMsg string `json:"adequacy_rate_msg"`
}
