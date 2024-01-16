package main

type config struct {
	ServiceName string           `yaml:"serviceName"`
	Consul      consulConfig     `yaml:"consul"`
	API         apiConfig        `yaml:"api"`
	Jaeger      jaegerConfig     `yaml:"jaeger"`
	Prometheus  prometheusConfig `yaml:"prometheus"`
}

type apiConfig struct {
	Port int `yaml:"port"`
}

type consulConfig struct {
	URL string `yaml:"url"`
}

type jaegerConfig struct {
	URL string `yaml:"url"`
}

type prometheusConfig struct {
	MetricsPort int `yaml:"metricsPort"`
}
