package autobind

import (
	"context"
	"os"
	"testing"
	"time"

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
	assert.NoError(t, os.WriteFile(d+"/config.yaml", []byte(`
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

func TestShouldBindDuration(t *testing.T) {
	type Config struct {
		Duration time.Duration `viper:"duration"`
	}

	config := Config{}
	viper := viper.New()
	viper.Set("duration", "1s")

	autobinder := &Autobinder{
		ConfigObject: &config,
		Viper:        viper,
		UseNesting:   true,
	}

	autobinder.Bind(context.Background(), nil, []string{})

	assert.Equal(t, time.Second, config.Duration)
}

func TestShouldBindSliceOfSomething(t *testing.T) {
	type Something struct {
		Value string `viper:"value"`
	}

	type Config struct {
		Slice []Something `viper:"slice"`
	}

	d := t.TempDir()
	assert.NoError(t, os.WriteFile(d+"/config.yaml", []byte(`
slice:
  - value: test
`), 0600))

	config := Config{}
	viper := viper.New()
	viper.SetConfigFile(d + "/config.yaml")
	assert.NoError(t, viper.ReadInConfig())

	autobinder := &Autobinder{
		ConfigObject: &config,
		Viper:        viper,
		UseNesting:   true,
		Casters:      map[string]Caster{},
	}

	autobinder.Cast("slice", func(v interface{}) interface{} {
		arr := make([]Something, len(v.([]interface{})))
		for i, item := range v.([]interface{}) {
			arr[i] = Something{
				Value: item.(map[string]interface{})["value"].(string),
			}
		}
		return arr
	})

	autobinder.Bind(context.Background(), nil, []string{})

	assert.Equal(t, 1, len(config.Slice))
	assert.Equal(t, "test", config.Slice[0].Value)
}
