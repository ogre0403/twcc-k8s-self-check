package tester

import (
	log "github.com/golang/glog"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"regexp"
)

type NodeGpuUsageTester struct {
	podClient v1.PodInterface
	cfg       *config.Config
	pass      bool
	err       error
	ctx       map[string]int64
}

func NewNodeGPUUsageTester(cfg *config.Config, kclient *kubernetes.Clientset, ctx map[string]int64) *NodeGpuUsageTester {

	podClient := kclient.CoreV1().Pods(metav1.NamespaceAll)

	return &NodeGpuUsageTester{
		cfg:       cfg,
		podClient: podClient,
		ctx:       ctx,
		pass:      false,
		err:       nil,
	}
}

func (t *NodeGpuUsageTester) Run(req interface{}) Tester {
	t.pass = true
	return t
}

func (t *NodeGpuUsageTester) Check() Tester {
	if t.pass == false {
		return t
	}

	podList, err := t.podClient.List(v12.ListOptions{})

	// calculate gpu on all pod
	for _, pod := range podList.Items {
		count := parsePod(pod)

		if count == -1 {
			continue
		}

		//t.ctx[pod.Spec.NodeName] = count

		if c, exist := t.ctx[pod.Spec.NodeName]; exist {
			t.ctx[pod.Spec.NodeName] = count + c
		} else {
			t.ctx[pod.Spec.NodeName] = count
		}

	}

	if err != nil {
		log.V(1).Infof("List pod fail: %s", err.Error())
		t.pass = false
		t.err = err
	} else {
		t.pass = true
	}

	return t

}

func (t *NodeGpuUsageTester) Report(report interface{}) Tester {

	result := []model.NodeGPUUsage{}

	for k, v := range t.ctx {
		result = append(result, model.NodeGPUUsage{
			Node:  k,
			Count: v,
		})
	}

	if !t.pass {
		report.(*model.NodeGPUUsageResult).ErrorMsg = t.err.Error()
	} else {
		report.(*model.NodeGPUUsageResult).Status = result

	}
	return t

}

func (t *NodeGpuUsageTester) Next() bool {
	return t.pass
}

func (t *NodeGpuUsageTester) Close() {
	//t.ctx = make(map[string]int64)
	for k := range t.ctx {
		delete(t.ctx, k)
	}

}

func (t *NodeGpuUsageTester) String() string {
	return "NodeGpuUsageTester"
}

func parsePod(pod corev1.Pod) int64 {

	if pod.Status.Phase != corev1.PodRunning {
		return -1
	}

	matrchENT, _ := regexp.Compile("ent[0-9]{6}")
	matrchMST, _ := regexp.Compile("mst[0-9]{6}")
	matrchGOV, _ := regexp.Compile("gov[0-9]{6}")
	matrchACD, _ := regexp.Compile("acd[0-9]{6}")

	matchers := []*regexp.Regexp{matrchENT, matrchMST, matrchGOV, matrchACD}

	for _, reg := range matchers {
		s := reg.FindString(pod.Namespace)

		if s == "" {
			continue
		}

		quantity := pod.Spec.Containers[0].Resources.Requests["nvidia.com/gpu"]

		return quantity.Value()

	}
	return -1
}
