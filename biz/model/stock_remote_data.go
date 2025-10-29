package model

import (
	"strconv"
	"time"

	"github.com/zhikongming/stock/utils"
)

type GetRemoteStockBasicResp struct {
	ErrorCode        int             `json:"error_code"`
	ErrorDescription string          `json:"error_description"`
	Data             *StockBasicData `json:"data"`
}

type StockBasicData struct {
	Company *StockBasicDataCompany `json:"company"`
}

type StockBasicDataCompany struct {
	OrgID                    string                               `json:"org_id"`
	OrgNameCN                string                               `json:"org_name_cn"`
	OrgShortNameCN           string                               `json:"org_short_name_cn"`
	OrgNameEN                string                               `json:"org_name_en"`
	OrgShortNameEN           string                               `json:"org_short_name_en"`
	MainPperationBusiness    string                               `json:"main_operation_business"`
	OperatingScope           string                               `json:"operating_scope"`
	DistrictEncode           string                               `json:"district_encode"`
	OrgCnIntroduction        string                               `json:"org_cn_introduction"`
	LegalRepresentative      string                               `json:"legal_representative"`
	GeneralManager           string                               `json:"general_manager"`
	Secretary                string                               `json:"secretary"`
	EstablishedDate          int64                                `json:"established_date"`
	RegAsset                 float64                              `json:"reg_asset"`
	StaffNum                 int                                  `json:"staff_num"`
	Telephone                string                               `json:"telephone"`
	Postcode                 string                               `json:"postcode"`
	Fax                      string                               `json:"fax"`
	Email                    string                               `json:"email"`
	OrgWebsite               string                               `json:"org_website"`
	RegAddressCN             string                               `json:"reg_address_cn"`
	RegAddressEN             string                               `json:"reg_address_en"`
	OfficeAddressCN          string                               `json:"office_address_cn"`
	OfficeAddressEN          string                               `json:"office_address_en"`
	CurrencyEncode           string                               `json:"currency_encode"`
	Currency                 string                               `json:"currency"`
	ListedDate               int64                                `json:"listed_date"`
	ProvincialName           string                               `json:"provincial_name"`
	ActualController         string                               `json:"actual_controller"`
	ClassiName               string                               `json:"classi_name"`
	PreNameCN                string                               `json:"pre_name_cn"`
	Chairman                 string                               `json:"chairman"`
	ExecutivesNums           int                                  `json:"executives_nums"`
	ActualIssueVol           float64                              `json:"actual_issue_vol"`
	IssuePrice               float64                              `json:"issue_price"`
	ActualRCNetAmt           float64                              `json:"actual_rc_net_amt"`
	PeAfterIssuing           float64                              `json:"pe_after_issuing"`
	OnlineSuccessRateOfIssue float64                              `json:"online_success_rate_of_issue"`
	AffiliateIndustry        *StockBasicRespDataAffiliateIndustry `json:"affiliate_industry"`
}

type StockBasicRespDataAffiliateIndustry struct {
	IndCode string `json:"ind_code"`
	IndName string `json:"ind_name"`
}

type GetRemoteStockDailyResp struct {
	ErrorCode        int             `json:"error_code"`
	ErrorDescription string          `json:"error_description"`
	Data             *StockDailyData `json:"data"`
}

type StockDailyData struct {
	Symbol    string          `json:"symbol"`
	Column    []string        `json:"column"`
	Item      [][]interface{} `json:"item"`
	ColumnMap map[string]int
}

func (d *StockDailyData) GetColumnIndexByKey(keyName string) int {
	if d.ColumnMap == nil {
		d.ColumnMap = make(map[string]int)
		for index, key := range d.Column {
			d.ColumnMap[key] = index
		}
	}
	if value, ok := d.ColumnMap[keyName]; ok {
		return value
	}
	return -1
}

func (d *StockDailyData) ToDatePriceList() []*DatePrice {
	datePriceList := make([]*DatePrice, 0)
	for _, item := range d.Item {
		timestampIndex := d.GetColumnIndexByKey("timestamp")
		timestamp, _ := strconv.ParseInt(utils.ToString(item[timestampIndex]), 10, 64)
		date := utils.TimestampToDate(timestamp / int64(time.Microsecond))
		datePriceList = append(datePriceList, &DatePrice{
			Date:  date,
			Price: utils.Float64KeepDecimal(utils.ToFloat64(item[d.GetColumnIndexByKey("close")]), 2),
		})
	}
	return datePriceList
}

type StockRelationResp struct {
	ErrorCode        int                `json:"error_code"`
	ErrorDescription string             `json:"error_description"`
	Data             *StockRelationData `json:"data"`
}

type StockRelationData struct {
	StockItemList []*StockRelationItem `json:"item"`
}

type StockRelationItem struct {
	Symbol   string  `json:"symbol"`
	Current  float64 `json:"current"`
	Chg      float64 `json:"chg"`
	Premium  float64 `json:"premium"`
	Name     string  `json:"name"`
	Type     int     `json:"type"`
	Percent  float64 `json:"percent"`
	TickSize float64 `json:"tick_size"`
}
