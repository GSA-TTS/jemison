//nolint:godox
package vcap

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/tidwall/gjson"
)

type VcapServices struct {
	Source    string
	VCAP      gjson.Result
	Raw       string
	Buckets   []Bucket
	Databases []Database
}

func VcapServicesFromEnv(env_var string) VcapServices {
	vcs := VcapServices{}
	vcs.EnvStringToJson(env_var)
	vcs.ParseBuckets()
	vcs.ParseDatabases()

	return vcs
}

func (vcs *VcapServices) EnvStringToJson(env_var string) {
	// Read it in from the VCAP_SERVICES env var,
	// which will provide a large JSON structure.
	vcs.Source = "env"
	vcs.VCAP = gjson.Parse(os.Getenv(env_var))
	vcs.Raw = os.Getenv(env_var)
}

type Bucket struct {
	ServiceName     string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	// FIXME: Check the endpoint shape, and set it
	// URI             string
	// Endpoint        string
}

func (vcs *VcapServices) ParseBuckets() {
	buckets := make([]Bucket, 0)
	// Get each of the values under the S3 tag.
	for _, b := range vcs.VCAP.Get("s3").Array() {
		buckets = append(buckets, Bucket{
			ServiceName:     b.Get("name").String(),
			Region:          b.Get("credentials.region").String(),
			AccessKeyID:     b.Get("credentials.access_key_id").String(),
			SecretAccessKey: b.Get("credentials.secret_access_key").String(),
			// FIXME: Check the endpoint shape, and set it
			// URI:             b.Get("credentials.uri").String(),
			// Endpoint:        b.Get("credentials.endpoint").String(),
		})
	}

	vcs.Buckets = buckets
}

type Database struct {
	ServiceName  string
	InstanceName string
	Username     string
	Host         string
	Name         string
	Password     string
	URI          string
	Endpoint     string
}

func (vcs *VcapServices) ParseDatabases() {
	databases := make([]Database, 0)
	// Get each of the values under the S3 tag.
	for _, db := range vcs.VCAP.Get("aws-rds").Array() {
		databases = append(databases, Database{
			ServiceName:  db.Get("name").String(),
			InstanceName: db.Get("instance_name").String(),
			Username:     db.Get("credentials.username").String(),
			Name:         db.Get("credentials.db_name").String(),
			Host:         db.Get("credentials.host").String(),
			Password:     db.Get("credentials.password").String(),
			URI:          db.Get("credentials.uri").String(),
			Endpoint:     db.Get("credentials.endpoint").String(),
		})
	}

	vcs.Databases = databases
}

func (vcs *VcapServices) GetBucketByName(bucket_name string) *Bucket {
	for _, b := range vcs.Buckets {
		if b.ServiceName == bucket_name {
			return &b
		}
	}

	return nil
}

func (vcs *VcapServices) ToS3Config(service_name string) *aws.Config {
	cfg := aws.Config{}
	creds := aws.Credentials{}

	for _, b := range vcs.Buckets {
		if b.ServiceName == service_name {
			cfg.Region = b.Region
			creds.AccessKeyID = b.AccessKeyID
			creds.SecretAccessKey = b.SecretAccessKey
			credsProvider := credentials.StaticCredentialsProvider{
				Value: creds,
			}
			cfg.Credentials = credsProvider
		}
	}

	return &cfg
}
