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
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GpuPodTester struct {
	PodTester
}

const (
	GPUCOUNT    = "gpucount"
	CHECKGPUCMD = "nvidia-smi -L"
)

func NewGPUPodTester(cfg *config.Config, kclient *kubernetes.Clientset, ctx map[string]string) *GpuPodTester {
	podClient := kclient.CoreV1().Pods(v12.NamespaceDefault)

	return &GpuPodTester{
		PodTester: PodTester{
			podClient: podClient,
			ctx:       ctx,
			cfg:       cfg,
			pass:      false,
			err:       nil,
		},
	}
}

func (t *GpuPodTester) Run(req interface{}) Tester {

	request := req.(*model.Request)
	gpu := request.Gpu

	gpuInt, err := strconv.Atoi(gpu)

	if err != nil {
		t.pass = false
		t.err = err
		return t
	}

	if gpuInt < 1 {
		t.pass = false
		t.err = errors.New(fmt.Sprintf("GPU count must be larger than 0 in %s", t.String()))
		return t
	}

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

	t.ctx[GPUCOUNT] = gpu

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
					Image:   "nvidia/cuda:9.0-base",
					Command: []string{"sh", "-c", "while true; do sleep 1; done"},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"nvidia.com/gpu": resource.MustParse(gpu),
						},
						Requests: corev1.ResourceList{
							"nvidia.com/gpu": resource.MustParse(gpu),
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

func (t *GpuPodTester) Check() Tester {
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

	stdout, _, err := k8sutil.ExecToPodThroughAPI(t.ctx[KUBECONFIGPATH], CHECKGPUCMD, t.cfg.Pod, t.cfg.Pod, v12.NamespaceDefault, nil)
	if err != nil {
		t.pass = false
		t.err = errors.New(fmt.Sprintf("Run Shell command inside pod %s fail: %s", t.cfg.Pod, err.Error()))
		return t
	}

	if !checkPodGPUValue(stdout, t.ctx[GPUCOUNT]) {
		t.pass = false
		t.err = errors.New("GPU resource is not enforce in Pod")
		return t
	}

	t.pass = true
	return t
}

func (t *GpuPodTester) Report(report interface{}) Tester {
	if !t.pass {
		report.(*model.CheckResult).PodCreate = FAIL
		report.(*model.CheckResult).ErrorMsg = t.err.Error()
	} else {
		report.(*model.CheckResult).PodCreate = PASS
	}
	return t
}

func (t *GpuPodTester) Next() bool {
	return t.pass
}

func (t *GpuPodTester) Close() {

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

func (t *GpuPodTester) String() string {
	return "GPUPodTester"
}

func checkPodGPUValue(df_stdout, gpu string) bool {
	oo := regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(df_stdout), "\n")
	a := strings.Split(oo, "\n")

	return strconv.Itoa(len(a)) == gpu
}
