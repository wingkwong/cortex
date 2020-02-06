package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cortexlabs/cortex/pkg/lib/k8s"
)

func main() {
	k, err := k8s.New("default", false)
	if err != nil {
		exit(err)
	}

	for true {
		getNumConnections(k)
		time.Sleep(2 * time.Second)
	}
}

func getNumConnections(k *k8s.Client) {
	pods, err := k.ListPodsWithLabelKeys("apiName")
	if err != nil {
		exit(err)
	}

	if len(pods) == 0 {
		return
	}

	out, err := k.Exec(pods[0].Name, "api", []string{"ss"})
	if err != nil {
		exit(err)
	}

	// fmt.Println(out)
	fmt.Printf("NUM CONNECTIONS: %d\n", strings.Count(out, "tcp"))
}

func exit(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}
