// Copyright 2019-2020 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !privileged_tests

package ipam

import (
	"context"
	"math"

	apimock "github.com/cilium/cilium/pkg/azure/api/mock"
	"github.com/cilium/cilium/pkg/azure/types"
	"github.com/cilium/cilium/pkg/cidr"
	ipamTypes "github.com/cilium/cilium/pkg/ipam/types"

	"gopkg.in/check.v1"
)

var (
	subnets = []*ipamTypes.Subnet{
		{
			ID:               "subnet-1",
			CIDR:             cidr.MustParseCIDR("1.1.0.0/16"),
			VirtualNetworkID: "vpc-1",
			Tags: map[string]string{
				"tag1": "tag1",
			},
		},
		{
			ID:               "subnet-2",
			CIDR:             cidr.MustParseCIDR("2.2.0.0/16"),
			VirtualNetworkID: "vpc-2",
			Tags: map[string]string{
				"tag1": "tag1",
			},
		},
	}

	subnets2 = []*ipamTypes.Subnet{
		{
			ID:               "subnet-1",
			CIDR:             cidr.MustParseCIDR("1.1.0.0/16"),
			VirtualNetworkID: "vpc-1",
			Tags: map[string]string{
				"tag1": "tag1",
			},
		},
		{
			ID:               "subnet-2",
			CIDR:             cidr.MustParseCIDR("2.2.0.0/16"),
			VirtualNetworkID: "vpc-2",
			Tags: map[string]string{
				"tag1": "tag1",
			},
		},
		{
			ID:               "subnet-3",
			CIDR:             cidr.MustParseCIDR("3.3.0.0/16"),
			VirtualNetworkID: "vpc-1",
			Tags: map[string]string{
				"tag2": "tag2",
			},
		},
	}

	vnets = []*ipamTypes.VirtualNetwork{
		{ID: "vpc-0"},
		{ID: "vpc-1"},
	}

	instances = types.InstanceMap{
		"i-1": &types.Instance{
			Interfaces: map[string]*types.AzureInterface{
				"intf-1": {
					ID:            "intf-1",
					SecurityGroup: "sg1",
					Addresses: []types.AzureAddress{
						{
							IP:     "1.1.1.1",
							Subnet: "subnet-1",
							State:  types.StateSucceeded,
						},
					},
					State: types.StateSucceeded,
				},
			},
		},
		"i-2": &types.Instance{
			Interfaces: map[string]*types.AzureInterface{
				"intf-3": {
					ID:            "intf-3",
					SecurityGroup: "sg3",
					Addresses: []types.AzureAddress{
						{
							IP:     "1.1.3.3",
							Subnet: "subnet-1",
							State:  types.StateSucceeded,
						},
					},
					State: types.StateSucceeded,
				},
			},
		},
	}

	instances2 = types.InstanceMap{
		"i-1": &types.Instance{
			Interfaces: map[string]*types.AzureInterface{
				"intf-1": {
					ID:            "intf-1",
					SecurityGroup: "sg1",
					Addresses: []types.AzureAddress{
						{
							IP:     "1.1.1.1",
							Subnet: "subnet-1",
							State:  types.StateSucceeded,
						},
					},
					State: types.StateSucceeded,
				},
				"intf-2": {
					ID:            "intf-2",
					SecurityGroup: "sg2",
					Addresses: []types.AzureAddress{
						{
							IP:     "3.3.3.3",
							Subnet: "subnet-3",
							State:  types.StateSucceeded,
						},
					},
					State: types.StateSucceeded,
				},
			},
		},
		"i-2": &types.Instance{
			Interfaces: map[string]*types.AzureInterface{
				"intf-3": {
					ID:            "intf-3",
					SecurityGroup: "sg3",
					Addresses: []types.AzureAddress{
						{
							IP:     "1.1.3.3",
							Subnet: "subnet-1",
							State:  types.StateSucceeded,
						},
					},
					State: types.StateSucceeded,
				},
			},
		},
	}
)

