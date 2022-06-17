package autobind

import (
	"context"
	"reflect"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ViperTag = "viper"
	CobraTag = "cobra"
	EnvTag   = "env"
)

func AutoBind(vp *viper.Viper, cfg interface{}) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		autoBind(cmd.Context(), cmd, vp, cfg)
		return nil
	}
}

func autoBind(ctx context.Context, cmd *cobra.Command, vp *viper.Viper, cfg interface{}) {
	log := log.Ctx(ctx)
	pv := reflect.ValueOf(cfg)
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

		// If the field is a struct, go deeper
		if ft.Type.Kind() == reflect.Struct {
			autoBind(ctx, cmd, vp, f.Addr().Interface())
			continue
		}

		vip := ft.Tag.Get(ViperTag)
		pflg := ft.Tag.Get(CobraTag)
		env := ft.Tag.Get(EnvTag)

		if vip == "" {
			continue
		}

		if env != "" {
			logger.Trace().Str("env", env).Msg("Binding env")
			vp.BindEnv(vip, env) //nolint:errcheck
		}

		if pflg != "" && cmd.Flags().Lookup(pflg) != nil {
			logger.Trace().Str("pflag", pflg).Msg("Binding env")
			vp.BindPFlag(vip, cmd.Flags().Lookup(pflg)) //nolint:errcheck
		}

		if f.CanSet() {
			s := vp.GetString(vip)
			logger.Debug().Str("value", s).Msg("Setting value")
			f.Set(reflect.ValueOf(vp.GetString(vip)))
		}
	}

	log.Info().Interface("config", cfg).Msg("Bound configuration")
}
