package main

import (
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"time"
)

func main() {
	found := false
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("Could not find in-cluster config")
	} else {
		found = true
	}

	if !found {
		homeDir, _ := os.UserHomeDir()
		restConfig, err = clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/%s", homeDir, ".kube/config"))
		if err != nil {
			fmt.Errorf("could not load config from file")
		}
	}

	fmt.Println("Starting Reconciler")
	r := Reconciler{restConfig}
	for {
		r.Run()
		time.Sleep(3 * time.Second)
	}
}
