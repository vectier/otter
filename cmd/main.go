package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vectier/otter/pkg/api"
	"github.com/vectier/otter/pkg/seaweedfs"
)

type config struct {
	MasterURL  string
	FilerURL   string
	ServerAddr string
}

func Run(args []string) int {
	config := setupEnv()
	stopCh := SetupSignalContext().Done()

	sc := seaweedfs.NewClient(config.MasterURL, config.FilerURL)

	s := api.NewServer(config.ServerAddr, sc)
	shutdownCh := s.Serve(stopCh)

	<-shutdownCh
	return 0
}

func setupEnv() *config {
	var (
		env        = getEnv("env", "development")
		masterURL  = getEnv("masterUrl", "http://localhost:9333")
		filerURL   = getEnv("filerUrl", "http://localhost:8888")
		serverAddr = getEnv("serverAddr", "localhost:14000")
	)

	if env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return &config{
		MasterURL:  masterURL,
		FilerURL:   filerURL,
		ServerAddr: serverAddr,
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		_ = os.Setenv(key, defaultValue)
		return defaultValue
	}
	return value
}
