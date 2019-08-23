package selfcheck

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type CheckResult struct {
	NamespaceCreate    bool `json:"CreateNamespace"`
	PodCreate          bool `json:"CreatePod"`
	SvcCreate          bool `json:"CreateSVC"`
	IntranetConnection bool `json:"IntraConnection"`
	InternetConnection bool `json:"InterConnection"`
	Teardown           bool `json:"Teardown"`
}

type SelfChecker struct {
}

func NewSelfChecker() *SelfChecker {
	return nil
}

func (s *SelfChecker) Check(c *gin.Context) {

	c.JSON(http.StatusOK, CheckResult{
		NamespaceCreate:    false,
		PodCreate:          false,
		SvcCreate:          false,
		IntranetConnection: false,
		InternetConnection: false,
		Teardown:           false,
	})
}
