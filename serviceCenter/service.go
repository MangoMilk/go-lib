package serviceCenter

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// Etcd Key: /{Prefix}/{Env}/{SupportService}/{ServiceAddr}
	keyUnitCount           = 4
	separator              = "/"
	servicePrefixUnitIndex = 0
	serviceEnvUnitIndex    = 1
	serviceNameUnitIndex   = 2
	serviceAddrUnitIndex   = 3
)

type ServiceEnv string

const (
	ServiceEnvTest   = "test"
	ServiceEnvOnline = "online"
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
	Env    ServiceEnv
	Name   SupportService
	Addr   string
	Status ServiceStatus
	// TODO more service node info to support more load balance strategy
	//Weight   int
	//Area     string
	//Distance int
	//Load     int
}

type ServiceMap map[ServiceEnv]map[SupportService][]*Service

func (m ServiceMap) put(k string, v string) {
	var service Service
	if err := m.parse(k, &service); err != nil {
		fmt.Println(err)
		return
	}

	services := m[service.Env][service.Name]

	for _, v := range services {
		if v.Addr == service.Addr {
			m.update(v, service.Env, service.Name, service.Addr, ServiceStatusNormal)
			return
		}
	}

	m[service.Env][service.Name] = append(services, &service)
}

func (m ServiceMap) del(k string) {
	var service Service
	if err := m.parse(k, &service); err != nil {
		fmt.Println(err)
		return
	}

	services := m[service.Env][service.Name]

	for idx, v := range services {
		if v.Addr == service.Addr {
			m[service.Env][service.Name] = append(services[0:idx], services[idx+1:]...)
			return
		}
	}
}

func (m ServiceMap) parse(k string, service *Service) error {
	keyUnits := strings.Split(strings.Trim(k, separator), separator)

	if len(keyUnits) < keyUnitCount {
		return errors.New(fmt.Sprintf("【Error】service key format error: %s", k))
	}

	m.update(
		service,
		ServiceEnv(keyUnits[serviceEnvUnitIndex]),
		SupportService(keyUnits[serviceNameUnitIndex]),
		keyUnits[serviceAddrUnitIndex],
		ServiceStatusNormal,
	)

	servicesEnv, hasEnv := m[service.Env]
	if !hasEnv {
		return errors.New(fmt.Sprintf("【Error】service env not found: %s", k))
	}

	_, hasService := servicesEnv[service.Name]
	if !hasService {
		return errors.New(fmt.Sprintf("【Error】service not support: %s", k))
	}

	return nil
}

func (m ServiceMap) update(service *Service, env ServiceEnv, name SupportService, addr string, status ServiceStatus) {
	// TODO assign new service info
	service.Env = env
	service.Name = name
	service.Addr = addr
	service.Status = status
}

func NewServiceMap(env ServiceEnv) ServiceMap {

	return ServiceMap{
		env: {
			SupportServiceGood:    nil,
			SupportServiceAddress: nil,
		}}
}
