package main

import (
	"atlan/idp/pkg/jobmanager"
	"atlan/idp/pkg/server"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	namespace = "test"
)

func main() {
	kConfig, err := rest.InClusterConfig()
	if err != nil {
		//trying for local
		home, exists := os.LookupEnv("HOME")
		if !exists {
			home = "/root"
		}
		configPath := filepath.Join(home, ".kube", "config")
		kConfig, err = clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			log.Fatalln("failed to create K8s config", err.Error())
		}
	}
	jm := jobmanager.NewJobManager(kConfig, namespace)

	idpServer := server.New(jm)
	idpServer.Start()
}
