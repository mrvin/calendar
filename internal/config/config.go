package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

func Parse(configPath string, conf any) error {
	if configPath == "" {
		if err := cleanenv.ReadEnv(conf); err != nil {
			return fmt.Errorf("read env: %w", err)
		}
		return nil
	}

	if err := cleanenv.ReadConfig(configPath, conf); err != nil {
		return fmt.Errorf("read config %s: %w", configPath, err)
	}

	return nil
}
