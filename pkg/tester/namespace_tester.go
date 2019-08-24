package tester

import (
	"errors"
	"fmt"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
)

type NamespaceTester struct {
	pass bool
	err  error
}

func NewNamespaceTester(cfg *config.Config) *NamespaceTester {
	return &NamespaceTester{
		pass: false,
		err:  errors.New("not implemented"),
	}
}

func (t *NamespaceTester) Run() Tester {

	//time.Sleep(10 * time.Second)

	return t
}

func (t *NamespaceTester) Check() Tester {
	t.pass = true
	return t
}

func (t *NamespaceTester) Report(report interface{}) Tester {

	if !t.pass {
		report.(*model.CheckResult).NamespaceCreate = FAIL
		report.(*model.CheckResult).ErrorMsg = t.err.Error()
	} else {
		report.(*model.CheckResult).NamespaceCreate = PASS
	}
	return t
}

func (t *NamespaceTester) Next() bool {
	return t.pass
}

func (t *NamespaceTester) Close() {
	// todo
	fmt.Printf("%s close\n", t)

}

func (t *NamespaceTester) String() string {
	return "NamespaceTester"
}
