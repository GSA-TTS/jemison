package env

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/GSA-TTS/jemison/config"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Env *env
var DEBUG_ENV = false

// Constants for the attached services
// These reach into the VCAP_SERVICES and are
// defined in the Terraform.
const QueueDatabase = "jemison-queues-db"
const JemisonWorkDatabase = "jemison-work-db"
const SearchDatabase = "jemison-search-db"

var validBucketNames = []string{
	"extract",
	"fetch",
	"serve",
}

func IsValidBucketName(name string) bool {
	for _, v := range validBucketNames {
		if name == v {
			return true
		}
	}
	return false
}

type Credentials interface {
	CredentialString(string) string
	CredentialInt(string) int64
}

type Service struct {
	Name        string                 `mapstructure:"name"`
	Credentials map[string]interface{} `mapstructure:"credentials"`
	Parameters  map[string]interface{} `mapstructure:"parameters"`
}

// FIXME: This should be string, err
func (s *Service) CredentialString(key string) string {
	if v, ok := s.Credentials[key]; ok {
		return v.(string)
	} else {
		zap.L().Error("cannot find credential for key",
			zap.String("key", key))
		return fmt.Sprintf("NOVAL:%s", v)
	}
}

func (s *Service) CredentialInt(key string) int64 {
	if v, ok := s.Credentials[key]; ok {
		return int64(v.(int))
	} else {
		zap.L().Error("cannot find credential for key",
			zap.String("key", key))
		return -1
	}
}

type Database = Service
type Bucket = Service
type env struct {
	AppEnv        string               `mapstructure:"APPENV"`
	Home          string               `mapstructure:"HOME"`
	MemoryLimit   string               `mapstructure:"MEMORY_LIMIT"`
	Pwd           string               `mapstructure:"PWD"`
	TmpDir        string               `mapstructure:"TMPDIR"`
	User          string               `mapstructure:"USER"`
	EightServices map[string][]Service `mapstructure:"EIGHT_SERVICES"`
	Port          string               `mapstructure:"PORT"`
	AllowedHosts  string               `mapstructure:"SCHEDULE"`
	VcapServices  map[string][]Service
	UserServices  []Service
	ObjectStores  []Bucket
	Databases     []Database
}

type container_env struct {
	VcapServices map[string][]Service `mapstructure:"VCAP_SERVICES"`
}

var container_envs = []string{"DOCKER", "GH_ACTIONS"}
var cf_envs = []string{"SANDBOX", "PREVIEW", "DEV", "STAGING", "PROD"}
var test_envs = []string{"LOCALHOST"}

func InitGlobalEnv(this_service string) {
	Env = &env{}
	configName := "NO_CONFIG_NAME_SET"

	viper.AddConfigPath("/home/vcap/app/config")
	viper.SetConfigType("yaml")

	log.Println("ENV is", os.Getenv("ENV"))
	log.Println("ENV", "LOCAL", IsLocalTestEnv(), "CONT", IsContainerEnv(), "CLOUD", IsCloudEnv())

	if IsContainerEnv() {
		log.Println("IsContainerEnv")
		viper.SetConfigName("container")
		configName = "container"
	}

	if IsLocalTestEnv() {
		log.Println("IsLocalTestEnv")
		viper.AddConfigPath("../../config")
		viper.AddConfigPath("config")
		viper.SetConfigName("localhost")
		configName = "localhost"
	}

	if IsCloudEnv() {
		log.Println("IsCloudEnv")
		viper.SetConfigName("cf")
		configName = "cf"
		// https://github.com/spf13/viper/issues/1706
		// https://github.com/spf13/viper/issues/1671
		viper.AutomaticEnv()
	}

	// Grab the PORT in the cloud and locally from os.Getenv()
	err := viper.BindEnv("PORT")
	if err != nil {
		zap.L().Fatal("ENV could not bind env", zap.String("err", err.Error()))
	}

	//err := viper.ReadInConfig()
	err = viper.ReadConfig(config.GetYamlFileReader(configName + ".yaml"))
	if err != nil {
		log.Fatal("ENV cannot load in the config file ", viper.ConfigFileUsed())
	}

	err = viper.Unmarshal(&Env)

	if err != nil {
		log.Fatal("ENV can't find config files: ", err)
	}

	if err != nil {
		log.Fatal("ENV environment can't be loaded: ", err)
	}

	// CF puts VCAP_* in a string containing JSON.
	// This means we don't have 1:1 locally *yet*, but
	// if we unpack things right, we end up with one struct
	// with everything in the rgiht places.
	if IsContainerEnv() || IsLocalTestEnv() {
		ContainerEnv := container_env{}
		err := viper.Unmarshal(&ContainerEnv)
		if err != nil {
			log.Println("ENV could not unmarshal VCAP_SERVICES to new")
			log.Fatal(err)
		}
		Env.VcapServices = ContainerEnv.VcapServices
	}

	if IsCloudEnv() {
		new_vcs := make(map[string][]Service, 0)
		err := json.Unmarshal([]byte(os.Getenv("VCAP_SERVICES")), &new_vcs)
		if err != nil {
			log.Println("ENV could not unmarshal VCAP_SERVICES to new")
			log.Fatal(err)
		}
		Env.VcapServices = new_vcs
	}

	// Configure the buckets and databases
	Env.ObjectStores = Env.VcapServices["s3"]
	Env.Databases = Env.VcapServices["aws-rds"]
	Env.UserServices = Env.EightServices["user-provided"]

	if IsLocalTestEnv() {
		log.Println(Env.Databases)
	}

	SetupLogging(this_service)
	SetGinReleaseMode(this_service)

	// Grab the schedule
	s, err := Env.GetUserService(this_service)
	if err != nil {
		log.Println("could not get service for ", this_service)
	}
	Env.AllowedHosts = s.GetParamString("allowed_hosts")
	log.Println("Setting Schedule: ", Env.AllowedHosts)

}

