package service

import (
	"context"
	"strings"

	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

type ShareholderFilterType string

const (
	ShareholderFilterTypeIncrease ShareholderFilterType = "increase"
	ShareholderFilterTypeDecrease ShareholderFilterType = "decrease"
	ShareholderFilterTypeExist    ShareholderFilterType = "exist"

	ShareholderNameForNumber = "股东人数"
	ShareholderNameForGJD    = "国家队"
)

var NameForGJDList = []string{"中国证券金融股份有限公司", "中央汇金资产管理有限责任公司"}

func GetShareholderName(name string) []string {
	if ShareholderNameForGJD == name {
		return NameForGJDList
	}
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return []string{}
	}
	return []string{name}
}

type ShareholderFilter interface {
	Filter(ctx context.Context, report *model.ShareholderAnalysisReport) bool
}

func GetShareholderFilter(req *model.ShareholderReportReq) ShareholderFilter {
	switch ShareholderFilterType(req.Operation) {
	case ShareholderFilterTypeIncrease:
		return &ShareholderIncreaseFilter{Object: req.Object}
	case ShareholderFilterTypeDecrease:
		return &ShareholderDecreaseFilter{Object: req.Object}
	case ShareholderFilterTypeExist:
		return &ShareholderExistFilter{Object: req.Object}
	default:
		return nil
	}
}

type ShareholderIncreaseFilter struct {
	Object string
}

func (s *ShareholderIncreaseFilter) Filter(ctx context.Context, report *model.ShareholderAnalysisReport) bool {
	// 有两种可能, 一个是股东人数, 一个是股东持仓
	if ShareholderNameForNumber == s.Object {
		return report.ShareholderNumberDiff > 0
	} else {
		nameList := GetShareholderName(s.Object)
		if len(nameList) == 0 {
			return false
		}
		// 多个名称的, 得用总量来判断
		matchDiff := 0.0
		for _, name := range nameList {
			matchShareholderList := []*model.ShareholderWithDiff{}
			for idx, topShareholder := range report.TopShareholderList {
				if !strings.Contains(topShareholder.ShareholderName, name) {
					continue
				}
				matchShareholderList = append(matchShareholderList, topShareholder)
				if topShareholder.Diff == "不变" {
				} else if topShareholder.Diff == "新进" {
					// 这里不知道增加了多少其实, 只能预估一下
					if idx == len(report.TopShareholderList)-1 {
						// 在最后一个, 所以直接认定为增加为当前持仓
						curShareholderNumber, curShareholderUnit := utils.GetShareholderNumberUnit(topShareholder.ShareholderNumber)
						curShareholderNumber = curShareholderNumber * float64(utils.GetGetShareholderNumberByUnit(curShareholderUnit))
						matchDiff += curShareholderNumber
					} else {
						// 只能和最后一个对比, 能知道至少增加了多少.
						curShareholderNumber, curShareholderUnit := utils.GetShareholderNumberUnit(topShareholder.ShareholderNumber)
						lastShareholderNumber, lastShareholderUnit := utils.GetShareholderNumberUnit(report.TopShareholderList[len(report.TopShareholderList)-1].ShareholderNumber)
						curShareholderNumber = curShareholderNumber * float64(utils.GetGetShareholderNumberByUnit(curShareholderUnit))
						lastShareholderNumber = lastShareholderNumber * float64(utils.GetGetShareholderNumberByUnit(lastShareholderUnit))
						matchDiff += curShareholderNumber - lastShareholderNumber
					}
				} else {
					// 直接计算数据
					curShareholderNumber, curShareholderUnit := utils.GetShareholderNumberUnit(topShareholder.Diff)
					curShareholderNumber = curShareholderNumber * float64(utils.GetGetShareholderNumberByUnit(curShareholderUnit))
					matchDiff += curShareholderNumber
				}
			}
			// 在计算一下退出前十大股东的, 我们只能假定退出的股东持仓为0
			for _, delShareholder := range report.DelShareholderList {
				if strings.Contains(delShareholder.ShareholderName, name) {
					curShareholderNumber, curShareholderUnit := utils.GetShareholderNumberUnit(delShareholder.Diff)
					curShareholderNumber = curShareholderNumber * float64(utils.GetGetShareholderNumberByUnit(curShareholderUnit))
					matchDiff -= curShareholderNumber
				}
			}
		}
		if matchDiff > 0 {
			return true
		}
		return false
	}
}

type ShareholderDecreaseFilter struct {
	Object string
}

func (s *ShareholderDecreaseFilter) Filter(ctx context.Context, report *model.ShareholderAnalysisReport) bool {
	// 有两种可能, 一个是股东人数, 一个是股东持仓
	if ShareholderNameForNumber == s.Object {
		return report.ShareholderNumberDiff < 0
	} else {
		nameList := GetShareholderName(s.Object)
		if len(nameList) == 0 {
			return false
		}
		// 多个名称的, 得用总量来判断
		matchDiff := 0.0
		for _, name := range nameList {
			matchShareholderList := []*model.ShareholderWithDiff{}
			for idx, topShareholder := range report.TopShareholderList {
				if !strings.Contains(topShareholder.ShareholderName, name) {
					continue
				}
				matchShareholderList = append(matchShareholderList, topShareholder)
				if topShareholder.Diff == "不变" {
				} else if topShareholder.Diff == "新进" {
					// 这里不知道增加了多少其实, 只能预估一下
					if idx == len(report.TopShareholderList)-1 {
						// 在最后一个, 所以直接认定为增加为0
					} else {
						// 只能和最后一个对比, 能知道至少增加了多少.
						curShareholderNumber, curShareholderUnit := utils.GetShareholderNumberUnit(topShareholder.ShareholderNumber)
						lastShareholderNumber, lastShareholderUnit := utils.GetShareholderNumberUnit(report.TopShareholderList[len(report.TopShareholderList)-1].ShareholderNumber)
						curShareholderNumber = curShareholderNumber * float64(utils.GetGetShareholderNumberByUnit(curShareholderUnit))
						lastShareholderNumber = lastShareholderNumber * float64(utils.GetGetShareholderNumberByUnit(lastShareholderUnit))
						matchDiff += curShareholderNumber - lastShareholderNumber
					}
				} else {
					// 直接计算数据
					curShareholderNumber, curShareholderUnit := utils.GetShareholderNumberUnit(topShareholder.Diff)
					curShareholderNumber = curShareholderNumber * float64(utils.GetGetShareholderNumberByUnit(curShareholderUnit))
					matchDiff += curShareholderNumber
				}
			}
			// 在计算一下退出前十大股东的, 我们只能假定退出的股东持仓为0
			for _, delShareholder := range report.DelShareholderList {
				if strings.Contains(delShareholder.ShareholderName, name) {
					curShareholderNumber, curShareholderUnit := utils.GetShareholderNumberUnit(delShareholder.Diff)
					curShareholderNumber = curShareholderNumber * float64(utils.GetGetShareholderNumberByUnit(curShareholderUnit))
					matchDiff -= curShareholderNumber
				}
			}
		}
		if matchDiff < 0 {
			return true
		}
		return false
	}
}

type ShareholderExistFilter struct {
	Object string
}

func (s *ShareholderExistFilter) Filter(ctx context.Context, report *model.ShareholderAnalysisReport) bool {
	nameList := GetShareholderName(s.Object)
	if len(nameList) == 0 {
		return false
	}
	// 多个名称的, 任何一个存在都可以
	for _, name := range nameList {
		for _, topShareholder := range report.TopShareholderList {
			if strings.Contains(topShareholder.ShareholderName, name) {
				return true
			}
		}
	}
	return false
}
