package settings

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"k8s.io/client-go/util/homedir"
)

var (
	debug       = flag.Bool("debug", false, "sets log level to debug")
	IP          = flag.String("listen-ip", "0.0.0.0", "(optional) Local ip to be listened")
	Port        = flag.Int("listen-port", 9527, "(optional) Local port to be listened")
	PoolSize    = flag.Int("pool-size", 10, "(optional) Maximum number of connections at the same time")
	PoolMaxIdle = flag.Int("pool-max-idle", 2, "(optional) Maximum number of idle connections at the same time")

	Kubeconfig *string
	Namespace  = flag.String("gangway-namespace", "default", "(optional) The namespace of gangway deployment")
	Name       = flag.String("gangway-deploy", "gangway", "(optional) Gangway agent deployment name")

	EnableDNSPorxy = flag.Bool("enabled-dns-proxy", true, "forward dns question from local dns to cluster dns")
	DNSPort        = flag.Int("dns listener port", 8553, "(optional) local dns port")
	EnableRouter   = flag.Bool("enabled-cidr-proxy", true, "[WARNING!!!]This will modify your eth device.")
	CIDR           = flag.String("cidr", "10.0.0.0/8", "target addr in cidr will be proxy to cluster")
)

func init() {
	testing.Init()

	getKubeconfig()
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
func getKubeconfig() {
	if home := homedir.HomeDir(); home != "" {
		Kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) Absolute path to the kubeconfig file")
	} else {
		Kubeconfig = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file")
	}
}
