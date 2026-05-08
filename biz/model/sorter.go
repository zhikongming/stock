package model

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
