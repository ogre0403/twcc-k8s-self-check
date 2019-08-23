package tester

import (
	"errors"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

type NamespaceTester struct {
	pass bool
	err  error
}

func (t *NamespaceTester) Run() Tester {

	//time.Sleep(10 * time.Second)
	t.pass = false
	t.err = errors.New("not implemented")
	return t
}

func (t *NamespaceTester) Report(report *model.CheckResult) Tester {

	if !t.pass {
		report.NamespaceCreate = FAIL
		report.ErrorMsg = t.err.Error()
	} else {
		report.NamespaceCreate = PASS
	}
	return t
}

func (t *NamespaceTester) Next() bool {
	return t.pass
}
