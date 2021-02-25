package kube

import (
	"gangway/src/settings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kc *kubeClient
)

type kubeClient struct {
	Config    *rest.Config
	Clientset *kubernetes.Clientset
}

func Init() {
	var err error
	kc, err = newClient()
	if err != nil {
		panic(err)
	}

	gangwayPod, err = getGangwayPod()
	if err != nil {
		panic(err)
	}
}

func newClient() (*kubeClient, error) {
	var config *rest.Config
	var err error
	if settings.ContextType == settings.ContextTypeClient {
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *settings.Kubeconfig)
		if err != nil {
			return nil, err
		}
	} else if settings.ContextType == settings.ContextTypeController {
		config, err = rest.InClusterConfig()
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &kubeClient{
		Config:    config,
		Clientset: clientset,
	}, nil
}
