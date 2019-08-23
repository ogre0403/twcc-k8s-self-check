package tester


import (
	"errors"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

type SvcTester struct {
	pass bool
	err  error
}

func (t *SvcTester) Run() Tester {

	//time.Sleep(10 * time.Second)
	t.pass = false
	t.err = errors.New("not implemented")
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

