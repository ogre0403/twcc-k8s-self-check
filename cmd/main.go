package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/config"
	"gitlab.com/twcc/twcc-k8s-self-check/pkg/k8sutil"
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
	kubeconfig = os.ExpandEnv(*flag.String("kubeconfig", "$HOME/.kube/config", "kubernetes configuration"))
	listenAddr = *flag.String("listen-addr", ":8080", "http server listen addr [addr:port]")
	user = *flag.String("user", WOODPECKER, "user name for httop basic auth")
	password = *flag.String("password", WOODPECKER, "user password for http basic auth")
	cfg.Namespace = *flag.String("namespace", WOODPECKER, "namespace name used for test")
	cfg.Pod = *flag.String("pod", WOODPECKER, "pod name used for test")
	cfg.Svc = *flag.String("svc", WOODPECKER, "service name used for test")
	cfg.Image = *flag.String("image", "nginx:latest", "container image used for test")
	cfg.Port = *flag.Int("port", 80, "container port used for test")
	cfg.Timout = *flag.Int("timeout", 30, "timeout for check test result")
	flag.Parse()
}

func showVersion() {
	log.V(1).Infof("BuildTime: %s", buildTime)
	log.V(1).Infof("CommitID: %s", commitID)
}

func main() {

	parserFlags()
	showVersion()

	kclient := k8sutil.GetK8SClientSet(kubeconfig)

	if kclient == nil {
		log.Fatal("Create kubernetes clientset fail")
		return
	}

	checker := selfcheck.NewSelfChecker(cfg, kclient)

	router := gin.Default()

	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		user: password,
	}))

	authorized.GET("/lifeCycleCheck", checker.Check)

	err := router.Run(listenAddr)
	if err != nil {
		log.Fatal(err.Error())
	}

}
