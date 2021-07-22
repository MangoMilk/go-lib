package serviceCenter

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	leaseTTL int64 = 5 // second
)

func RunRegister(host string, port int, prefix string, service SupportService, serviceAddr string) {
	addr := fmt.Sprintf("%s:%d", host, port)
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
	key := fmt.Sprintf("%s/%s/%s", prefix, service, serviceAddr)
	_, putErr := clientv3.NewKV(cli).Put(context.Background(), key, serviceAddr, clientv3.WithLease(leaseRes.ID))
	if putErr != nil {
		panic(putErr)
	}

	//fmt.Println("Put Res: ")
	//fmt.Println(putRes)

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
