package tester

import (
	"errors"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

type PodTester struct {
	pass bool
	err  error
}

func NewPodTester(cfg *config.Config) *PodTester {
	return &PodTester{
		pass: false,
		err:  errors.New("not implemented"),
	}
}

func (t *PodTester) Run() Tester {

	return t
}

func (t *PodTester) Check() Tester {
	t.pass = true

	return t
}

func (t *PodTester) Report(report interface{}) Tester {

	if !t.pass {
		report.(*model.CheckResult).PodCreate = FAIL
		report.(*model.CheckResult).ErrorMsg = t.err.Error()
	} else {
		report.(*model.CheckResult).PodCreate = PASS
	}
	return t
}

func (t *PodTester) Next() bool {
	return t.pass
}

func (t *PodTester) Close() {}

func (t *PodTester) String() string {
	return "PodTester"
}
