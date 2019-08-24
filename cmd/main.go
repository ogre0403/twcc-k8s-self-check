package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/selfcheck"
)

const WOODPECKER = "woodpecker"

var (
	commitID   = "%COMMITID%"
	buildTime  = "%BUILDID%"
	kubeconfig string
	listenAddr string
	user       string
	password   string
	cfg        = &config.Config{}
)

func parserFlags() {

	flag.Set("logtostderr", "true")
	kubeconfig = *flag.String("kubeconfig", "$HOME/.kube/config", "kubernetes configuration")
	listenAddr = *flag.String("listen-addr", ":8080", "http server listen addr [addr:port]")
	user = *flag.String("user", WOODPECKER, "http server listen addr [addr:port]")
	password = *flag.String("password", WOODPECKER, "http server listen addr [addr:port]")

	cfg.Namespace = *flag.String("namespace", WOODPECKER, "http server listen addr [addr:port]")
	cfg.Pod = *flag.String("pod", WOODPECKER, "http server listen addr [addr:port]")
	cfg.Svc = *flag.String("svc", WOODPECKER, "http server listen addr [addr:port]")
	cfg.Image = *flag.String("image", "registry.twcc.ai/ngc/nvidia/tensorflow-18.12-py3-v1:latest", "http server listen addr [addr:port]")
	flag.Parse()
}

func showVersion() {
	log.V(1).Infof("BuildTime: %s", buildTime)
	log.V(1).Infof("CommitID: %s", commitID)
}

func main() {

	parserFlags()
	showVersion()

	checker := selfcheck.NewSelfChecker(cfg)

	router := gin.Default()

	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		"woodpecker": "woodpecker",
	}))

	authorized.GET("/selfcheck", checker.Check)

	err := router.Run(listenAddr)
	if err != nil {
		log.Fatal(err.Error())
	}

}
