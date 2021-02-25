package kube

import (
	"context"
	"fmt"
	"gangway/src/settings"
	"io"
	"log"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

var (
	gangwayPod *corev1.Pod
)

func init() {
	var err error
	gangwayPod, err = getGangwayPod()
	if err != nil {
		panic(err)
	}
}

func NewPipe() (*Pipe, error) {
	// get long live stream to gangway agent
	req := kc.Clientset.RESTClient().
		Post().
		Namespace(*settings.Namespace).
		Resource("pods").
		Name(gangwayPod.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: []string{"repeater"},
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
			TTY:     true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(kc.Config, "POST", req.URL())
	if err != nil {
		return nil, err
	}

	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()
	go func() {
		defer inReader.Close()
		defer outWriter.Close()
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  inReader,
			Stdout: outWriter,
			Stderr: os.Stderr,
			Tty:    true,
		})
		if err != nil {
			log.Printf("Connection Broken: %v\n", err)
		}
	}()

	return &Pipe{out: outReader, in: inWriter}, nil
}

func getGangwayPod() (*corev1.Pod, error) {
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

type Pipe struct {
	out *io.PipeReader
	in  *io.PipeWriter
}

func (p Pipe) Read(b []byte) (n int, err error) {
	return p.out.Read(b)
}
func (p Pipe) Write(b []byte) (n int, err error) {
	return p.in.Write(b)
}
func (p Pipe) Close() error {
	err := p.out.Close()
	if err != nil {
		return err
	}
	return p.in.Close()
}

// TODO
func (p Pipe) Alive() error {
	return nil
}
