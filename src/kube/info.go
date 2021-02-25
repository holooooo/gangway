package kube

type clusterInfo struct {
	cidr []string
	dns  string
}

// todo
// kubectl get nodes -o jsonpath='{.items[*].spec.podCIDR}'
func listenCidr() []string {
	return nil
}

func getDns() string {
	return ""
}
