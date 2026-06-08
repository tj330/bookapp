package main

// config struct consists of all the required configurations
// required for the application.
type config struct {
	API              apiConfig              `yaml:"api"`
	ServiceDiscovery serviceDiscoveryConfig `yaml:"serviceDiscovery"`
	Jaeger           jaegerConfig           `yaml:"jaeger"`
	Jwt              jwtConfig              `yaml:"jwt"`
	Prometheus       prometheusConfig       `yaml:"prometheus"`
}

type apiConfig struct {
	Port int `yaml:"port"`
}

type serviceDiscoveryConfig struct {
	Consul consulConfig `yaml:"consul"`
}

type consulConfig struct {
	Address string `yaml:"address"`
}

type jaegerConfig struct {
	URL string `yaml:"url"`
}

type jwtConfig struct {
	Secret string `yaml:"secret"`
}

type prometheusConfig struct {
	MetricsPort int `yaml:"metricsPort"`
}
