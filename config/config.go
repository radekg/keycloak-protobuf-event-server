package config

// ServerConfig represents the server configuration.
type ServerConfig struct {
	BindHostPort                   string
	NoTLS                          bool
	TLSTrustedCertificatesFilePath string
	TLSCertificateFilePath         string
	TLSKeyFilePath                 string
	GracefulStopTimeoutMillis      int
}

// LogConfig represents logging configuration.
type LogConfig struct {
	LogLevel      string
	LogColor      bool
	LogForceColor bool
	LogAsJSON     bool
}
