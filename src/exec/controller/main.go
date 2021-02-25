package main

import (
	"gangway/src/proxy"
)

func main() {
	proxy.Serve(proxy.TypeController, "0.0.0.0", 7925)
}
