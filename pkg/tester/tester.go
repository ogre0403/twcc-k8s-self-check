package tester

import (
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

const (
	PASS = "PASS"
	FAIL = "FAIL"
)

type Tester interface {
	Run() Tester
	Report(*model.CheckResult) Tester
	Next() bool
}
