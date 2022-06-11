package start

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/radekg/keycloak-protobuf-event-server/config"
	"github.com/radekg/keycloak-protobuf-event-server/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sharedConfig = new(config.ServerConfig)
var logConfig = new(config.LogConfig)

var Command = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Run:   run,
}

func initFlags() {
	// common:
	Command.Flags().StringVar(&sharedConfig.BindHostPort, "bind-host-port", ":5000", "Bind host port for the server")
	Command.Flags().BoolVar(&sharedConfig.NoTLS, "no-tls", false, "When set, server does not use TLS")
	Command.Flags().StringVar(&sharedConfig.TLSTrustedCertificatesFilePath, "tls-trusted-cert-file-path", "", "TLS trusted certificate file path")
	Command.Flags().StringVar(&sharedConfig.TLSCertificateFilePath, "tls-cert-file-path", "", "TLS certificate file path")
	Command.Flags().StringVar(&sharedConfig.TLSKeyFilePath, "tls-key-file-path", "", "TLS key file path")
	Command.Flags().IntVar(&sharedConfig.GracefulStopTimeoutMillis, "timeout-graceful-stop-millis", 5000, "How long to wait for graceful stop of the service")
	// logs:
	Command.Flags().StringVar(&logConfig.LogLevel, "log-level", "info", "Log level")
	Command.Flags().BoolVar(&logConfig.LogColor, "log-color", false, "Log with colors enabled")
	Command.Flags().BoolVar(&logConfig.LogForceColor, "log-force-color", false, "Force colors in log output even when terminal does not support colors")
	Command.Flags().BoolVar(&logConfig.LogAsJSON, "log-json", false, "Log output as JSON")

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func init() {
	initFlags()
}

func run(_ *cobra.Command, _ []string) {

	loggerColorOption := hclog.ColorOff
	if logConfig.LogColor {
		loggerColorOption = hclog.AutoColor
	}
	if logConfig.LogForceColor {
		loggerColorOption = hclog.ForceColor
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "keycloak-protobuf-event-server",
		Level:      hclog.LevelFromString(logConfig.LogLevel),
		Color:      loggerColorOption,
		JSONFormat: logConfig.LogAsJSON,
	})

	logger.Info("Starting server")

	server := server.NewServer(sharedConfig, logger)
	server.Start()

	select {
	case <-server.ReadyNotify():
	case <-server.FailedNotify():
		// TODO: log, crash
		logger.Error("Server failed to start", "reason", server.StartFailureReason())
		os.Exit(1)
	}

	logger.Info("Server running")

	waitForStop()

	logger.Info("Stopping server")

	server.Stop()

	<-server.StoppedNotify()

	logger.Info("All done, bye")

}

func waitForStop() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
