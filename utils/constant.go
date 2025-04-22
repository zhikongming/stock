package utils

type MALineType int
type Tendency string

const (
	LocalStorageDir = "data"
	StockBasicFile  = "basic.json"
	StockDailyFile  = "daily.json"

	MALineToUp    MALineType = 0
	MALineToDown  MALineType = 1
	MALineLess5   MALineType = 2
	MALineLess10  MALineType = 3
	MALineLess20  MALineType = 4
	MALineLess30  MALineType = 5
	MALineLess60  MALineType = 6
	MALineGreat60 MALineType = 7

	TendencyUp     Tendency = "up"
	TendencyDown   Tendency = "down"
	TendencyMiddle Tendency = "middle"
)

func GetMALineString(lineType MALineType) string {
	value := ""
	if lineType&MALineToUp == MALineToUp {
		value += "up"
	} else {
		value += "down"
	}
	if lineType&MALineLess5 == MALineLess5 {
		value += "P<5"
	} else if lineType&MALineLess10 == MALineLess10 {
		value += "5<P<10"
	} else if lineType&MALineLess20 == MALineLess20 {
		value += "10<p<20"
	} else if lineType&MALineLess30 == MALineLess30 {
		value += "20<P<30"
	} else if lineType&MALineLess60 == MALineLess60 {
		value += "30<P<60"
	} else if lineType&MALineGreat60 == MALineGreat60 {
		value += "P>60"
	}
	return value
}
