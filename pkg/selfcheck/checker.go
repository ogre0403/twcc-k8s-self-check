package selfcheck

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/k8sutil"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/tester"
	"net/http"
	"sync/atomic"
)

type TestCase struct {
	Name string
	Step []tester.Tester
}

type SelfChecker struct {
	testCases   map[string]TestCase
	testingCase string
	locker      uint32
}

const (
	BasicTestCase  = "BasicTest"
	ShmTestCase    = "ShmTest"
	GpuTestCase    = "GpuTest"
	NodeGPUUsage   = "NodeGpuUsage"
	KUBECONFIGPATH = "kubeconfigpath"
)

func NewSelfChecker(cfg *config.Config, kubeconfig string) *SelfChecker {

	kclient := k8sutil.GetK8SClientSet(kubeconfig)
	crdClient := k8sutil.GetInwinClientSet(kubeconfig)

	if kclient == nil || crdClient == nil {
		log.Fatal("Create kubernetes clientset fail")
		return nil
	}

	ctx := make(map[string]string)

	// test case has five sequential steps to check
	basicTestCase := []tester.Tester{
		// check if namespace is created corrected
		tester.NewNamespaceTester(cfg, kclient, ctx),
		// check if pod is created corrected
		tester.NewPodTester(cfg, kclient, ctx),
		// check if service is created corrected
		tester.NewSvcTester(cfg, kclient, crdClient, ctx),
		// check
		tester.NewIntraConnTester(cfg, ctx),
		// check if connection available from public ip
		tester.NewInterConnTester(cfg, ctx),
	}

	ctx2 := make(map[string]string)
	ctx2[KUBECONFIGPATH] = kubeconfig
	shmTestCase := []tester.Tester{
		tester.NewShmPodTester(cfg, kclient, ctx2),
	}

	ctx3 := make(map[string]string)
	ctx3[KUBECONFIGPATH] = kubeconfig
	gpuTestCase := []tester.Tester{
		tester.NewGPUPodTester(cfg, kclient, ctx3),
	}

	ctx4 := make(map[string]int64)
	nodeGPUUsage := []tester.Tester{
		tester.NewNodeGPUUsageTester(cfg, kclient, ctx4),
	}

	testCase := map[string]TestCase{
		BasicTestCase: {
			Name: BasicTestCase,
			Step: basicTestCase,
		},
		ShmTestCase: {
			Name: ShmTestCase,
			Step: shmTestCase,
		},
		GpuTestCase: {
			Name: GpuTestCase,
			Step: gpuTestCase,
		},
		NodeGPUUsage: {
			Name: NodeGPUUsage,
			Step: nodeGPUUsage,
		},
	}

	return &SelfChecker{
		testingCase: "",
		testCases:   testCase,
	}
}

func (s *SelfChecker) BasicCheck(c *gin.Context) {
	// test TestCase1

	if !atomic.CompareAndSwapUint32(&s.locker, 0, 1) {
		c.JSON(http.StatusTooManyRequests, model.CheckResult{
			ErrorMsg: fmt.Sprintf("Another Check %s is running", s.testingCase),
		})
		return
	}
	s.testingCase = BasicTestCase

	// deferred calls are executed in last-in-first-out
	defer atomic.StoreUint32(&s.locker, 0)
	defer s.shutdown()

	result := model.CheckResult{}
	for _, t := range s.testCases[BasicTestCase].Step {
		if !t.Run(nil).Check().Report(&result).Next() {
			c.JSON(http.StatusOK, result)
			return
		}
	}

	c.JSON(http.StatusOK, result)
}

func (s *SelfChecker) ShmCheck(c *gin.Context) {
	// test TestCase2

	var req model.Request
	err := c.BindJSON(&req)
	if err != nil {
		log.Errorf("Failed to parse spec request request: %s", err.Error())
		return
	}

	if !atomic.CompareAndSwapUint32(&s.locker, 0, 1) {
		c.JSON(http.StatusTooManyRequests, model.CheckResult{
			ErrorMsg: fmt.Sprintf("Another Check %s is running", s.testingCase),
		})
		return
	}
	s.testingCase = ShmTestCase
	// deferred calls are executed in last-in-first-out
	defer atomic.StoreUint32(&s.locker, 0)
	defer s.shutdown()

	result := model.CheckResult{}
	for _, t := range s.testCases[ShmTestCase].Step {
		if !t.Run(&req).Check().Report(&result).Next() {
			c.JSON(http.StatusOK, result)
			return
		}
	}
	c.JSON(http.StatusOK, result)
}

func (s *SelfChecker) GpuCheck(c *gin.Context) {
	var req model.Request
	err := c.BindJSON(&req)
	if err != nil {
		log.Errorf("Failed to parse spec request request: %s", err.Error())
		return
	}

	if !atomic.CompareAndSwapUint32(&s.locker, 0, 1) {
		c.JSON(http.StatusTooManyRequests, model.CheckResult{
			ErrorMsg: fmt.Sprintf("Another Check %s is running", s.testingCase),
		})
		return
	}
	s.testingCase = GpuTestCase
	// deferred calls are executed in last-in-first-out
	defer atomic.StoreUint32(&s.locker, 0)
	defer s.shutdown()

	result := model.CheckResult{}
	for _, t := range s.testCases[GpuTestCase].Step {
		if !t.Run(&req).Check().Report(&result).Next() {
			c.JSON(http.StatusOK, result)
			return
		}
	}
	c.JSON(http.StatusOK, result)
}

func (s *SelfChecker) NodeGpuStatus(c *gin.Context) {
	if !atomic.CompareAndSwapUint32(&s.locker, 0, 1) {
		c.JSON(http.StatusTooManyRequests, model.CheckResult{
			ErrorMsg: fmt.Sprintf("Another Check %s is running", s.testingCase),
		})
		return
	}

	s.testingCase = NodeGPUUsage
	// deferred calls are executed in last-in-first-out
	defer atomic.StoreUint32(&s.locker, 0)
	defer s.shutdown()

	result := model.NodeGPUUsageResult{}
	for _, t := range s.testCases[NodeGPUUsage].Step {
		if !t.Run(nil).Check().Report(&result).Next() {
			c.JSON(http.StatusOK, result)
			return
		}
	}
	c.JSON(http.StatusOK, result)
}

func (s *SelfChecker) shutdown() {
	log.V(1).Infof("Teardown all created resource in test %s", s.testingCase)
	for _, t := range s.testCases[s.testingCase].Step {
		t.Close()
	}
	s.testingCase = ""
}
