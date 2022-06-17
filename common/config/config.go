package config

type Config struct {
	Port            string                   `toml:"port" default:"80" env:"PORT"`
	ServerMode      string                   `toml:"server_mode" env:"SERVER_MODE"`
	TimeZone        string                   `toml:"timezone"`
	Cors            string                   `toml:"cors"`
	ServiceDB       DatabaseConnection       `toml:"service_db"`
	MongoDB         MongoDBConnection        `toml:"mongo_db"`
	Redis           Redis                    `toml:"redis"`
	Auth            map[string]Authorization `toml:"auth"`
	AdminOAuth2     AdminOAuth2              `toml:"adminOAuth2"`
	Referer         Referer                  `toml:"referer"`
	RejectExtension RejectExtension          `toml:"reject_extension"`
	DebugLevel      bool                     `toml:"debug_level"`
	BodyLog         BodyLog                  `toml:"body_log"`
	ErrorLogFile    string                   `toml:"error_log_file"`
	ServiceID       string                   `toml:"service_id"`
	ProviderID      string                   `toml:"provider_id"`
}

type Cors struct {
	Methods []string `toml:"methods" env:"CORS_METHODS"`
	Origins []string `toml:"origins" env:"CORS_ORIGINS"`
	Headers []string `toml:"headers" env:"CORS_HEADERS"`
}

type Referer struct {
	RefererDomains []string `toml:"referer_domains" env:"REFERER_DOMAINS"`
}

type RejectExtension struct {
	CompressedFileExtensions []string `toml:"compressed_file_extensions"`
}

type Database struct {
	DriverName     string `toml:"driver_name"`
	DataSourceName string `toml:"data_source_name"`
	RootSourceName string `toml:"root_source_name"`
}

type APIServerInfo struct {
	Host            string   `toml:"host"`
	GRPCPort        int      `toml:"grpc_port"`
	GRPCGatewayPort int      `toml:"grpc_gateway_port"`
	Port            int      `toml:"database"`
	Database        Database `toml:"database"`
}

type Authorization struct {
	AuthType     string `toml:"type"`
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
}

type AdminOAuth2 struct {
	ClientID      string `toml:"client_id"`
	ClientSecret  string `toml:"client_secret"`
	RedirectURI   string `toml:"redirect_uri"`
	Authorization string `toml:"authorization"`
	URL           string `toml:"url"`
}

type Redis struct {
	AuthStore    string `toml:"auth_store" env:"REDIS_AUTH_STORE"`
	AuthStorePwd string `toml:"auth_store_pwd" env:"REDIS_AUTH_STORE_PWD"`
}

type DatabaseConnection struct {
	DriverName     string `toml:"driver_name" env:"DRIVER_NAME"`
	DataSourceName string `toml:"data_source_name" env:"DATA_SOURCE_NAME"`
	RootSourceName string `toml:"root_source_name" env:"ROOT_SOURCE_NAME"`
	LogMode        bool   `toml:"log_mode" env:"LOG_MODE"`
	MaxIdleConns   int    `toml:"max_idle_conns" env:"MAX_IDLE_CONNS"`
	MaxOpenConns   int    `toml:"max_open_conns" env:"MAX_OPEN_CONNS"`
}

type MongoDBConnection struct {
	URL      string `toml:"url"`
	Database string `toml:"database"`
}

type BodyLog struct {
	EnableRequestBody   bool `toml:"enable_request_body" default:"true" env:"ENABLE_REQUEST_BODY_LOG"`
	EnableResponsesBody bool `toml:"enable_response_body" default:"true" env:"ENABLE_RESPONSE_BODY_LOG"`
}
