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
	kv := consul.KV()
	uuidV, _ := uuid.GenerateUUID()
	pair := consulapi.KVPair{}
	pair.Key = fmt.Sprintf("%s/%s", prefix, uuidV)
	pair.Value = []byte(pair.Key)
	wOpts := consulapi.WriteOptions{RelayFactor: 4}

	for i := 0; i < count; i++ {

		pair.Value = []byte(fmt.Sprintf("%i", i))
		ret, err := kv.Put(&pair, &wOpts)
		fmt.Printf("%v / %v", ret, err)
	}
}
