package main

import (
	"flag"
	"svc-op-service/router"

	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Set("v", "4")
	flag.Parse()

	r := router.InitRouter()
	r.Run(":8080")
}
