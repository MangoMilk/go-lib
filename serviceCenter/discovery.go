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
	"math/rand"
	"strings"
	"time"
)

const (
	// Etcd Key: /{Prefix}/{SupportService}/{ServiceAddr}
	serviceNameUnitIndex = 2
	serviceAddrUnitIndex = 3

	dialTimeOut = time.Second * 5
	initTimeOut = time.Second * 5
)

type SupportService string

const (
	SupportServiceGood    SupportService = "good"
	SupportServiceAddress SupportService = "address"
)

type ServiceStatus int

const (
	ServiceStatusNormal  ServiceStatus = 1
	ServiceStatusFailure ServiceStatus = 0
)

type Service struct {
	Name   SupportService
	Addr   string
	Status ServiceStatus
	// TODO more service node info to support more load balance strategy
	//Weight   int
	//Area     string
	//Distance int
	//Load     int
}

type ServiceMap map[SupportService][]*Service

func (m ServiceMap) put(k string, v string) {
	//fmt.Println("Put Service:")

	keyUnits := strings.Split(k, "/")

	if len(keyUnits) < 4 {
		return
	}

	name, addr := SupportService(keyUnits[serviceNameUnitIndex]), keyUnits[serviceAddrUnitIndex]

	for _, v := range m[name] {
		if v.Addr == addr {
			// TODO assign new service info
			v.Name = name
			v.Addr = addr
			v.Status = ServiceStatusNormal
			return
		}
	}

	m[name] = append(m[name], &Service{
		Name:   name,
		Addr:   addr,
		Status: ServiceStatusNormal,
	})

	//fmt.Println(m)
}

func (m ServiceMap) del(k string) {
	//fmt.Println("Del Service: ",k)

	keyUnits := strings.Split(k, "/")
	name := SupportService(keyUnits[serviceNameUnitIndex])
	addr := keyUnits[serviceAddrUnitIndex]
	for idx, v := range m[name] {
		if v.Addr == addr {
			m[name] = append(m[name][0:idx], m[name][idx+1:]...)
			return
		}
	}
}

type Discovery struct {
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

	if d.serviceMap[s] != nil {
		services := d.serviceMap[s]
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

type SupportLoadBalance int

func (lb SupportLoadBalance) polling(services []*Service) (addr string) {

	l := len(services)
	countL := l

	for countL > 0 {
		if pollingCount >= l {
			pollingCount = 0
		}

		if services[pollingCount%l].Status == ServiceStatusNormal {
			addr = services[pollingCount%l].Addr
			pollingCount++
			break
		}

		pollingCount++
		countL--
	}

	return
}

// range: [rangeFrom,rangeTo)
func (lb SupportLoadBalance) random(services []*Service, from int, to int) string {

	if len(services) == 0 || from >= to || from < 0 || to < 0 {
		return ""
	}

	if len(services) == 1 {
		if services[0].Status == ServiceStatusNormal {
			return services[0].Addr
		} else {
			return ""
		}
	}

	rand.Seed(time.Now().Unix())
	r := from + rand.Intn(to-from)
	if services[r].Status == ServiceStatusNormal {
		return services[r].Addr
	} else {
		var s []*Service
		if r == from {
			s = services[r+1:]
		} else if r == to-1 {
			s = services[:r]
		} else {
			s = append(services[:r], services[r+1:]...)
		}

		return lb.random(s, 0, len(s))
	}
}

var pollingCount = 0

const (
	SupportLoadBalancePolling SupportLoadBalance = iota
	SupportLoadBalanceRandom

	// TODO more strategy will open after more node info be supported.
	//SupportLoadBalanceWeight
	//SupportLoadBalanceDistance
	//SupportLoadBalanceLoad
)

func (d *Discovery) SetLoadBalance(lb SupportLoadBalance) {
	d.loadBalance = lb
}

func (d *Discovery) UpdateServiceStatus(s SupportService, addr string, status ServiceStatus) {
	for _, v := range d.serviceMap[s] {
		if v.Addr == addr {
			v.Status = status
		}
	}
}

func RunDiscovery(host string, port int, prefix string, lb SupportLoadBalance) {
	// TODO support cluster discovery && auto find the work one.
	// dial
	addr := fmt.Sprintf("%s:%d", host, port)
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
	gresp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}

	// new Discovery
	D = &Discovery{
		loadBalance: lb,
		serviceMap:  make(ServiceMap),
	}

	// assign serviceMap
	for _, v := range gresp.Kvs {
		D.serviceMap.put(string(v.Key), string(v.Value))
	}

	// watch
	D.watch(cli, prefix)
}
