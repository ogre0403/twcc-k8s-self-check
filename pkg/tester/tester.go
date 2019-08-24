package tester

import (
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

const (
	PASS = "PASS"
	FAIL = "FAIL"
)

type Tester interface {
	// Run Test case
	Run() Tester

	// Check Test result
	Check() Tester

	// fill report
	Report(*model.CheckResult) Tester

	// if need to run next step
	Next() bool

	// close opened resource by tester
	Close()
}
