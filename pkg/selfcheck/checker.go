package selfcheck

import (
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

type SelfChecker struct {
	testcases []tester.Tester
	locker    uint32
}

func NewSelfChecker(cfg *config.Config, kclient *kubernetes.Clientset, crdClient *blendedset.Clientset) *SelfChecker {

	ctx := make(map[string]string)
	cases := []tester.Tester{
		tester.NewNamespaceTester(cfg, kclient, ctx),
		tester.NewPodTester(cfg, kclient, ctx),
		tester.NewSvcTester(cfg, kclient, crdClient, ctx),
		tester.NewIntraConnTester(cfg, ctx),
		tester.NewInterConnTester(cfg, ctx),
	}

	return &SelfChecker{
		testcases: cases,
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
