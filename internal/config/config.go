package config

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"
)

const VERSION = "unknown"

type Config struct {
	Env            string `yaml:"env" env:"ENV" env-default:"dev"`
	Cache          Cache
	Http           Http
	PartitionedMap PartitionedMap
}

type PartitionedMap struct {
	Partitions    uint64 `yaml:"partitions" env:"PARTITIONS" env-default:"2"`
	PartitionSize uint64 `yaml:"partition_size" env:"PARTITION_SIZE" env-default:"10"`
}

type Cache struct {
	Driver string `yaml:"driver" env:"CACHE_DRIVER" env-default:"partitioned_map"`
}

type Http struct {
	Addr string `yaml:"addr" env:"HTTP_ADDR" env-default:":8000"`
}

func MustLoad() Config {
	var cfgFile string

	flag.StringVar(&cfgFile, "config", "config/dev.yml", "path to config file")
	flag.Parse()

	cfg := Config{}

	if err := cleanenv.ReadConfig(cfgFile, &cfg); err != nil {
		panic(err)
	}

	return cfg
}
