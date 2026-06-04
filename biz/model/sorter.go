package model

import "github.com/zhikongming/stock/utils"

type VolumeReportItemSorter []*VolumeReportItem

func (s VolumeReportItemSorter) Len() int {
	return len(s)
}

func (s VolumeReportItemSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s VolumeReportItemSorter) Less(i, j int) bool {
	return s[i].Diff > s[j].Diff
}

type LimitUpReportItemSorter []*LimitUpReportItem

func (s LimitUpReportItemSorter) Len() int {
	return len(s)
}

func (s LimitUpReportItemSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s LimitUpReportItemSorter) Less(i, j int) bool {
	return s[i].IndustryName < s[j].IndustryName
}

type ConceptRespNameSorter []*ConceptResp

func (s ConceptRespNameSorter) Len() int {
	return len(s)
}
func (s ConceptRespNameSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ConceptRespNameSorter) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

type ConceptRespChangeSorter []*ConceptResp

func (s ConceptRespChangeSorter) Len() int {
	return len(s)
}
func (s ConceptRespChangeSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ConceptRespChangeSorter) Less(i, j int) bool {
	return s[i].Percent > s[j].Percent
}

type UnusualStockSorter []*UnusualStock

func (s UnusualStockSorter) Len() int {
	return len(s)
}
func (s UnusualStockSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s UnusualStockSorter) Less(i, j int) bool {
	iEndDate := utils.ParseDate(s[i].EndDate)
	jEndDate := utils.ParseDate(s[j].EndDate)
	if iEndDate == jEndDate {
		return s[i].Type > s[j].Type
	}
	return iEndDate.Before(jEndDate)
}
