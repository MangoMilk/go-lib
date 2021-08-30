package serviceCenter

import "testing"

func TestCenter(t *testing.T) {
	prefix := "/micro_service"

	dConfig := DiscoveryConfig{
		Host:        "127.0.0.1",
		Port:        2379,
		Prefix:      prefix,
		Env:         ServiceEnvTest,
		LoadBalance: SupportLoadBalancePolling,
	}
	go RunDiscovery(&dConfig)

	rConfig := RegisterConfig{
		Host:        "127.0.0.1",
		Port:        2379,
		Prefix:      prefix,
		Env:         ServiceEnvTest,
		Service:     SupportServiceGood,
		ServiceAddr: "127.0.0.1:5002",
	}
	go RunRegister(&rConfig)

	select {}
}