func iteration1(api *apimock.API, mngr *InstancesManager) {
	api.UpdateInstances(instances)
	mngr.Resync(context.Background())
}

func iteration2(api *apimock.API, mngr *InstancesManager) {
	api.UpdateSubnets(subnets2)
	api.UpdateInstances(instances2)
	mngr.Resync(context.TODO())
}

func (e *IPAMSuite) TestGetVpcsAndSubnets(c *check.C) {
	api := apimock.NewAPI(subnets, vnets)
	c.Assert(api, check.Not(check.IsNil))

	mngr := NewInstancesManager(api)
	c.Assert(mngr, check.Not(check.IsNil))

	c.Assert(mngr.getAllocator().PoolExists("subnet-1"), check.Equals, false)
	c.Assert(mngr.getAllocator().PoolExists("subnet-2"), check.Equals, false)
	c.Assert(mngr.getAllocator().PoolExists("subnet-3"), check.Equals, false)

	iteration1(api, mngr)

	c.Assert(mngr.getAllocator().PoolExists("subnet-1"), check.Equals, true)
	c.Assert(mngr.getAllocator().PoolExists("subnet-2"), check.Equals, true)
	c.Assert(mngr.getAllocator().PoolExists("subnet-3"), check.Equals, false)

	iteration2(api, mngr)

	c.Assert(mngr.getAllocator().PoolExists("subnet-1"), check.Equals, true)
	c.Assert(mngr.getAllocator().PoolExists("subnet-2"), check.Equals, true)
	c.Assert(mngr.getAllocator().PoolExists("subnet-3"), check.Equals, true)
}

func (e *IPAMSuite) TestInstances(c *check.C) {
	api := apimock.NewAPI(subnets, vnets)
	c.Assert(api, check.Not(check.IsNil))

	mngr := NewInstancesManager(api)
	c.Assert(mngr, check.Not(check.IsNil))

	iteration1(api, mngr)
	c.Assert(len(mngr.GetInterfaces("i-1")), check.Equals, 1)
	c.Assert(len(mngr.GetInterfaces("i-2")), check.Equals, 1)
	c.Assert(len(mngr.GetInterfaces("i-unknown")), check.Equals, 0)

	iteration2(api, mngr)
	c.Assert(len(mngr.GetInterfaces("i-1")), check.Equals, 2)
	c.Assert(len(mngr.GetInterfaces("i-2")), check.Equals, 1)
	c.Assert(len(mngr.GetInterfaces("i-unknown")), check.Equals, 0)
}

func (e *IPAMSuite) TestPoolQuota(c *check.C) {
	api := apimock.NewAPI(subnets, vnets)
	c.Assert(api, check.Not(check.IsNil))

	mngr := NewInstancesManager(api)
	c.Assert(mngr, check.Not(check.IsNil))

	quota := mngr.GetPoolQuota()
	c.Assert(len(quota), check.Equals, 0)

	iteration1(api, mngr)
	quota = mngr.GetPoolQuota()
	c.Assert(len(quota), check.Equals, 2)
	// 2 IPs should be allocated
	c.Assert(quota["subnet-1"].AvailableIPs, check.Equals, int(math.Pow(2.0, 16.0)-4))
	// No IPs should be allocated
	c.Assert(quota["subnet-2"].AvailableIPs, check.Equals, int(math.Pow(2.0, 16.0)-2))

	iteration2(api, mngr)
	quota = mngr.GetPoolQuota()
	c.Assert(len(quota), check.Equals, 3)
	// 2 IPs should be allocated
	c.Assert(quota["subnet-1"].AvailableIPs, check.Equals, int(math.Pow(2.0, 16.0)-4))
	// No IP should be allocated
	c.Assert(quota["subnet-2"].AvailableIPs, check.Equals, int(math.Pow(2.0, 16.0)-2))
	// 1 IP should be allocated
	c.Assert(quota["subnet-3"].AvailableIPs, check.Equals, int(math.Pow(2.0, 16.0)-3))
}
