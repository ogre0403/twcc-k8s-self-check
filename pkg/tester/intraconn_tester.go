package tester

import (
	"errors"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

type IntraConnTester struct {
	pass bool
	err  error
}

func (t *IntraConnTester) Run() Tester {

	//time.Sleep(10 * time.Second)
	t.pass = false
	t.err = errors.New("not implemented")
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
