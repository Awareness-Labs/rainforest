package config

type Config struct {
	Port       string   `mapstructure:"port"`
	Domain     string   `mapstructure:"domain"`
	HubURLs    []string `mapstructure:"hub-urls"`
	LeafPort   int      `mapstructure:"leaf-port"`
	StreamPath string   `mapstructure:"stream-path"`
	KVPath     string   `mapstructure:"kv-path"`
	SinkPath   string   `mapstructure:"sink-path"`
}
