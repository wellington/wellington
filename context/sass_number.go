package context

type SassNumber struct {
	value float64
	unit  string
}

var sassUnitConversions = map[string]map[string]float64{
	"in": {
		"in": 1,
		"cm": 2.54,
	},
	"cm": {
		"in": 1.0 / 2.54,
		"cm": 1,
	},
}

func (sn SassNumber) Add(sn2 SassNumber) SassNumber {
	return sn2
}
