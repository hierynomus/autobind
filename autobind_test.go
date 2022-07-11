package autobind

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestShouldDeepAutobind(t *testing.T) {
	type Config struct {
		Name   string `viper:"name"`
		Nested struct {
			Value int `viper:"value" env:"VALUE"`
		} `viper:"nested"`
	}

	config := Config{}
	viper := viper.New()
	viper.Set("name", "test")
	viper.Set("nested.value", 1)

	autobinder := &Autobinder{
		ConfigObject: &config,
		Viper:        viper,
		UseNesting:   true,
	}

	autobinder.Bind(context.Background(), nil, []string{})

	assert.Equal(t, "test", config.Name)
	assert.Equal(t, 1, config.Nested.Value)
}

func TestShouldAutobindFromConfig(t *testing.T) {
	type Config struct {
		Name   string `viper:"name"`
		Nested struct {
			Value int `viper:"value" env:"VALUE"`
		} `viper:"nested"`
	}

	d := t.TempDir()
	assert.NoError(t, ioutil.WriteFile(d+"/config.yaml", []byte(`
name: test
nested:
  value: 1
`), 0600))

	config := Config{}
	viper := viper.New()
	viper.SetConfigFile(d + "/config.yaml")
	assert.NoError(t, viper.ReadInConfig())

	autobinder := &Autobinder{
		ConfigObject: &config,
		Viper:        viper,
		UseNesting:   true,
	}

	autobinder.Bind(context.Background(), nil, []string{})

	assert.Equal(t, "test", config.Name)
	assert.Equal(t, 1, config.Nested.Value)
}

func TestShouldSetDefaults(t *testing.T) {
	type Config struct {
		Name string `viper:"name" default:"test"`
	}

	config := Config{}
	viper := viper.New()

	autobinder := &Autobinder{
		ConfigObject: &config,
		Viper:        viper,
		UseNesting:   true,
		SetDefaults:  true,
	}

	autobinder.Bind(context.Background(), nil, []string{})

	assert.Equal(t, "test", config.Name)
}
