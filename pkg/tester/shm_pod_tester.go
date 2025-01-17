package tester

import (
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	log "github.com/golang/glog"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/k8sutil"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
	"strings"
	"time"
)

type ShmPodTester struct {
	PodTester
}

const (
	SHMLIMIT       = "shmlimit"
	KUBECONFIGPATH = "kubeconfigpath"
)

func NewShmPodTester(cfg *config.Config, kclient *kubernetes.Clientset, ctx map[string]string) *ShmPodTester {
	podClient := kclient.CoreV1().Pods(v12.NamespaceDefault)

	return &ShmPodTester{
		PodTester: PodTester{
			podClient: podClient,
			ctx:       ctx,
			cfg:       cfg,
			pass:      false,
			err:       nil,
		},
	}
}

func (t *ShmPodTester) Run(req interface{}) Tester {

	request := req.(*model.Request)
	shm := request.ShmLimit

	// setup node selector affinity
	var affinity *corev1.Affinity = nil
	if request.Node != "" {
		log.V(1).Infof("Node Selector is defined, select node %s", request.Node)
		affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "kubernetes.io/hostname",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{request.Node},
								},
							},
						},
					},
				},
			},
		}
	}

	quantity, err := resource.ParseQuantity(shm)

	t.ctx[SHMLIMIT] = shm

	if err != nil {
		log.V(1).Infof("Parse quantity %s fail: %s", shm, err.Error())
		t.pass = false
		t.err = err
		return t
	}

	pod := corev1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Namespace: v12.NamespaceDefault,
			Name:      t.cfg.Pod,
		},
		Spec: corev1.PodSpec{
			Affinity: affinity,
			Containers: []corev1.Container{
				{
					Name:    t.cfg.Pod,
					Image:   "busybox",
					Command: []string{"sh", "-c", "while true; do sleep 1; done"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "dshm",
							MountPath: "/dev/shm",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "dshm",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium:    corev1.StorageMediumMemory,
							SizeLimit: &quantity,
						},
					},
				},
			},
		},
	}

	_, err = t.podClient.Create(&pod)

	if err != nil {
		log.V(1).Infof("Create pod %s fail: %s", t.cfg.Pod, err.Error())
		t.pass = false
		t.err = err
		return t
	}

	t.pass = true
	return t
}

func (t *ShmPodTester) Check() Tester {
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
		return t
	}

	stdout, _, err := k8sutil.ExecToPodThroughAPI(t.ctx[KUBECONFIGPATH], "df -B 1 /dev/shm", t.cfg.Pod, t.cfg.Pod, v12.NamespaceDefault, nil)
	if err != nil {
		t.pass = false
		t.err = errors.New(fmt.Sprintf("Run Shell command inside pod %s fail: %s", t.cfg.Pod, err.Error()))
		return t
	}

	if !checkPodShmValue(stdout, t.ctx[SHMLIMIT]) {
		t.pass = false
		t.err = errors.New("SHM limit is not enforce in Pod")
		return t
	}

	t.pass = true
	return t
}

func (t *ShmPodTester) Report(report interface{}) Tester {
	if !t.pass {
		report.(*model.CheckResult).PodCreate = FAIL
		report.(*model.CheckResult).ErrorMsg = t.err.Error()
	} else {
		report.(*model.CheckResult).PodCreate = PASS
	}
	return t
}

func (t *ShmPodTester) Next() bool {
	return t.pass
}

func (t *ShmPodTester) Close() {

	log.V(1).Infof("Delete pod %s", t.cfg.Pod)
	err := t.podClient.Delete(t.cfg.Pod, &v12.DeleteOptions{})
	if err != nil {
		log.V(1).Infof("Delete Pod %s fail: %s", t.cfg.Pod, err.Error())
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 600 * time.Second

	checkPodExist := func() error {
		_, err := t.podClient.Get(t.cfg.Pod, v12.GetOptions{})

		if k8serr.IsNotFound(err) {
			return nil
		}

		return errors.New(fmt.Sprintf("Pod %s is still in Terminating", t.cfg.Pod))
	}

	err = backoff.Retry(checkPodExist, b)
	if err != nil {
		log.V(1).Infof("Pod %s is hang in Terminating state after timeout: %s", t.cfg.Pod, err.Error())
	}

}

func (t *ShmPodTester) String() string {
	return "ShmPodTester"
}

func checkPodShmValue(df_stdout, shm string) bool {
	q, _ := resource.ParseQuantity(shm)
	aa := strings.Split(df_stdout, "\n")
	bb := strings.Fields(aa[1])
	return strconv.Itoa(int(q.Value())) == bb[1]
}
