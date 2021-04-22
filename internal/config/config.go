package config

import (
	"github.com/Confialink/wallet-pkg-env_config"
)

type Config struct {
	Env          string
	Db           *env_config.Db
	Port         string
	ProtoBufPort string
	Cors         *env_config.Cors
	AwsConfig    AwsConfig
	Storage      string
}

type AwsConfig struct {
	S3Bucket string
	Region   string
}
