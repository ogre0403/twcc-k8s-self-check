package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/selfcheck"
)

var (
	commitID  = "%COMMITID%"
	buildTime = "%BUILDID%"
)

func init() {
	flag.Set("logtostderr", "true")
}

func showVersion() {
	log.V(1).Infof("BuildTime: %s", buildTime)
	log.V(1).Infof("CommitID: %s", commitID)
}

func main() {

	flag.Parse()
	showVersion()

	checker := selfcheck.NewSelfChecker()

	router := gin.Default()

	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		"woodpecker": "woodpecker",
	}))

	authorized.GET("/selfcheck", checker.Check)

	router.Run()

}
