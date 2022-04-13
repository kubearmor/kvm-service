package config

import (
	"flag"
	"os"

	kg "github.com/kubearmor/KVMService/src/log"
	"github.com/spf13/viper"
)

// KVMServiceConfig struct
type KVMServiceConfig struct {
	Port     int
	EtcdPort int
	NonK8s   bool
}

// GlobalCfg Global configuration for Kubearmor
var GlobalCfg KVMServiceConfig

// ConfigPort
const ConfigPort string = "port"

// ConfigEtcdPort
var ConfigEtcdPort string = "etcd-port"

// ConfigNonK8s
var ConfigNonK8s string = "non-k8s"

func readCmdLineParams() {
	portPtr := flag.Int(ConfigPort, 32770, "Cluster Port")
	etcdPortPtr := flag.Int(ConfigEtcdPort, 2379, "Etcd Port")
	nonK8sPtr := flag.Bool(ConfigNonK8s, false, "Non K8s control plane")

	flag.Parse()

	viper.SetDefault(ConfigPort, *portPtr)
	viper.SetDefault(ConfigEtcdPort, *etcdPortPtr)
	viper.SetDefault(ConfigNonK8s, *nonK8sPtr)
}

// LoadConfig Load configuration
func LoadConfig() error {
	// Read configuration from command line
	readCmdLineParams()

	// Read configuration from env var
	// Note that the env var has to be set in uppercase for e.g, CLUSTER=xyz ./kubearmor
	viper.AutomaticEnv()

	// Read configuration from config file
	cfgfile := os.Getenv("KUBEARMOR_CFG")
	if cfgfile == "" {
		cfgfile = "kvmservice.yaml"
	}
	if _, err := os.Stat(cfgfile); err == nil {
		kg.Printf("setting config from file [%s]", cfgfile)
		viper.SetConfigFile(cfgfile)
		err := viper.ReadInConfig()
		if err != nil {
			return err
		}
	}

	GlobalCfg.Port = int(viper.GetInt32(ConfigPort))
	GlobalCfg.EtcdPort = int(viper.GetInt32(ConfigEtcdPort))
	GlobalCfg.NonK8s = viper.GetBool(ConfigNonK8s)

	kg.Printf("config [%+v]", GlobalCfg)

	return nil
}
