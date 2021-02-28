package kube

import (
	"context"
	"fmt"
	"gangway/src/settings"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kc         *kubeClient
	gangwayPod *corev1.Pod
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
	} else if settings.ContextType == settings.ContextTypeController {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
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

func getGangwayPod() (*corev1.Pod, error) {
	log.Info().Msgf("Looking for Gangway Controller pod in %v:%v", *settings.Namespace, *settings.Name)

	deploy, err := kc.Clientset.AppsV1().
		Deployments(*settings.Namespace).
		Get(context.TODO(), *settings.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	podLabels := labels.FormatLabels(deploy.Spec.Template.Labels)
	pods, err := kc.Clientset.CoreV1().
		Pods(*settings.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: podLabels})
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("Gangway agent not deploy in target cluster")
	}
	return &pods.Items[0], nil
}
