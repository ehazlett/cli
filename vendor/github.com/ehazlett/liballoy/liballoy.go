package liballoy

import (
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/homedir"
)

type Alloy struct {
	client   client.APIClient
	cacheDir string
	variants map[string]*Variant
}

type Variant struct {
	Version    string
	Image      string
	Entrypoint string
}

type Config struct {
	Variants []*Variant
	Client   client.APIClient
}

func New(cfg *Config) (*Alloy, error) {
	variants := map[string]*Variant{}
	for _, v := range cfg.Variants {
		variants[v.Version] = v
	}

	c, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &Alloy{
		client:   c,
		cacheDir: filepath.Join(homedir.Get(), ".alloy"),
		variants: variants,
	}, nil
}
