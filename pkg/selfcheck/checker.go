package selfcheck

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/model"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/tester"
	"net/http"
)

const (
	PASS = "PASS"
	FAIL = "FAIL"
)

type SelfChecker struct {
	testcases []tester.Tester
	lock      bool
	cfg       *config.Config
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
		lock:      false,
		cfg:       cfg,
	}
}

func (s *SelfChecker) Check(c *gin.Context) {

	result := model.CheckResult{}

	if s.lock {
		result.ErrorMsg = "Another Self Check is running"
		c.JSON(http.StatusTooManyRequests, result)
		return
	}

	s.lock = true

	for _, t := range s.testcases {
		if !t.Run().Report(&result).Next() {
			s.lock = false
			c.JSON(http.StatusOK, result)
			return
		}
	}

	c.JSON(http.StatusOK, result)
	s.lock = false
}
