package serviceCenter

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	//"time"
)

const (
	leaseTTL int64 = 5 // second
)

type RegisterConfig struct {
	Host        string         `yaml:"Host"`
	Port        int            `yaml:"Port"`
	Prefix      string         `yaml:"Prefix"`
	Env         string         `yaml:"Env"`
	Service     SupportService `yaml:"Service"`
	ServiceAddr string         `yaml:"ServiceAddr"`
}

func RunRegister(config *RegisterConfig) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conf := clientv3.Config{
		Endpoints:   []string{addr},
		DialTimeout: dialTimeOut,
	}
	cli, err := clientv3.New(conf)
	if err != nil {
		panic(err)
	}

	defer cli.Close()

	lease := clientv3.NewLease(cli)

	// set lease expire
	leaseRes, grantErr := lease.Grant(context.Background(), leaseTTL)
	if grantErr != nil {
		panic(grantErr)
	}

	// register service node
	key := fmt.Sprintf("%s/%s/%s/%s", config.Prefix, config.Env, config.Service, config.ServiceAddr)
	_, putErr := clientv3.NewKV(cli).Put(context.Background(), key, config.ServiceAddr, clientv3.WithLease(leaseRes.ID))
	if putErr != nil {
		panic(putErr)
	}

	// auto lease
	klCh, klErr := lease.KeepAlive(context.Background(), leaseRes.ID)
	if klErr != nil {
		panic(klErr)
	}

	for {
		select {
		case res := <-klCh:
			if res != nil {
				//fmt.Println("lease success", time.Now())
				//fmt.Println(res)
			}
		}
	}

}
