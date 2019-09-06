package selfcheck

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	blendedset "github.com/inwinstack/blended/generated/clientset/versioned"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/tester"
	"k8s.io/client-go/kubernetes"
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
	BasicTestCase = "BasicTest"
	ShmTestCase   = "ShmTest"
	GpuTestCase   = "GpuTest"
)

func NewSelfChecker(cfg *config.Config, kclient *kubernetes.Clientset, crdClient *blendedset.Clientset) *SelfChecker {

	ctx := make(map[string]string)

	basicTestCase := []tester.Tester{
		tester.NewNamespaceTester(cfg, kclient, ctx),
		tester.NewPodTester(cfg, kclient, ctx),
		tester.NewSvcTester(cfg, kclient, crdClient, ctx),
		tester.NewIntraConnTester(cfg, ctx),
		tester.NewInterConnTester(cfg, ctx),
	}

	shmTestCase := []tester.Tester{
		tester.NewShmPodTester(cfg, kclient, ctx),
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
}

func (s *SelfChecker) shutdown() {
	log.V(1).Infof("Teardown all created resource in test %s", s.testingCase)
	for _, t := range s.testCases[s.testingCase].Step {
		t.Close()
	}
	s.testingCase = ""
}
