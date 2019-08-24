package tester

import (
	"errors"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

type IntraConnTester struct {
	pass bool
	err  error
}

func NewIntraConnTester(cfg *config.Config) *IntraConnTester {
	return &IntraConnTester{
		pass: false,
		err:  errors.New("not implemented"),
	}
}

func (t *IntraConnTester) Run() Tester {

	return t
}

func (t *IntraConnTester) Check() Tester {
	return t
}

func (t *IntraConnTester) Report(report *model.CheckResult) Tester {

	if !t.pass {
		report.IntranetConnection = FAIL
		report.ErrorMsg = t.err.Error()
	} else {
		report.IntranetConnection = PASS
	}
	return t
}

func (t *IntraConnTester) Next() bool {
	return t.pass
}

func (t *IntraConnTester) Close() {}

func (t *IntraConnTester) String() string {
	return "IntraConnTester"
}
