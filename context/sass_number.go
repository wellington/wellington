package context

import (
	"fmt"
	"math"
)

type SassNumber struct {
	value float64
	unit  string
}

var sassUnitConversions = map[string]map[string]float64{
	"in": {
		"in":   1,
		"cm":   2.54,
		"pc":   6,
		"mm":   25.4,
		"pt":   72,
		"px":   96,
		"deg":  1,
		"grad": 1,
		"rad":  1,
		"turn": 1,
	},
	"cm": {
		"in":   1.0 / 2.54,
		"cm":   1,
		"pc":   6.0 / 2.54,
		"mm":   10,
		"pt":   72.0 / 2.54,
		"px":   96.0 / 2.54,
		"deg":  1,
		"grad": 1,
		"rad":  1,
		"turn": 1,
	},
	"pc": {
		"in":   1.0 / 6.0,
		"cm":   2.54 / 6.0,
		"pc":   1,
		"mm":   25.4 / 6.0,
		"pt":   72.0 / 6.0,
		"px":   96.0 / 6.0,
		"deg":  1,
		"grad": 1,
		"rad":  1,
		"turn": 1,
	},
	"mm": {
		"in":   1.0 / 25.4,
		"cm":   1.0 / 10.0,
		"pc":   6.0 / 25.4,
		"mm":   1,
		"pt":   72.0 / 25.4,
		"px":   96.0 / 25.4,
		"deg":  1,
		"grad": 1,
		"rad":  1,
		"turn": 1,
	},
	"pt": {
		"in":   1.0 / 72.0,
		"cm":   2.54 / 72.0,
		"pc":   6.0 / 72.0,
		"mm":   25.4 / 72.0,
		"pt":   1,
		"px":   96.0 / 72.0,
		"deg":  1,
		"grad": 1,
		"rad":  1,
		"turn": 1,
	},
	"px": {
		"in":   1.0 / 96.0,
		"cm":   2.54 / 96.0,
		"pc":   6.0 / 96.0,
		"mm":   25.4 / 96.0,
		"pt":   72.0 / 96.0,
		"px":   1,
		"deg":  1,
		"grad": 1,
		"rad":  1,
		"turn": 1,
	},
	"deg": {
		"in":   1,
		"cm":   1,
		"pc":   1,
		"mm":   1,
		"pt":   1,
		"px":   1,
		"deg":  1,
		"grad": 40.0 / 36.0,
		"rad":  math.Pi / 180.0,
		"turn": 1.0 / 360.0,
	},
	"grad": {
		"in":   1,
		"cm":   1,
		"pc":   1,
		"mm":   1,
		"pt":   1,
		"px":   1,
		"deg":  36.0 / 40.0,
		"grad": 1,
		"rad":  math.Pi / 200.0,
		"turn": 1.0 / 400.0,
	},
	"rad": {
		"in":   1,
		"cm":   1,
		"pc":   1,
		"mm":   1,
		"pt":   1,
		"px":   1,
		"deg":  180.0 / math.Pi,
		"grad": 200.0 / math.Pi,
		"rad":  1,
		"turn": math.Pi / 2.0,
	},
	"turn": {
		"in":   1,
		"cm":   1,
		"pc":   1,
		"mm":   1,
		"pt":   1,
		"px":   1,
		"deg":  360.0,
		"grad": 400.0,
		"rad":  2.0 * math.Pi,
		"turn": 1,
	},
}

var sassUnitTypes = map[string]string{
	"in":   "distance",
	"cm":   "distance",
	"pc":   "distance",
	"mm":   "distance",
	"pt":   "distance",
	"px":   "distance",
	"deg":  "angle",
	"grad": "angle",
	"rad":  "angle",
	"turn": "angle",
}

func (sn SassNumber) Add(sn2 SassNumber) SassNumber {
	sn1Value, sn2Value := getConvertedUnits(sn, sn2)
	return SassNumber{value: sn1Value + sn2Value, unit: sn.unit}
}

func (sn SassNumber) Subtract(sn2 SassNumber) SassNumber {
	sn1Value, sn2Value := getConvertedUnits(sn, sn2)
	return SassNumber{value: sn1Value - sn2Value, unit: sn.unit}
}

func (sn SassNumber) Multiply(sn2 SassNumber) SassNumber {
	sn1Value, sn2Value := getConvertedUnits(sn, sn2)
	return SassNumber{value: sn1Value * sn2Value, unit: sn.unit}
}

func (sn SassNumber) Divide(sn2 SassNumber) SassNumber {
	sn1Value, sn2Value := getConvertedUnits(sn, sn2)
	return SassNumber{value: sn1Value / sn2Value, unit: sn.unit}
}

func getConvertedUnits(sn1 SassNumber, sn2 SassNumber) (float64, float64) {
	var sn2Value float64
	if sn2.unit != sn1.unit {
		sn2Value = convertUnits(sn2, sn1)
	} else {
		sn2Value = sn2.value
	}

	return sn1.value, sn2Value
}

func convertUnits(from SassNumber, to SassNumber) float64 {
	return sassUnitConversions[from.unit][to.unit] * from.value
}

func safeConvert(from SassNumber, to SassNumber) error {
	if _, ok := sassUnitConversions[from.unit][to.unit]; !ok {
		return fmt.Errorf("Can not convert from %s to %s", from.unit, to.unit)
	}

	if sassUnitTypes[from.unit] != sassUnitTypes[to.unit] {
		return fmt.Errorf("Can not convert sass units between angles and distances: %s, %s", from.unit, to.unit)
	}

	return nil
}