// https://stackoverflow.com/questions/3582552/what-is-the-format-for-the-postgresql-connection-string-url
// postgresql://[user[:password]@][netloc][:port][/dbname][?param1=value1&...]
func (e *env) GetDatabaseUrl(name string) (string, error) {
	for _, db := range e.Databases {
		if db.Name == name {
			params := ""
			if IsContainerEnv() || IsLocalTestEnv() {
				params = "?sslmode=disable"
				return fmt.Sprintf("postgresql://%s@%s:%d/%s%s",
					db.CredentialString("username"),
					db.CredentialString("host"),
					db.CredentialInt("port"),
					db.CredentialString("db_name"),
					params,
				), nil
			}
			if IsCloudEnv() {
				return db.CredentialString("uri"), nil
			}

		}
	}
	return "", fmt.Errorf("ENV no db found with name %s", name)
}

func (e *env) GetObjectStore(name string) (Bucket, error) {
	for _, b := range e.ObjectStores {
		if DEBUG_ENV {
			zap.L().Debug("GetObjectStore",
				zap.String("bucket_name", b.Name),
				zap.String("search_key", name),
				zap.Bool("is_equal", b.Name == name),
			)
		}

		if b.Name == name {
			return b, nil
		}
	}
	return Bucket{}, fmt.Errorf("ENV no bucket with name %s", name)
}

func (e *env) GetUserService(name string) (Service, error) {
	for _, s := range e.UserServices {
		if s.Name == name {
			return s, nil
		}
	}
	return Service{}, fmt.Errorf("ENV no service with name %s", name)
}

func IsContainerEnv() bool {
	return slices.Contains(container_envs, os.Getenv("ENV"))
}

func IsLocalTestEnv() bool {
	return slices.Contains(test_envs, os.Getenv("ENV"))
}

func IsCloudEnv() bool {
	return slices.Contains(cf_envs, os.Getenv("ENV"))
}

func (s *Service) GetParamInt64(key string) int64 {
	for _, global_s := range Env.UserServices {
		if s.Name == global_s.Name {
			if global_param_val, ok := global_s.Parameters[key]; ok {
				return int64(global_param_val.(int))
			} else {
				log.Fatalf("ENV no int64 param found for %s", key)
			}
		}
	}
	return -1
}

func (s *Service) GetParamString(key string) string {

	for _, global_s := range Env.UserServices {
		if s.Name == global_s.Name {
			if global_param_val, ok := global_s.Parameters[key]; ok {
				return global_param_val.(string)
			} else {
				log.Fatalf("ENV no string param found for %s", key)
			}
		}
	}

	return fmt.Sprintf("NO VALUE FOUND FOR KEY: [%s, %s]", s.Name, key)
}

func (s *Service) GetParamBool(key string) bool {
	for _, global_s := range Env.UserServices {
		if s.Name == global_s.Name {
			if global_param_val, ok := global_s.Parameters[key]; ok {
				return global_param_val.(bool)
			} else {
				log.Fatalf("ENV no bool param found for %s", key)
			}
		}
	}
	return false
}

func (s *Service) AsJson() string {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	return string(b)
}
