package k8sutil

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	blendedset "github.com/inwinstack/blended/generated/clientset/versioned"
	"io"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"strings"
)

func getRestConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func GetK8SClientSet(kubeconfig string) *kubernetes.Clientset {

	config, err := getRestConfig(kubeconfig)
	if err != nil {
		log.Fatalf("create kubenetes config fail: %s", err.Error())
		return nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("create kubenetes client set fail: %s", err.Error())
		return nil
	}

	return clientset
}

func GetInwinClientSet(kubeconfig string) *blendedset.Clientset {
	config, err := getRestConfig(kubeconfig)
	if err != nil {
		log.Fatalf("create kubenetes config fail: %s", err.Error())
		return nil
	}

	clientset, err := blendedset.NewForConfig(config)
	if err != nil {
		log.Fatalf("create inwin CRD kubenetes client set fail: %s", err.Error())
		return nil
	}

	return clientset
}

// ExecToPodThroughAPI uninterractively exec to the pod with the command specified.
// :param string command: list of the str which specify the command.
// :param string pod_name: Pod name
// :param string namespace: namespace of the Pod.
// :param io.Reader stdin: Standerd Input if necessary, otherwise `nil`
// :return: string: Output of the command. (STDOUT)
//          string: Errors. (STDERR)
//           error: If any error has occurred otherwise `nil`
func ExecToPodThroughAPI(kubeconfig, command, containerName, podName, namespace string, stdin io.Reader) (string, string, error) {
	config, err := getRestConfig(kubeconfig)
	if err != nil {
		return "", "", err
	}

	clientset := GetK8SClientSet(kubeconfig)
	if clientset == nil {
		return "", "", errors.New("Init clientset fail")
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	scheme := runtime.NewScheme()
	if err := core_v1.AddToScheme(scheme); err != nil {
		return "", "", fmt.Errorf("error adding to scheme: %v", err)
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&core_v1.PodExecOptions{
		Command:   strings.Fields(command),
		Container: containerName,
		Stdin:     stdin != nil,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("error while creating Executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return "", "", fmt.Errorf("error in Stream: %v", err)
	}

	return stdout.String(), stderr.String(), nil
}
