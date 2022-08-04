package cmd

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func viperFlag(set *pflag.FlagSet, name string, value interface{}, usage string) error {
	switch v := value.(type) {
	case int:
		set.IntP(name, "", v, usage)
	case string:
		set.StringP(name, "", v, usage)
	}

	return viper.BindPFlag(name, set.Lookup(name))
}
