package main

import (
	"encoding/json"
	// "log"
	"os"

	worker "github.com/innotech/hydra-worker-pong/vendors/github.com/innotech/hydra-worker-lib"
)

func main() {
	if len(os.Args) < 3 {
		panic("Invalid number of arguments, you need to add at least the arguments for the server address and the service name")
	}
	serverAddr := os.Args[1]  // e.g. "tcp://localhost:5555"
	serviceName := os.Args[2] // e.g. map-sort
	verbose := len(os.Args) >= 4 && os.Args[3] == "-v"

	// New Worker connected to Hydra Load Balancer
	mapAndSortWorker := worker.NewWorker(serverAddr, serviceName, verbose)
	fn := func(instances []map[string]interface{}, args map[string]string) []interface{} {
		mappedInstances = make(map[string]interface{})
		for _, instance := range instances {
			if val, ok := mappedInstances[instance[args["mapAttr"]]]; ok {
				mappedIstances[instance[args["mapAttr"]]] = append(mappedIstances[instance[args["mapAttr"]]], instance)
			} else {
				mappedIstances[instance[args["mapAttr"]]] = []interface{}{instance}
			}
		}

		var mapSort []string
		if err := json.Unmarshal(args["mapSort"], &mapSort); err != nil {
			// TODO: process error
		}

		computedInstances := make([]interface{}, 0)
		for _, mapAttr := range mapSort {
			computedInstances = append(computedInstances, mappedInstances[mapAttr])
		}

		return computedInstances
	}
	mapAndSortWorker.Run(fn)
}
