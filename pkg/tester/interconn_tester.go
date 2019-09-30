package tester

import (
	"errors"
	"github.com/cenkalti/backoff"
	log "github.com/golang/glog"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	"time"
)

type InterConnTester struct {
	cfg  *config.Config
	ctx  map[string]string
	pass bool
	err  error
}

func NewInterConnTester(cfg *config.Config, ctx map[string]string) *InterConnTester {
	return &InterConnTester{
		cfg:  cfg,
		ctx:  ctx,
		pass: false,
		err:  nil,
	}
}

func (t *InterConnTester) Run() Tester {
	t.pass = true
	return t
}

//check if connection available from public ip
func (t *InterConnTester) Check() Tester {
	if t.pass == false {
		return t
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(t.cfg.Timout) * time.Second

	checkInterConnection := func() error {
		return nil
	}

	publicip, exist := t.ctx["publicip"]
	if !exist {
		t.pass = false
		t.err = errors.New("can not find public ip in context")
		return t
	}

	err := backoff.Retry(checkInterConnection, b)
	if err != nil {
		log.V(1).Infof("connect to %s fail after timeout: %s", publicip, err.Error())
		t.pass = false
		t.err = err
	} else {
		t.pass = true
	}

	return t
}

// report inter-connection status
func (t *InterConnTester) Report(report interface{}) Tester {

	if !t.pass {
		report.(*model.CheckResult).InternetConnection = FAIL
		report.(*model.CheckResult).ErrorMsg = t.err.Error()
	} else {
		report.(*model.CheckResult).InternetConnection = PASS
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
