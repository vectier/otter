package main

import (
	"github.com/vectier/otter/pkg/api"
	"github.com/vectier/otter/pkg/seaweedfs"
)

// TODO: move url to config file
func main() {
	sc := seaweedfs.NewClient("http://localhost:9333", "http://localhost:8888")

	s := api.NewServer("localhost:14000", sc)
	shutdownCh := s.Serve()

	<-shutdownCh
}
