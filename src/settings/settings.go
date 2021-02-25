package settings

import (
	"flag"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

var (
	IP          *string
	Port        *int
	PoolSize    *int
	PoolMaxIdle *int

	Kubeconfig *string
	Namespace  *string
	Name       *string

	EnableDNSPorxy *bool
	EnableRouter   *bool
)

func init() {
	getConfig()
	getKubeconfig()
	getFeature()

	flag.Parse()
}

func getConfig() {
	IP = flag.String("ip", "0.0.0.0", "(optional) Local ip to be listened")
	Port = flag.Int("port", 9527, "(optional) Local port to be listened")

	PoolSize = flag.Int("PoolSize", 10, "(optional) Maximum number of connections at the same time")
	PoolMaxIdle = flag.Int("PoolMaxIdle", 2, "(optional) Maximum number of idle connections at the same time")
}

func getKubeconfig() {
	if home := homedir.HomeDir(); home != "" {
		Kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) Absolute path to the kubeconfig file")
	} else {
		Kubeconfig = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file")
	}
	Namespace = flag.String("namespace", "default", "(optional) The namespace of gangway deployment")
	Name = flag.String("name", "gangway", "(optional) Gangway agent deployment name")
}

func getFeature() {
	EnableDNSPorxy = flag.Bool("EnableDNSPorxy", true, "This will change your DNS settings. Only resolve domain names whose names end with '.svc'")
	EnableRouter = flag.Bool("EnableRouter", true, "This will modify your route. Only proxy requests that belong to cluster CIDR scope")
}
