package configuration

import (
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// GetConfiguration loads the configuration from Viper and validates it.
func GetConfiguration(viper *viper.Viper, configStruct interface{}) error {
	// Load configuration from viper
	err := viper.Unmarshal(configStruct)
	if err != nil {
		return errors.Wrap(err, "Cannot unmarshall")
	}

	// Validate configuration
	validationErr := validator.New().Struct(configStruct)
	if validationErr != nil {

		var errs validator.ValidationErrors
		errors.As(validationErr, &errs)

		hasErr := false
		for _, fieldError := range errs {
			zap.L().Error("Validation failed on field", zap.Error(fieldError))
			hasErr = true
		}

		if hasErr {
			return errors.New("Validation of the config failed")
		}
	}

	return nil
}

// SetupEnv sets up the environment variables for the service.
func SetupEnv(serviceName string) {
	viper.SetEnvPrefix(serviceName)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// InitConfig initializes the configuration for the service, either from file or from ETCD (if ETCD was set).
func InitConfig(configurationFilePath string, additionalDirs ...string) {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	// Set defaults for searching for config files
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(home)

	for _, dir := range additionalDirs {
		viper.AddConfigPath(dir)
	}

	// Check if path is specified
	if configurationFilePath != "" {
		viper.SetConfigFile(configurationFilePath)

		// Read the configuration
		err := viper.ReadInConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				zap.L().With(zap.Error(err)).Error("Config file not found")
			} else {
				zap.L().With(zap.Error(err)).Fatal("Something went wrong")
			}
		}
	}
}
