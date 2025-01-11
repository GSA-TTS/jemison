//nolint:testpackage
package vcap

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testVCAP = `{
    "s3": [
    {
            "label": "s3",
            "provider": "alpha-provider",
            "plan": "basic",
            "name": "backups",
            "tags": [
                "AWS",
                "S3",
                "object-storage"
            ],
            "instance_guid": "UUIDALPHA1",
            "instance_name": "backups",
            "binding_guid": "UUIDALPHA2",
            "binding_name": null,
            "credentials": {
                "uri": "s3://ACCESSKEYIDALPHA:SECRETACCESSKEYALPHA@s3-us-gov-alpha-1.amazonaws.com/BROKEREDBUCKETALPHA",
                "insecure_skip_verify": false,
                "access_key_id": "ACCESSKEYIDALPHA",
                "secret_access_key": "SECRETACCESSKEY+ALPHA",
                "region": "us-gov-west-1",
                "bucket": "BROKEREDBUCKETALPHA",
                "endpoint": "s3-us-gov-alpha-1.amazonaws.com",
                "fips_endpoint": "s3-fips.us-gov-alpha-1.amazonaws.com",
                "additional_buckets": []
            },
            "syslog_drain_url": "https://ALPHA.drain.url",
            "volume_mounts": ["no_mounts"]
        }
    ],
    "user-provided": [
        {
            "label": "mc",
            "name": "backups",
            "tags": [],
            "instance_guid": "UUIDALPHA1",
            "instance_name": "backups",
            "binding_guid": "UUIDALPHA2",
            "binding_name": null,
            "credentials": {
                "access_key_id": "longtest",
                "secret_access_key": "longtest",
                "bucket": "gsa-fac-private-s3",
                "endpoint": "localhost",
                "admin_username": "minioadmin",
                "admin_password": "minioadmin"
            }
        }
    ],
    "aws-rds": [
        {
            "label": "aws-rds",
            "provider": null,
            "plan": "medium-gp-psql",
            "name": "fac-db",
            "tags": [
                "database",
                "RDS"
            ],
            "instance_guid": "source-guid",
            "instance_name": "fac-db",
            "binding_guid": "source-binding-guid",
            "binding_name": null,
            "credentials": {
                "db_name": "the-source-db-name",
                "host": "the-source-db.us-gov-west-1.rds.amazonaws.com",
                "name": "the-source-name",
                "password": "the-source-password",
                "port": "54321",
                "uri": "the-source-uri",
                "username": "source-username"
            },
            "syslog_drain_url": null,
            "volume_mounts": []
        },
        {
            "label": "aws-rds",
            "provider": null,
            "plan": "medium-gp-psql",
            "name": "fac-snapshot-db",
            "tags": [
                "database",
                "RDS"
            ],
            "instance_guid": "dest-instance-guid",
            "instance_name": "fac-snapshot-db",
            "binding_guid": "dest-binding-guid",
            "binding_name": null,
            "credentials": {
                "db_name": "the-dest-db-name",
                "host": "the-dest-db.us-gov-west-1.rds.amazonaws.com",
                "name": "the-dest-name",
                "password": "the-dest-password",
                "port": "65432",
                "uri": "the-dest-uri",
                "username": "dest-username"
            },
            "syslog_drain_url": null,
            "volume_mounts": []
        }
    ]
}`

func TestReadEnv(t *testing.T) {
	os.Setenv("VCAP_SERVICES", testVCAP)

	vcs := VcapServicesFromEnv("VCAP_SERVICES")

	// Expected, actual
	assert.Equal(t, 1, len(vcs.Buckets))
	assert.Equal(t, 2, len(vcs.Databases))
}

func TestDatbases(t *testing.T) {
	os.Setenv("VCAP_SERVICES", testVCAP)

	vcs := VcapServicesFromEnv("VCAP_SERVICES")

	assert.Equal(t, "fac-db", vcs.Databases[0].ServiceName)
	assert.Equal(t, "fac-snapshot-db", vcs.Databases[1].ServiceName)
	assert.Equal(t, "the-dest-db-name", vcs.Databases[1].Name)
	assert.Equal(t, "the-dest-password", vcs.Databases[1].Password)
}
