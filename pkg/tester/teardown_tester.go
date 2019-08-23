package tester

import (
	"errors"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

type TeardownTester struct {
	pass bool
	err  error
}

func (t *TeardownTester) Run() Tester {

	//time.Sleep(10 * time.Second)
	t.pass = false
	t.err = errors.New("not implemented")
	return t
}

func (t *TeardownTester) Report(report *model.CheckResult) Tester {

	if !t.pass {
		report.Teardown = FAIL
		report.ErrorMsg = t.err.Error()
	} else {
		report.Teardown = PASS
	}
	return t
}

func (t *TeardownTester) Next() bool {
	return t.pass
}
