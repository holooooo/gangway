package kube

import (
	"context"
	"gangway/src/settings"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

var (
	informer informers.SharedInformerFactory
	stop     chan struct{} = make(chan struct{})

	DNS        string
	gangwayPod *corev1.Pod
)

func getClusterInfo() {
	defer close(stop)
	log.Info().Msg("start get info from cluster")

	if *settings.EnableDNSPorxy {
		DNS = getDns()
	}

	informer = informers.NewSharedInformerFactory(kc.Clientset, 10*time.Second)
	listenGangwayPod()
}

func listenGangwayPod() {
	log.Info().Msgf("looking for Gangway Controller pod in %v:%v", *settings.Namespace, *settings.Name)
	pods, err := kc.Clientset.CoreV1().Pods(*settings.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Warn().Err(err).Msg("init gangway controller pod name failed")
	}
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, *settings.Name) && pod.Status.Phase == corev1.PodRunning {
			gangwayPod = &pod
			break
		}
	}
	if gangwayPod == nil {
		log.Warn().Msg("no gangway pod has been find")
	}
	log.Info().Msgf("find gangway pod %v", gangwayPod.Name)

	podInformer := informer.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: updatePod,
	})
	podInformer.Run(stop)
}

func updatePod(old interface{}, cur interface{}) {
	curPod := cur.(*v1.Pod)
	oldPod := old.(*v1.Pod)
	if curPod.ResourceVersion == oldPod.ResourceVersion {
		return
	}

	isTargetPod := curPod.Namespace == *settings.Namespace || strings.HasPrefix(curPod.Name, *settings.Name)
	if isTargetPod && curPod.Status.Phase == corev1.PodRunning && curPod.Name != gangwayPod.Name {
		log.Info().Msgf("update remote controller pod to %v", curPod.Name)
		gangwayPod = curPod
	}
}

func getDns() string {
	log.Info().Msg("looking for cluster dns")
	svc, err := kc.Clientset.CoreV1().Services("kube-system").Get(context.TODO(), "kube-dns", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	if svc != nil {
		return kc.Config.Host + "/api/v1/namespaces/kube-system/services/kube-dns/proxy"
	}
	return ""
}
