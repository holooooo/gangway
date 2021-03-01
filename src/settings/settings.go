package settings

import (
	"flag"
	"path/filepath"

	"github.com/rs/zerolog"
	"k8s.io/client-go/util/homedir"
)

var (
	debug       *bool
	IP          *string
	Port        *int
	PoolSize    *int
	PoolMaxIdle *int

	Kubeconfig *string
	Namespace  *string
	Name       *string

	EnableDNSPorxy *bool
	EnableRouter   *bool
	CIDR           *string
)

func init() {
	getConfig()
	getFeature()
	getKubeconfig()
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func getConfig() {
	debug = flag.Bool("debug", false, "sets log level to debug")
	IP = flag.String("listen-ip", "0.0.0.0", "(optional) Local ip to be listened")
	Port = flag.Int("listen-port", 9527, "(optional) Local port to be listened")

	PoolSize = flag.Int("pool-size", 10, "(optional) Maximum number of connections at the same time")
	PoolMaxIdle = flag.Int("pool-max-idle", 2, "(optional) Maximum number of idle connections at the same time")
}

func getKubeconfig() {
	if home := homedir.HomeDir(); home != "" {
		Kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) Absolute path to the kubeconfig file")
	} else {
		Kubeconfig = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file")
	}
	Namespace = flag.String("gangway-namespace", "default", "(optional) The namespace of gangway deployment")
	Name = flag.String("gangway-deploy", "gangway", "(optional) Gangway agent deployment name")
}

func getFeature() {
	EnableDNSPorxy = flag.Bool("enabld-dns-porxy", true, "[WARNING!!!]This will change your DNS settings. Only resolve domain names whose names end with '.svc'")
	EnableRouter = flag.Bool("enabld-cidr-proxy", true, "[WARNING!!!]This will modify your route. Only proxy requests that belong to cluster CIDR scope")
	CIDR = flag.String("cidr", "10.0.0.0/32", "target addr in cidr will be proxy to cluster")
}
