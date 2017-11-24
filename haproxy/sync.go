package main

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"net"
	"time"
)

func main() {
	config := consulapi.DefaultConfig()
	services(config)

}

func services(config *consulapi.Config) {
	var active = make(map[string]chan int)
	var ok bool
	index := uint64(0)
	consul, _ := consulapi.NewClient(config)
	catalog := consul.Catalog()
	options := consulapi.QueryOptions{
		WaitIndex: index,
		WaitTime:  50 * time.Second,
	}
	for {
		services, res, _ := catalog.Services(&options)
		fmt.Printf("service %v \n", services)
		for k, v := range services {

			fmt.Printf("%v | %v \n", k, v)
			_, ok = active[k]
			if !ok {
				active[k] = make(chan int)
				go monitor(config, k, active[k])

			}
			options.WaitIndex = res.LastIndex

		} //close disabled services
		for k, v := range active {

			_, ok = services[k]
			if !ok {
				v <- 1
				delete(active, k)

			}
		}

	}

}

func monitor(config *consulapi.Config, service string, control chan int) {
	consul, _ := consulapi.NewClient(config)
	health := consul.Health()
	index := uint64(0)
	options := consulapi.QueryOptions{
		WaitIndex: index,
		WaitTime:  50 * time.Second,
	}
	for {
		nodes, headers, err := health.Service(service, "", false, &options)
		for x := 0; x < len(nodes); x++ {
			passing := true
			for y := 0; y < len(nodes[x].Checks); y++ {
				if nodes[x].Checks[y].Status != "passing" {
					passing = false
				}
			}
			if passing {
				fmt.Printf("enable %v for pool %v \n", nodes[x].Node.Node, service)
			} else {
				fmt.Printf("disable %v for pool %v \n", nodes[x].Node.Node, service)
			}
		}
		if err == nil {
			options.WaitIndex = headers.LastIndex

		}

		select {
		case <-control:
			fmt.Printf("disabling monitoring of pool %v \n", service)
			return
		default:
			// receiving from channel would block
			fmt.Printf("still  monitoring of pool %v \n", service)
		}

	}

}

/*
Changing state in haproxy per documentation https://cbonte.github.io/haproxy-dconv/1.6/management.html#9.2
*/

func changehaproxy(service, node, status string) {
	haproxy, err := net.Dial("unix", "/var/haproxy.socket")
	if err != nil {
		fmt.Printf("error conntecting to haprxy: %v", err)
	}
	if status == "passing" {
		//renable disiabled server
		fmt.Fprintf(haproxy, "enable server %s/%s\r\n", service, node)
	} else {
		//put server in maintance
		fmt.Fprintf(haproxy, "disable server %s/%s\r\n", service, node)
	}
}
