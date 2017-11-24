package kv

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-uuid"
	"sync"
)

func main() {
	config := consulapi.DefaultConfig()
	var wg sync.WaitGroup
	wg.Add(1000)

	for i := 0; i < 1000; i++ {
		go kvAssault(config, 1000, fmt.Sprintf("%i", i), wg)
	}
	wg.Wait()

}

func kvAssault(config *consulapi.Config, count int, prefix string, wg sync.WaitGroup) {
	defer wg.Done()
	consul, _ := consulapi.NewClient(config)
	consul.Catalog()
}
