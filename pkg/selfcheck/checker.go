package selfcheck

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/tester"
	"net/http"
	"sync/atomic"
)

type SelfChecker struct {
	testcases []tester.Tester
	cfg       *config.Config
	locker    uint32
}

func NewSelfChecker(cfg *config.Config) *SelfChecker {

	cases := []tester.Tester{
		&tester.NamespaceTester{},
		&tester.PodTester{},
		&tester.SvcTester{},
		&tester.IntraConnTester{},
		&tester.InterConnTester{},
		&tester.TeardownTester{},
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
	defer atomic.StoreUint32(&s.locker, 0)

	result := model.CheckResult{}
	for _, t := range s.testcases {
		if !t.Run().Report(&result).Next() {
			c.JSON(http.StatusOK, result)
			return
		}
	}

	c.JSON(http.StatusOK, result)
}
