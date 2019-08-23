package tester

import (
	"errors"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

type InterConnTester struct {
	pass bool
	err  error
}

func (t *InterConnTester) Run() Tester {

	//time.Sleep(10 * time.Second)
	t.pass = false
	t.err = errors.New("not implemented")
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
