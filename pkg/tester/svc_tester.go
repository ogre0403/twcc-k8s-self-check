package tester

import (
	"errors"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

type SvcTester struct {
	pass bool
	err  error
}

func NewSvcTester(cfg *config.Config) *SvcTester {
	return &SvcTester{
		pass: false,
		err:  errors.New("not implemented"),
	}
}

func (t *SvcTester) Run() Tester {

	return t
}

func (t *SvcTester) Check() Tester {
	return t
}

func (t *SvcTester) Report(report *model.CheckResult) Tester {

	if !t.pass {
		report.SvcCreate = FAIL
		report.ErrorMsg = t.err.Error()
	} else {
		report.SvcCreate = PASS
	}
	return t
}

func (t *SvcTester) Next() bool {
	return t.pass
}

func (t *SvcTester) Close() {}

func (t *SvcTester) String() string {
	return "SvcTester"
}
