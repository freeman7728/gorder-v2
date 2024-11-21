package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func init() {
	err := NewViperConfig()
	if err != nil {
		return
	}
}

var once sync.Once

func NewViperConfig() (err error) {
	once.Do(func() {
		err = newViperConfig()
	})
	return err
}

func newViperConfig() error {
	relPath, err := getRelativePathFromCaller()
	if err != nil {
		return err
	}
	viper.SetConfigName("global")
	viper.SetConfigType("yaml")
	//使用viper的文件与config的相对路径
	viper.AddConfigPath(relPath)
	viper.EnvKeyReplacer(strings.NewReplacer("_", "-"))
	viper.AutomaticEnv()
	_ = viper.BindEnv("stripe-key", "STRIPE_KEY")
	_ = viper.BindEnv("endpoint-stripe-secret", "ENDPOINT_STRIPE_SECRET")
	return viper.ReadInConfig()
}

func getRelativePathFromCaller() (relPath string, err error) {
	callPwd, err := os.Getwd()
	if err != nil {
		return
	}
	_, here, _, _ := runtime.Caller(0)
	relPath, err = filepath.Rel(callPwd, filepath.Dir(here))
	fmt.Println("caller from %s ,get relPath %s", callPwd, relPath)
	if err != nil {
		return "", err
	}
	return
}
