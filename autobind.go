package autobind

import (
	"context"
	"reflect"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ViperTag = "viper"
	CobraTag = "cobra"
	EnvTag   = "env"
)

type Autobinder struct {
	ConfigObject interface{}
	Viper        *viper.Viper
	UseNesting   bool
	EnvPrefix    string // Viper prefix for environment variables, viper does not expose this, and because we construct the ENV variables, the prefix isn't set by Viper.
}

func AutoBind(vp *viper.Viper, cfg interface{}) func(cmd *cobra.Command, args []string) error {
	binder := &Autobinder{
		ConfigObject: cfg,
		Viper:        vp,
		UseNesting:   true,
	}

	return func(cmd *cobra.Command, args []string) error {
		binder.Bind(cmd.Context(), cmd, []string{})
		return nil
	}
}

func (b *Autobinder) sub(subConfig interface{}) *Autobinder {
	return &Autobinder{
		ConfigObject: subConfig,
		Viper:        b.Viper,
		UseNesting:   b.UseNesting,
	}
}

func (b *Autobinder) Bind(ctx context.Context, cmd *cobra.Command, prefix []string) {
	log := log.Ctx(ctx)
	pv := reflect.ValueOf(b.ConfigObject)
	v := pv
	pt := v.Type()
	t := pt
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		logger := log.With().Str("field", t.Field(i).Name).Logger()
		logger.Debug().Msg("Binding field")
		f := v.Field(i)
		ft := t.Field(i)

		vip := ft.Tag.Get(ViperTag)
		pflg := ft.Tag.Get(CobraTag)
		env := ft.Tag.Get(EnvTag)

		nestedViperKey := strings.Join(append(prefix, vip), ".")

		// If the field is a struct, go deeper
		if ft.Type.Kind() == reflect.Struct {
			if b.UseNesting {
				b.sub(f.Addr().Interface()).Bind(ctx, cmd, append(prefix, vip))
			} else {
				b.sub(f.Addr().Interface()).Bind(ctx, cmd, []string{})
			}
			continue
		}

		if vip == "" {
			continue
		}

		if env != "" {
			nestedEnvKey := strings.Join(append(prefix, env), "_")
			if b.EnvPrefix != "" {
				nestedEnvKey = strings.ToUpper(b.EnvPrefix + "_" + nestedEnvKey)
			}

			logger.Trace().Str("env", nestedEnvKey).Msg("Binding env")
			b.Viper.BindEnv(nestedViperKey, nestedEnvKey) //nolint:errcheck
		}

		if pflg != "" && cmd.Flags().Lookup(pflg) != nil {
			logger.Trace().Str("pflag", pflg).Msg("Binding pflag")
			b.Viper.BindPFlag(nestedViperKey, cmd.Flags().Lookup(pflg)) //nolint:errcheck
		}

		if f.CanSet() {
			s := b.Viper.Get(nestedViperKey)
			logger.Debug().Interface("value", s).Msg("Setting value")
			f.Set(reflect.ValueOf(s))
		}
	}

	log.Info().Interface("config", b.ConfigObject).Msg("Bound configuration")
}
