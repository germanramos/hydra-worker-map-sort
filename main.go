package main

import (
	"os"
	"strconv"

	worker "github.com/innotech/hydra-worker-map-sort/vendors/github.com/innotech/hydra-worker-lib"
)

const UNDEFINED_MAP string = "undefined"

func main() {
	// New Map and Sort Worker connected to Hydra Load Balancer
	mapAndSortWorker := worker.NewWorker(os.Args)
	fn := func(instances []interface{}, args map[string]interface{}) []interface{} {
		var mappedInstances map[string][]interface{}
		mappedInstances = make(map[string][]interface{})
		for _, i := range instances {
			instance := i.(map[string]interface{})
			var key string = UNDEFINED_MAP
			mapAttr, ok := args["mapAttr"]
			if ok {
				keyInterface, ok := instance["Info"].(map[string]interface{})[mapAttr.(string)]
				if ok {
					key = keyInterface.(string)
				}
			}
			if len(mappedInstances[key]) == 0 {
				mappedInstances[key] = make([]interface{}, 0)
			}
			if _, ok := mappedInstances[key]; ok {
				mappedInstances[key] = append(mappedInstances[key], instance)
			} else {
				mappedInstances[key] = []interface{}{instance}
			}
		}

		mapSortValues := args["mapSort"].(map[string]interface{})
		mapSort := make([]string, len(mapSortValues), len(mapSortValues))
		for i, value := range mapSortValues {
			index, _ := strconv.Atoi(i)
			mapSort[index] = value.(string)
		}

		computedInstances := make([]interface{}, 0)
		for _, mapAttr := range mapSort {
			if _, ok := mappedInstances[mapAttr]; ok {
				computedInstances = append(computedInstances, mappedInstances[mapAttr])
			}
		}
		if value, ok := mappedInstances[UNDEFINED_MAP]; ok {
			computedInstances = append(computedInstances, value)
		}

		return computedInstances
	}
	mapAndSortWorker.Run(fn)
}
