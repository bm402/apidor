package testcode

import (
	"errors"
	"strings"
)

// TestCodes is a testcode type that carries the set of tests to be executed
type TestCodes []TestCode

// TestCode is a testcode type that represents a type of test
type TestCode string

const (
	// HP is a TestCode const for a high privileged test
	HP TestCode = "hp"
	// LP is a TestCode const for a low privileged permutations test
	LP = "lp"
	// NP is a TestCode const for a no privileged test
	NP = "np"
	// RPP is a TestCode const for a request parameter pollution test
	RPP = "rpp"
	// BPP is a TestCode const for a body parameter pollution test
	BPP = "bpp"
	// MR is a TestCode const for a method replacement test
	MR = "mr"
	// RPW is a TestCode const for a request parameter wrapping test
	RPW = "rpw"
	// BPW is a TestCode const for a body parameter wrapping test
	BPW = "bpw"
	// RPS is a TestCode const for a request parameter substitution test
	RPS = "rps"
	// RPSPP is a TestCode const for a request parameter substitution test with parameter pollution
	RPSPP = "rpspp"
	// JSON is a TestCode const for a .json test
	JSON = "json"
	// ALL is a TestCode const for all tests
	ALL = "all"
)

// ParseTestCodes is a testcode function that parses a comma-separated string of test codes
// into a slice of test codes
func ParseTestCodes(str string) (TestCodes, error) {
	testCodes := TestCodes{}
	errs := ""
	testCodeStrings := strings.Split(str, ",")
	for _, testCodeString := range testCodeStrings {
		testCodeString = strings.TrimSpace(testCodeString)
		if testCode, ok := getTestCode(testCodeString); ok {
			testCodes = append(testCodes, testCode)
		} else {
			errs += testCodeString + ", "
		}
	}
	if len(errs) > 0 {
		return testCodes, errors.New("Unrecognised test codes: " + errs[:len(errs)-2])
	}
	return testCodes, nil
}

// Contains is a TestCodes function for checking whether a test code is included in the tests
func (tcs TestCodes) Contains(testCode TestCode) bool {
	for _, test := range tcs {
		if test == testCode {
			return true
		}
	}
	return false
}

func getTestCode(testCodeString string) (TestCode, bool) {
	testCode := TestCode(testCodeString)
	switch testCode {
	case HP, LP, NP, RPP, BPP, MR, RPW, BPW, RPS, RPSPP, JSON, ALL:
		return testCode, true
	}
	return "", false
}
