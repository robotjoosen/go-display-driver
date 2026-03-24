package env

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func StringToPathHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.String {
			return data, nil
		}
		return expandTilde(data.(string)), nil
	}
}

func Load[T any](scope map[string]any, transformers ...func(*T)) (env T, err error) {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	for key, value := range scope {
		viper.SetDefault(key, value)
	}

	err = viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return env, err
		}
	}

	err = viper.Unmarshal(&env,
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.TextUnmarshallerHookFunc(),
			StringToPathHookFunc(),
		)),
		func(config *mapstructure.DecoderConfig) {
			config.IgnoreUntaggedFields = true
		},
	)

	for _, transformer := range transformers {
		transformer(&env)
	}

	return env, err
}
