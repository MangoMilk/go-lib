package serviceCenter

import (
	"math/rand"
	"time"
)

type SupportLoadBalance int

const (
	SupportLoadBalancePolling SupportLoadBalance = iota
	SupportLoadBalanceRandom

	// TODO more strategy will open after more node info be supported.
	//SupportLoadBalanceWeight
	//SupportLoadBalanceDistance
	//SupportLoadBalanceLoad
)

var pollingCount = 0

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
