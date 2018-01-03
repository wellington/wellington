package libs

import (
	"math"
	"testing"
)

var tolerance = float64(0.00001)

func TestSassNumberAddDifferentUnits(t *testing.T) {
	var sn1 = SassNumber{50, "px"}
	var sn2 = SassNumber{15, "cm"}

	res := sn1.Add(sn2)

	expectedValue := 50 + ((96.0 / 2.54) * 15)
	if res.Unit != sn1.Unit {
		t.Errorf("SassNumber Add result Units are: %s, wanted %s", res.Unit, sn1.Unit)
	} else if !compareFloats(res.Value, expectedValue) {
		t.Errorf("SassNumber Add result Value expected: %f, got %f", expectedValue, res.Value)
	}
}

func TestSassNumberAddSameUnits(t *testing.T) {
	var sn1 = SassNumber{80.0, "mm"}
	var sn2 = SassNumber{25.0, "mm"}

	res := sn1.Add(sn2)

	expectedValue := 80.0 + 25.0
	if res.Unit != sn1.Unit {
		t.Errorf("SassNumber Add result Units are: %s, wanted %s", res.Unit, sn1.Unit)
	} else if !compareFloats(res.Value, expectedValue) {
		t.Errorf("SassNumber Add result Value expected: %f, got %f", expectedValue, res.Value)
	}
}

func TestSassNumberSubstractDifferentUnits(t *testing.T) {
	var sn1 = SassNumber{60, "grad"}
	var sn2 = SassNumber{25, "deg"}

	res := sn1.Subtract(sn2)

	expectedValue := 60 - ((40.0 / 36.0) * 25.0)
	if res.Unit != sn1.Unit {
		t.Errorf("SassNumber Subtract result Units are: %s, wanted %s", res.Unit, sn1.Unit)
	} else if !compareFloats(res.Value, expectedValue) {
		t.Errorf("SassNumber Subtract result Value expected: %f, got %f", expectedValue, res.Value)
	}
}

func TestSassNumberSubtractSameUnits(t *testing.T) {
	var sn1 = SassNumber{80.0, "mm"}
	var sn2 = SassNumber{25.0, "mm"}

	res := sn1.Subtract(sn2)

	expectedValue := 80.0 - 25.0
	if res.Unit != sn1.Unit {
		t.Errorf("SassNumber Subtract result Units are: %s, wanted %s", res.Unit, sn1.Unit)
	} else if !compareFloats(res.Value, expectedValue) {
		t.Errorf("SassNumber Subtract result Value expected: %f, got %f", expectedValue, res.Value)
	}
}

func TestSassNumberMultiplyDifferentUnits(t *testing.T) {
	var sn1 = SassNumber{15, "mm"}
	var sn2 = SassNumber{5, "pt"}

	res := sn1.Multiply(sn2)

	expectedValue := 15 * ((25.4 / 72.0) * 5)
	if res.Unit != sn1.Unit {
		t.Errorf("SassNumber Multiply result Units are: %s, wanted %s", res.Unit, sn1.Unit)
	} else if !compareFloats(res.Value, expectedValue) {
		t.Errorf("SassNumber Multiply result Value expected: %f, got %f", expectedValue, res.Value)
	}
}

func TestSassNumberMultiplySameUnits(t *testing.T) {
	var sn1 = SassNumber{.4, "rad"}
	var sn2 = SassNumber{.7, "rad"}

	res := sn1.Multiply(sn2)

	expectedValue := .4 * .7
	if res.Unit != sn1.Unit {
		t.Errorf("SassNumber add result Units are: %s, wanted %s", res.Unit, sn1.Unit)
	} else if !compareFloats(res.Value, expectedValue) {
		t.Errorf("SassNumber Add result Value expected: %f, got %f", expectedValue, res.Value)
	}
}

func TestSassNumberDivideDifferentUnits(t *testing.T) {
	var sn1 = SassNumber{5, "in"}
	var sn2 = SassNumber{15, "px"}

	res := sn1.Divide(sn2)

	expectedValue := 5 / ((1.0 / 96.0) * 15)
	if res.Unit != sn1.Unit {
		t.Errorf("SassNumber Divide result Units are: %s, wanted %s", res.Unit, sn1.Unit)
	} else if !compareFloats(res.Value, expectedValue) {
		t.Errorf("SassNumber Divide result Value expected: %f, got %f", expectedValue, res.Value)
	}
}

func TestSassNumberDivideSameUnits(t *testing.T) {
	var sn1 = SassNumber{80.0, "cm"}
	var sn2 = SassNumber{25.0, "cm"}

	res := sn1.Divide(sn2)

	expectedValue := 80.0 / 25.0
	if res.Unit != sn1.Unit {
		t.Errorf("SassNumber Divide result Units are: %s, wanted %s", res.Unit, sn1.Unit)
	} else if !compareFloats(res.Value, expectedValue) {
		t.Errorf("SassNumber Divide result Value expected: %f, got %f", expectedValue, res.Value)
	}
}

/*
func TestUnknownUnit(t *testing.T) {
	var sn1 = SassNumber{80.0, "mm"}
	var sn2 = SassNumber{25.0, "TalorSwift"}

	_, err := sn1.Divide(sn2)

	if err == nil {
		t.Errorf("Wanted: %s but did not get an error", fmt.Sprintf("Can not convert from %s to %s", sn2.Unit, sn1.Unit))
	} else if err.Error() != fmt.Sprintf("Can not convert from %s to %s", sn2.Unit, sn1.Unit) {
		t.Errorf("Wanted: %s got: %s", fmt.Sprintf("Can not convert from %s to %s", sn2.Unit, sn1.Unit), err.Error())
	}
}

func TestDistanceToAngleConversion(t *testing.T) {
	var sn1 = SassNumber{80.0, "mm"}
	var sn2 = SassNumber{25.0, "rad"}

	_, err := sn1.Divide(sn2)

	if err == nil {
		t.Errorf("Wanted: %s but did not get an error", fmt.Sprintf("Can not convert sass Units between angles and distances: %s, %s", sn2.Unit, sn1.Unit))
	} else if err.Error() != fmt.Sprintf("Can not convert sass Units between angles and distances: %s, %s", sn2.Unit, sn1.Unit) {
		t.Errorf("Wanted: %s got: %s", fmt.Sprintf("Can not convert sass Units between angles and distances: %s, %s", sn2.Unit, sn1.Unit), err.Error())
	}
}
*/

func TestChainedOperation(t *testing.T) {
	var sn1 = SassNumber{5, "in"}
	var sn2 = SassNumber{15, "px"}
	var sn3 = SassNumber{55, "px"}
	var sn4 = SassNumber{75, "mm"}
	var sn5 = SassNumber{25, "pt"}

	res := sn1.Add(sn2).Subtract(sn3).Multiply(sn4).Divide(sn5)

	expectedValue := (((5.0 + ((1.0 / 96.0) * 15)) - ((1.0 / 96.0) * 55)) * ((1.0 / 25.4) * 75)) / ((1.0 / 72.0) * 25)

	if res.Unit != sn1.Unit {
		t.Errorf("SassNumber Divide result Units are: %s, wanted %s", res.Unit, sn1.Unit)
	} else if !compareFloats(res.Value, expectedValue) {
		t.Errorf("SassNumber chained operation result Value expected: %f, got %f", expectedValue, res.Value)
	}
}

func compareFloats(f1 float64, f2 float64) bool {
	if math.Abs(f1-f2) < tolerance {
		return true
	}
	return false
}
