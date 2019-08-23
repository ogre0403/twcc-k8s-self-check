package tester

import (
	"errors"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

type PodTester struct {
	pass bool
	err  error
}

func (t *PodTester) Run() Tester {

	//time.Sleep(10 * time.Second)
	t.pass = false
	t.err = errors.New("not implemented")
	return t
}

func (t *PodTester) Report(report *model.CheckResult) Tester {

	if !t.pass {
		report.PodCreate = FAIL
		report.ErrorMsg = t.err.Error()
	} else {
		report.PodCreate = PASS
	}
	return t
}

func (t *PodTester) Next() bool {
	return t.pass
}
