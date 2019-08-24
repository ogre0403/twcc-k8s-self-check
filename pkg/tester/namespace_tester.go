package tester

import (
	log "github.com/golang/glog"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type NamespaceTester struct {
	nsClient v1.NamespaceInterface
	cfg      *config.Config
	pass     bool
	err      error
}

func NewNamespaceTester(cfg *config.Config, kclient *kubernetes.Clientset) *NamespaceTester {

	nsClient := kclient.CoreV1().Namespaces()

	return &NamespaceTester{
		cfg:      cfg,
		nsClient: nsClient,
		pass:     false,
		err:      nil,
	}
}

func (t *NamespaceTester) Run() Tester {

	ns := corev1.Namespace{
		ObjectMeta: v12.ObjectMeta{
			Name: t.cfg.Namespace,
		},
	}

	_, err := t.nsClient.Create(&ns)

	if err != nil {
		log.V(1).Infof("Create namespace %s fail: %s", t.cfg.Namespace, err.Error())
	}

	return t
}

func (t *NamespaceTester) Check() Tester {

	_, err := t.nsClient.Get(t.cfg.Namespace, v12.GetOptions{})
	if err != nil {
		t.pass = false
		t.err = err
	} else {
		t.pass = true
	}

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
	log.V(1).Infof("Delete namespace %s", t.cfg.Namespace)
	err := t.nsClient.Delete(t.cfg.Namespace, &v12.DeleteOptions{})
	if err != nil {
		log.V(1).Infof("Delete namespace %s fail: %s", t.cfg.Namespace, err.Error())
	}
}

func (t *NamespaceTester) String() string {
	return "NamespaceTester"
}
