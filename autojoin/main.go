package main

import (
	"fmt"

	consulapi "github.com/hashicorp/consul/api"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"os"
	"regexp"
)

var regionRe = regexp.MustCompile(`(\w*-\w*\d)-\w`)

func main() {
	project := os.Args[1]
	tag := os.Args[2]

	ctx := context.TODO()
	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		fmt.Errorf("Failed to connect to gcp: %v ", err)
	}
	computeService, err := compute.New(client)
	if err != nil {
		fmt.Errorf("Error : %v", err)
	}
	regionMap, _ := gceGetTaggedNodesByRegion(ctx, computeService, project, tag)
	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	agent := consul.Agent()
	joinConsulWan(regionMap, agent)

}

// joinConsulWan Connects to previously unknown DC
//
func joinConsulWan(regionMap map[string][]string, agent *consulapi.Agent) error {
	/*if regionMap == nil {
		return error("Region Map is empty")
	}*/
	for k, nodes := range regionMap {
		fmt.Printf("Region: %s, Nodes: %v \n", k, nodes)

		for _, node := range nodes {
			agent.Join(node, true)
		}

	}
	return nil
}

// gceGetalltaggednodesbyregion  discovers all nodes that match tag and maps them
// by region.
func gceGetTaggedNodesByRegion(ctx context.Context, computeService *compute.Service, project, tag string) (map[string][]string, error) {
	var ret = make(map[string][]string)
	zones, _ := gceDiscoverZones(ctx, computeService, project)
	for _, zone := range zones {
		instances, _ := gceInstancesAddressesForZone(ctx, computeService, project, zone, tag)
		region := regionRe.FindStringSubmatch(zone)[1]
		if len(instances) > 0 {
			if ret[region] == nil {
				ret[region] = make([]string, 0)
			}
			ret[region] = append(ret[region], instances...)
		}
	}
	return ret, nil
}

// gceDiscoverZones discovers a list of zones from a supplied zone pattern, or
// all of the zones available to a project.
func gceDiscoverZones(ctx context.Context, computeService *compute.Service, project string) ([]string, error) {
	var zones []string

	call := computeService.Zones.List(project)

	if err := call.Pages(ctx, func(page *compute.ZoneList) error {
		for _, v := range page.Items {
			zones = append(zones, v.Name)
		}
		return nil
	}); err != nil {
		return zones, err
	}

	return zones, nil
}

// gceInstancesAddressesForZone locates all instances within a specific project
// and zone, matching the supplied tag. Only the private IP addresses are
// returned, but ID is also logged.
func gceInstancesAddressesForZone(ctx context.Context, computeService *compute.Service, project, zone, tag string) ([]string, error) {
	var addresses []string
	call := computeService.Instances.List(project, zone)
	if err := call.Pages(ctx, func(page *compute.InstanceList) error {
		for _, v := range page.Items {
			for _, t := range v.Tags.Items {
				if t == tag && len(v.NetworkInterfaces) > 0 && v.NetworkInterfaces[0].NetworkIP != "" {
					addresses = append(addresses, v.NetworkInterfaces[0].NetworkIP)
				}
			}
		}
		return nil
	}); err != nil {
		return addresses, err
	}
	return addresses, nil
}
