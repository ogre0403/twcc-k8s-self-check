package tester

import (
	"errors"
	"github.com/cenkalti/backoff"
	log "github.com/golang/glog"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
)

type PodTester struct {
	podClient v1.PodInterface
	cfg       *config.Config
	pass      bool
	err       error
	ctx       map[string]string
}

func NewPodTester(cfg *config.Config, kclient *kubernetes.Clientset, ctx map[string]string) *PodTester {

	podClient := kclient.CoreV1().Pods(cfg.Namespace)

	return &PodTester{
		podClient: podClient,
		ctx:       ctx,
		cfg:       cfg,
		pass:      false,
		err:       nil,
	}
}

func (t *PodTester) Run() Tester {

	lbl := map[string]string{
		"app": t.cfg.Pod,
	}

	pod := corev1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Namespace: t.cfg.Namespace,
			Name:      t.cfg.Pod,
			Labels:    lbl,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  t.cfg.Pod,
					Image: t.cfg.Image,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 80,
						},
					},
				},
			},
		},
	}

	_, err := t.podClient.Create(&pod)

	if err != nil {
		log.V(1).Infof("Create pod %s fail: %s", t.cfg.Pod, err.Error())
		t.pass = false
		t.err = err
	} else {
		t.pass = true
	}

	return t
}

func (t *PodTester) Check() Tester {
	if t.pass == false {
		return t
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(t.cfg.Timout) * time.Second

	checkPodRunning := func() error {
		pod, err := t.podClient.Get(t.cfg.Pod, v12.GetOptions{})
		if err != nil {
			return err
		}

		if pod.Status.Phase != corev1.PodRunning {
			return errors.New("Pod State is not Running")
		}
		return nil
	}

	err := backoff.Retry(checkPodRunning, b)
	if err != nil {
		log.V(1).Infof("pod %s is not running after timeout: %s", t.cfg.Pod, err.Error())
		t.pass = false
		t.err = err
	} else {
		t.pass = true
	}

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
