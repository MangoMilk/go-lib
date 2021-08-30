//	Micro Service Discovery
//	Author: Vincent Young
//	Version: v1.0.0
//
//	Based: Etcd
//	EtcdKeyRule: /{Prefix}/{SupportService}/{ServiceAddr}
//
//	Support:
//	1. Load Balance Strategy: polling、random
//	2. Auto find the normal node
//	3. Multi services
//	4. Service cluster
//	5. Optional etcd prefix
//
//	TODO:
//	1. Discover center cluster
//	2. Etcd value more service info (v1.0.0: Only addr string)
//	3. Optional service (v1.0.0: Need to hard code in discovery file)
//	4. Optimize error output
//	5. Add get all services method, so can implement load balance out of discovery (v1.0.0: Load balance in discovery)
//
package serviceCenter

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

const (
	dialTimeOut = time.Second * 5
	initTimeOut = time.Second * 5
)

type Discovery struct {
	Env         ServiceEnv
	loadBalance SupportLoadBalance
	serviceMap  ServiceMap
}

var D *Discovery

func (d *Discovery) watch(cli *clientv3.Client, prefix string) {
	wch := cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wresp := range wch {
		for _, e := range wresp.Events {
			switch e.Type {
			case mvccpb.PUT:
				D.serviceMap.put(string(e.Kv.Key), string(e.Kv.Value))
				break
			case mvccpb.DELETE:
				D.serviceMap.del(string(e.Kv.Key))
				break
			}
		}
	}
}

func (d *Discovery) GetService(s SupportService) (addr string) {

	if d.serviceMap[d.Env][s] != nil {
		services := d.serviceMap[d.Env][s]
		if l := len(services); l > 0 {
			switch d.loadBalance {
			case SupportLoadBalancePolling:
				addr = d.loadBalance.polling(services)
			case SupportLoadBalanceRandom:
				addr = d.loadBalance.random(services, 0, l)
			}
		}
	}

	return
}

func (d *Discovery) SetLoadBalance(lb SupportLoadBalance) {
	d.loadBalance = lb
}

func (d *Discovery) UpdateServiceStatus(s SupportService, addr string, status ServiceStatus) {
	for _, v := range d.serviceMap[d.Env][s] {
		if v.Addr == addr {
			v.Status = status
		}
	}
}

type DiscoveryConfig struct {
	Host        string
	Port        int
	Prefix      string
	Env         ServiceEnv
	LoadBalance SupportLoadBalance
}

func RunDiscovery(config *DiscoveryConfig) {
	// TODO support cluster discovery && auto find the work one.
	// dial
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{addr},
		DialTimeout: dialTimeOut,
	})

	if err != nil {
		panic(err)
	}

	defer cli.Close()

	// first get
	ctx, cancel := context.WithTimeout(context.Background(), initTimeOut)
	defer cancel()
	gresp, err := cli.Get(ctx, config.Prefix, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}

	// new Discovery
	D = &Discovery{
		Env:         config.Env,
		loadBalance: config.LoadBalance,
		serviceMap:  NewServiceMap(config.Env),
	}

	// assign serviceMap
	for _, v := range gresp.Kvs {
		D.serviceMap.put(string(v.Key), string(v.Value))
	}

	// watch
	D.watch(cli, config.Prefix)
}
