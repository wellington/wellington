package context

import (
	"math"
	"testing"
)

var tolerance = float64(0.00001)

func TestSassNumberAddDifferentUnits(t *testing.T) {
	var sn1 = SassNumber{50, "px"}
	var sn2 = SassNumber{15, "cm"}

	res := sn1.Add(sn2)

	expectedValue := ((96.0 / 2.54) * 15) + 50
	if res.unit != "px" {
		t.Errorf("SassNumber add result units are: %s, wanted %s", res.unit, sn1.unit)
	} else if !compareFloats(res.value, expectedValue) {
		t.Errorf("SassNumber Add result value expected: %f, got %f", expectedValue, res.value)
	}
}

func TestSassNumberAddSameUnits(t *testing.T) {
	var sn1 = SassNumber{80.0, "mm"}
	var sn2 = SassNumber{25.0, "mm"}

	res := sn1.Add(sn2)

	expectedValue := 105.0
	if res.unit != "mm" {
		t.Errorf("SassNumber add result units are: %s, wanted %s", res.unit, sn1.unit)
	} else if !compareFloats(res.value, expectedValue) {
		t.Errorf("SassNumber Add result value expected: %f, got %f", expectedValue, res.value)
	}
}

func compareFloats(f1 float64, f2 float64) bool {
	if math.Abs(f1-f2) < tolerance {
		return true
	} else {
		return false
	}
}
