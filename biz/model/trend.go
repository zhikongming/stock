package model

type TrendType int
type ClassType int
type DivergencePointType int

const (
	TrendUp     TrendType = 1
	TrendDown   TrendType = -1
	TrendUnkown TrendType = 0

	ClassTop    ClassType = 1
	ClassBottom ClassType = 0

	DivergencePointBuy1  DivergencePointType = 1
	DivergencePointBuy2  DivergencePointType = 2
	DivergencePointBuy3  DivergencePointType = 3
	DivergencePointSell1 DivergencePointType = -1
	DivergencePointSell2 DivergencePointType = -2
	DivergencePointSell3 DivergencePointType = -3
)

type TrendRange struct {
	StartIndex int       `json:"start_index"`
	EndIndex   int       `json:"end_index"`
	Trend      TrendType `json:"trend"`
}

type FractalInterval struct {
	StartIndex int       `json:"start_index"`
	EndIndex   int       `json:"end_index"`
	Class      ClassType `json:"class"`
}

type PivotInterval struct {
	StartIndex int     `json:"start_index"`
	PriceHigh  float64 `json:"price_high"`
	EndIndex   int     `json:"end_index"`
	PriceLow   float64 `json:"price_low"`
}

type DivergencePoint struct {
	Index     int                 `json:"index"`
	PointType DivergencePointType `json:"point_type"`
}

func (d DivergencePointType) ToString() string {
	switch d {
	case DivergencePointBuy1:
		return "B1"
	case DivergencePointBuy2:
		return "B2"
	case DivergencePointBuy3:
		return "B3"
	case DivergencePointSell1:
		return "S1"
	case DivergencePointSell2:
		return "S2"
	case DivergencePointSell3:
		return "S3"
	default:
		return "no"
	}
}
