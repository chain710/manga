package cmd

import (
	"fmt"
	"github.com/chain710/manga/internal/strings"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"time"
)

var (
	viperBinds = strings.NewSet(nil)
)

func viperFlag(set *pflag.FlagSet, name string, value interface{}, usage string) {
	switch v := value.(type) {
	case int:
		set.IntP(name, "", v, usage)
	case string:
		set.StringP(name, "", v, usage)
	case time.Duration:
		set.DurationP(name, "", v, usage)
	}

	// delay viper.BindPFlag to RunE, to avoid same name entry conflict
	viperBinds.Add(name)
}

func viperBindPFlag(set *pflag.FlagSet) {
	set.VisitAll(func(flag *pflag.Flag) {
		if viperBinds.Contains(flag.Name) {
			if err := viper.BindPFlag(flag.Name, flag); err != nil {
				panic(fmt.Errorf("bind flag %s error: %s", flag.Name, err))
			}
		}
	})
}
