package main

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	uuid2 "github.com/hashicorp/go-uuid"
	"math/rand"
	"time"
)

func main() {
	config := consulapi.DefaultConfig()
	consul, _ := consulapi.NewClient(config)
	catalog := consul.Catalog()
	services, _, _ := catalog.Services(&consulapi.QueryOptions{})
	fmt.Printf("service %v \n", services)
	for k, v := range services {
		fmt.Printf("%v | %v \n", k, v)
	}

}
