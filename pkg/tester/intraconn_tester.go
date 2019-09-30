package tester

import (
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	log "github.com/golang/glog"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	"net"
	"time"
)

type IntraConnTester struct {
	cfg  *config.Config
	ctx  map[string]string
	pass bool
	err  error
}

func NewIntraConnTester(cfg *config.Config, ctx map[string]string) *IntraConnTester {
	return &IntraConnTester{
		ctx:  ctx,
		cfg:  cfg,
		pass: false,
		err:  nil,
	}
}

func (t *IntraConnTester) Run() Tester {
	t.pass = true
	return t
}

//check if connection available from external ip
func (t *IntraConnTester) Check() Tester {
	if t.pass == false {
		return t
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(t.cfg.Timout) * time.Second

	exteranlIP, exist := t.ctx["externalip"]
	if !exist {
		t.pass = false
		t.err = errors.New("can not find external ip in context")
		return t
	}

	checkIntraConnection := func() error {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", exteranlIP, t.cfg.ExternalPort))
		if err != nil {
			return err
		}
		defer conn.Close()
		return nil
	}

	err := backoff.Retry(checkIntraConnection, b)
	if err != nil {
		log.V(1).Infof("connect to %s fail after timeout: %s", exteranlIP, err.Error())
		t.pass = false
		t.err = err
	} else {
		t.pass = true
	}

	return t
}

// report intra-connection status
func (t *IntraConnTester) Report(report interface{}) Tester {

	if !t.pass {
		report.(*model.CheckResult).IntranetConnection = FAIL
		report.(*model.CheckResult).ErrorMsg = t.err.Error()
	} else {
		report.(*model.CheckResult).IntranetConnection = PASS
	}
	return t
}

func (t *IntraConnTester) Next() bool {
	return t.pass
}

func (t *IntraConnTester) Close() {}

func (t *IntraConnTester) String() string {
	return "IntraConnTester"
}
