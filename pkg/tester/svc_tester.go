package tester

import (
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	log "github.com/golang/glog"
	v13 "github.com/inwinstack/blended/apis/inwinstack/v1"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"

	blendedset "github.com/inwinstack/blended/generated/clientset/versioned"
	blendedsetv1 "github.com/inwinstack/blended/generated/clientset/versioned/typed/inwinstack/v1"
)

type SvcTester struct {
	cfg       *config.Config
	svcClient v1.ServiceInterface
	natClient blendedsetv1.NATInterface
	pass      bool
	err       error
	ctx       map[string]string
}

func NewSvcTester(cfg *config.Config, kclient *kubernetes.Clientset, crdClient *blendedset.Clientset, ctx map[string]string) *SvcTester {
	svcClient := kclient.CoreV1().Services(cfg.Namespace)
	natClient := crdClient.InwinstackV1().NATs(cfg.Namespace)

	return &SvcTester{
		cfg:       cfg,
		ctx:       ctx,
		svcClient: svcClient,
		natClient: natClient,
		pass:      false,
		err:       nil,
	}
}

func (t *SvcTester) Run() Tester {

	selector := map[string]string{
		"app": t.cfg.Pod,
	}

	svc := corev1.Service{
		ObjectMeta: v12.ObjectMeta{
			Namespace: t.cfg.Namespace,
			Name:      t.cfg.Svc,
		},
		Spec: corev1.ServiceSpec{
			ExternalIPs: []string{
				t.ctx["externalip"],
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "nginx",
					Port:       int32(t.cfg.ExternalPort),
					TargetPort: intstr.FromInt(t.cfg.Port),
				},
			},
			Selector: selector,
		},
	}

	_, err := t.svcClient.Create(&svc)

	if err != nil {
		log.V(1).Infof("Create svc %s fail: %s", t.cfg.Svc, err.Error())
		t.pass = false
		t.err = err
	} else {
		t.pass = true
	}

	return t
}

func (t *SvcTester) Check() Tester {
	if t.pass == false {
		return t
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(t.cfg.Timout) * time.Second

	checkNATActive := func() error {
		nats, err := t.natClient.List(v12.ListOptions{})
		if err != nil {
			return err
		}

		for _, nat := range nats.Items {
			if nat.Status.Phase == v13.NATActive {
				t.ctx["publicip"] = nat.Spec.DestinationAddresses[0]
				log.V(1).Infof("NAT %s has pulbic ip %s", nat.Name, nat.Spec.DestinationAddresses[0])
				return nil
			}
		}

		return errors.New(fmt.Sprintf("All NATs in namespace %s are not Active", t.cfg.Namespace))
	}

	err := backoff.Retry(checkNATActive, b)
	if err != nil {
		log.V(1).Infof("There is no active NATs after timeout: %s", err.Error())
		t.pass = false
		t.err = err
	} else {
		t.pass = true
	}
	return t
}

func (t *SvcTester) Report(report interface{}) Tester {

	if !t.pass {
		report.(*model.CheckResult).SvcCreate = FAIL
		report.(*model.CheckResult).ErrorMsg = t.err.Error()
	} else {
		report.(*model.CheckResult).SvcCreate = PASS
	}
	return t
}

func (t *SvcTester) Next() bool {
	return t.pass
}

func (t *SvcTester) Close() {}

func (t *SvcTester) String() string {
	return "SvcTester"
}
