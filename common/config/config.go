package common

type Config struct {
	Port            string                   `toml:"port"`
	ServerMode      string                   `toml:"server_mode"`
	DebugLevel      bool                     `toml:"debug_level"`
	ErrorLogFile    string                   `toml:"error_log_file"`
	ServiceID       string                   `toml:"service_id"`
	ProviderID      string                   `toml:"provider_id"`
	TimeZone        string                   `toml:"timezone"`
	Referer         Referer                  `toml:"referer"`
	RejectExtension RejectExtension          `toml:"reject_extension"`
	Cors            Cors                     `toml:"cors"`
	ServiceDB       DatabaseConnection       `toml:"service_db"`
	BodyLog         BodyLog                  `toml:"body_log"`
	MongoDB         MongoDBConnection        `toml:"mongo_db"`
	Redis           Redis                    `toml:"redis"`
	Auth            map[string]Authorization `toml:"auth"`
	GoogleOAuth2    GoogleOAuth2             `toml:"google_oauth2"`
	Trace           Trace                    `toml:"trace"`
}

type Cors struct {
	Methods []string `toml:"methods"`
	Origins []string `toml:"origins"`
	Headers []string `toml:"headers"`
}

type Referer struct {
	RefererDomains []string `toml:"referer_domains"`
}

type RejectExtension struct {
	CompressedFileExtensions []string `toml:"compressed_file_extensions"`
}

type DatabaseConnection struct {
	DriverName     string `toml:"driver_name"`
	DataSourceName string `toml:"data_source_name"`
	RootSourceName string `toml:"root_source_name"`
	LogMode        bool   `toml:"log_mode"`
	MaxIdleConns   int    `toml:"max_idle_conns"`
	MaxOpenConns   int    `toml:"max_open_conns"`
}

type BodyLog struct {
	EnableRequestBody   bool `toml:"enable_request_body"`
	EnableResponsesBody bool `toml:"enable_response_body"`
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
	Port            int      `toml:"port"`
	Database        Database `toml:"database"`
}

type Authorization struct {
	AuthType     string `toml:"type"`
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
}

type GoogleOAuth2 struct {
	CallbackURL  string `toml:"callback_url"`
	OathUrlAPI   string `toml:"oath_google_url_api"`
	ScopeEmail   string `toml:"scope_email"`
	ScopeProfile string `toml:"scope_profile"`
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
}

type Redis struct {
	AuthStore    string `toml:"auth_store" env:"REDIS_AUTH_STORE"`
	AuthStorePwd string `toml:"auth_store_pwd" env:"REDIS_AUTH_STORE_PWD"`
}

type MongoDBConnection struct {
	URL      string `toml:"url"`
	Database string `toml:"database"`
}

type Trace struct {
	JaegerAgentHost string  `toml:"jaeger_agent_host" env:"JAEGER_AGENT_HOST"`
	SamplerType     string  `toml:"sampler_type" env:"JAEGER_SAMPLER_TYPE"`
	SamplerParam    float64 `toml:"sampler_param" env:"JAEGER_SAMPLER_PARAM"`
}
