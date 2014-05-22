package main

import (
	"os"

	worker "github.com/innotech/hydra-worker-map-sort/vendors/github.com/innotech/hydra-worker-lib"
)

func main() {
	if len(os.Args) < 3 {
		panic("Invalid number of arguments, you need to add at least the arguments for the server address and the service name")
	}
	serverAddr := os.Args[1]  // e.g. "tcp://localhost:5555"
	serviceName := os.Args[2] // e.g. map-sort
	verbose := len(os.Args) >= 4 && os.Args[3] == "-v"

	// New Map and Sort Worker connected to Hydra Load Balancer
	mapAndSortWorker := worker.NewWorker(serverAddr, serviceName, verbose)
	fn := func(instances []interface{}, args map[string]interface{}) []interface{} {
		var mappedInstances map[string][]interface{}
		mappedInstances = make(map[string][]interface{})
		for _, i := range instances {
			instance := i.(map[string]interface{})
			key := instance["Info"].(map[string]interface{})[args["mapAttr"].(string)].(string)
			if len(mappedInstances[key]) == 0 {
				mappedInstances[key] = make([]interface{}, 0)
			}
			if _, ok := mappedInstances[key]; ok {
				mappedInstances[key] = append(mappedInstances[key], instance)
			} else {
				mappedInstances[key] = []interface{}{instance}
			}
		}

		computedInstances := make([]interface{}, 0)
		for _, mapAttr := range args["mapSort"].(map[string]interface{}) {
			if _, ok := mappedInstances[mapAttr.(string)]; ok {
				computedInstances = append(computedInstances, mappedInstances[mapAttr.(string)])
			}
		}

		return computedInstances
	}
	mapAndSortWorker.Run(fn)
}
