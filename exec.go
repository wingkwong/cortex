package main

import (
	"fmt"
	"os"

	"github.com/cortexlabs/cortex/pkg/lib/k8s"
)

func main() {
	k, err := k8s.New("default", false)
	if err != nil {
		exit(err)
	}

	// pod, err := k.GetPod("api-iris-classifier-95d88c4b7-478pl")
	// if err != nil {
	// 	exit(err)
	// }

	// out, err := k.Exec("api-iris-classifier-95d88c4b7-478pl", "api", []string{"/bin/sh", "-c", "echo hi"})
	// out, err := k.Exec("api-iris-classifier-95d88c4b7-478pl", "api", []string{"/bin/sh", "-c", "echo stderr >&2 && echo stdout && echo stderr >&2"})
	out, err := k.Exec("api-iris-classifier-95d88c4b7-478pl", "api", []string{"ss"})
	if err != nil {
		exit(err)
	}
	fmt.Println(out)
}

func exit(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}
