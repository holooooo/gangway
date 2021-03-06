package kube

import (
	"errors"
	"gangway/src/settings"
	"io"
	"os"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

var (
	noGangwayPod = errors.New("no gangway pod find")
)

func NewPipe() (*Pipe, error) {
	if gangwayPod == nil {
		return nil, noGangwayPod
	}

	// get long live stream to gangway agent
	req := kc.Clientset.CoreV1().RESTClient().
		Post().
		Namespace(*settings.Namespace).
		Resource("pods").
		Name(gangwayPod.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: gangwayPod.Spec.Containers[0].Name,
			Command:   []string{"repeater"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(kc.Config, "POST", req.URL())
	if err != nil {
		return nil, err
	}

	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()
	pipe := &Pipe{
		pod:  *gangwayPod,
		out:  outReader,
		in:   inWriter,
		stop: make(chan struct{}),
	}
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
			log.Warn().Err(err).Msg("pipe error")
			pipe.Close()
		}
		<-pipe.stop
	}()

	log.Debug().Msg("new kube stream pipe is created")
	return pipe, nil
}

type Pipe struct {
	pod  corev1.Pod
	out  *io.PipeReader
	in   *io.PipeWriter
	stop chan struct{}
}

func (p Pipe) Read(b []byte) (n int, err error) {
	return p.out.Read(b)
}
func (p Pipe) Write(b []byte) (n int, err error) {
	return p.in.Write(b)
}
func (p Pipe) Close() error {
	close(p.stop)
	log.Debug().Msg("kube stream pipe is closing")
	return p.in.Close()
}

// TODO
func (p Pipe) Alive() error {
	if gangwayPod != &p.pod {
		return errors.New("pod is not exist")
	}
	return nil
}
