package selfcheck

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/tester"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"sync/atomic"
	log "github.com/golang/glog"

)

type SelfChecker struct {
	testcases []tester.Tester
	cfg       *config.Config
	locker    uint32
}

func NewSelfChecker(cfg *config.Config, kclient *kubernetes.Clientset) *SelfChecker {

	cases := []tester.Tester{
		tester.NewNamespaceTester(cfg, kclient),
		tester.NewPodTester(cfg, kclient),
		tester.NewSvcTester(cfg),
		tester.NewIntraConnTester(cfg),
		tester.NewInterConnTester(cfg),
	}

	return &SelfChecker{
		testcases: cases,
		cfg:       cfg,
	}
}

func (s *SelfChecker) Check(c *gin.Context) {

	if !atomic.CompareAndSwapUint32(&s.locker, 0, 1) {
		c.JSON(http.StatusTooManyRequests, model.CheckResult{
			ErrorMsg: "Another Self Check is running",
		})
		return
	}
	// deferred calls are executed in last-in-first-out
	defer atomic.StoreUint32(&s.locker, 0)
	defer s.shutdown()

	result := model.CheckResult{}
	for _, t := range s.testcases {
		if !t.Run().Check().Report(&result).Next() {
			c.JSON(http.StatusOK, result)
			return
		}
	}

	c.JSON(http.StatusOK, result)
}

func (s *SelfChecker) shutdown() {
	log.V(1).Info("Teardown all created resource in this test")
	for _, t := range s.testcases {
		t.Close()
	}
}
