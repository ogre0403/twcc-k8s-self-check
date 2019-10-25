package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/selfcheck"
	"os"
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
	flag.StringVar(&kubeconfig, "kubeconfig", "$HOME/.kube/config", "kubernetes configuration")
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "http server listen addr [addr:port]")
	flag.StringVar(&user, "user", WOODPECKER, "user name for httop basic auth")
	flag.StringVar(&password, "password", WOODPECKER, "user password for http basic auth")
	flag.StringVar(&cfg.Namespace, "namespace", WOODPECKER, "namespace name used for test")
	flag.StringVar(&cfg.Pod, "pod", WOODPECKER, "pod name used for test")
	flag.StringVar(&cfg.Svc, "svc", WOODPECKER, "service name used for test")
	flag.StringVar(&cfg.Image, "image", "nginx:latest", "container image used for test")
	flag.IntVar(&cfg.Port, "port", 80, "container application port used for test")
	flag.IntVar(&cfg.ExternalPort, "externalPort", 12345, "access port for external IP used for test")
	flag.IntVar(&cfg.Timout, "timeout", 30, "timeout for check test result")
	flag.Parse()
}

func showVersion() {
	log.V(1).Infof("BuildTime: %s", buildTime)
	log.V(1).Infof("CommitID: %s", commitID)
}

func main() {

	parserFlags()
	showVersion()

	checker := selfcheck.NewSelfChecker(cfg, os.ExpandEnv(kubeconfig))

	if checker == nil {
		log.Fatal("Init Checker fail, Exit...")
		return
	}

	router := gin.Default()

	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		user: password,
	}))

	authorized.GET("/lifeCycleCheck", checker.BasicCheck)
	authorized.POST("/shmCheck", checker.ShmCheck)
	authorized.POST("/gpuCheck", checker.GpuCheck)
	authorized.GET("/nodeGpuStatus", checker.NodeGpuStatus)

	err := router.Run(listenAddr)
	if err != nil {
		log.Fatal(err.Error())
	}

}
