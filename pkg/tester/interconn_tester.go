package tester

import (
	"errors"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

type InterConnTester struct {
	pass bool
	err  error
}

func NewInterConnTester(cfg *config.Config) *InterConnTester {
	return &InterConnTester{
		pass: false,
		err:  errors.New("not implemented"),
	}
}

func (t *InterConnTester) Run() Tester {

	return t
}

func (t *InterConnTester) Check() Tester {
	return t
}

func (t *InterConnTester) Report(report *model.CheckResult) Tester {

	if !t.pass {
		report.InternetConnection = FAIL
		report.ErrorMsg = t.err.Error()
	} else {
		report.InternetConnection = PASS
	}
	return t
}

func (t *InterConnTester) Next() bool {
	return t.pass
}

func (t *InterConnTester) Close() {}

func (t *InterConnTester) String() string {
	return "InterConnTester"
}
