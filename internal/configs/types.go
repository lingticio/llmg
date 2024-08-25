package configs

type S3 struct {
	AccessKeyID     string `json:"access_key_id" yaml:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key" yaml:"secret_access_key"`
	BucketName      string `json:"bucket_name" yaml:"bucket_name"`
	Region          string `json:"region" yaml:"region"`
	Endpoint        string `json:"endpoint" yaml:"endpoint"`
}

type Redis struct {
	Host               string `json:"host" yaml:"host"`
	Port               string `json:"port" yaml:"port"`
	TLSEnabled         bool   `json:"tls_enabled" yaml:"tls_enabled"`
	Username           string `json:"username" yaml:"username"`
	Password           string `json:"password" yaml:"password"`
	DB                 int64  `json:"db" yaml:"db"`
	ClientCacheEnabled bool   `json:"client_cache_enabled" yaml:"client_cache_enabled"`
}

type Database struct {
	ConnectionString string `json:"connection_string" yaml:"connection_string"`
}
