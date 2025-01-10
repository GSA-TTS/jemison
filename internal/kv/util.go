package kv

import (
	"context"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/GSA-TTS/jemison/internal/env"
	minio "github.com/minio/minio-go/v7"
	minio_credentials "github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

var s3cache sync.Map

// NewS3FromBucketName creates an S3 object containing bucket information
// from VCAP and a minio client ready to talk to the bucket. S3JSON objects
// carry the information so they can load/save.
//
//nolint:cyclop,funlen
func newS3FromBucketName(bucket_name string) S3 {
	if !env.IsValidBucketName(bucket_name) {
		log.Fatal("KV INVALID BUCKET NAME ", bucket_name)
	}

	if v, ok := s3cache.Load(bucket_name); ok {
		cast, ok := v.(S3)
		if !ok {
			zap.L().Error("could not cast to s3 struct")
		}

		return cast
	}

	s3 := S3{}

	// Grab a reference to our bucket from the config.
	b, err := env.Env.GetObjectStore(bucket_name)
	if err != nil {
		zap.L().Error("could not get bucket from config", zap.String("bucket_name", bucket_name))
		os.Exit(1)
	}

	if DEBUG_S3 {
		zap.L().Debug("got reference to bucket from vcap",
			zap.String("name", b.Name),
			zap.String("bucket", b.CredentialString("bucket")),
			zap.String("region", b.CredentialString("region")))
	}

	s3.Bucket = b

	// Initialize minio client object.
	useSSL := true
	if env.IsContainerEnv() || env.IsLocalTestEnv() {
		// log.Println("ENV disabling SSL in containerized environment")
		useSSL = false
	}

	options := minio.Options{
		Creds: minio_credentials.NewStaticV4(
			b.CredentialString("access_key_id"),
			b.CredentialString("secret_access_key"), ""),
		Secure: useSSL,
	}

	minioClient, err := minio.New(b.CredentialString("endpoint"), &options)
	if err != nil {
		log.Fatalln(err)
	}

	s3.MinioClient = minioClient
	ctx := context.Background()

	found, err := minioClient.BucketExists(ctx, s3.Bucket.CredentialString("bucket"))
	if err != nil {
		zap.L().Fatal("could not check if bucket exists",
			zap.String("bucket_name", bucket_name),
			zap.String("err", err.Error()))
	}

	if found {
		if DEBUG_S3 {
			zap.L().Debug("pre-existing bucket in S3",
				zap.String("bucket_name", bucket_name))
		}
		// Make sure to insert the metadata into the sync.Map
		// when we find a bucket that already exists.
		// buckets.Store(bucket_name, s3)
		s3cache.Store(bucket_name, s3)

		return s3
	}

	if env.IsContainerEnv() {
		log.Println("KV creating new bucket ", bucket_name)
		// Try and make the bucket; if we're local, this is necessary.
		ctx := context.Background()
		err = minioClient.MakeBucket(
			ctx,
			s3.Bucket.CredentialString("bucket"),
			minio.MakeBucketOptions{Region: b.CredentialString("region")})

		if err != nil && !strings.Contains(err.Error(), "succeeded") {
			log.Println(err)
			log.Fatal("KV could not create bucket ", bucket_name)
		}
	} // Skip container creation in CF

	s3cache.Store(bucket_name, s3)

	return s3
}

func containsAll(target string, pieces []string) bool {
	allExist := true
	for _, s := range pieces {
		allExist = allExist && strings.Contains(target, s)
	}

	return allExist
}

const BACKUOFF_MS = 50

const BACKOFF_OFFSET = 25

func store(s3 *S3, destination_key string, size int64, reader io.Reader, mime_type string) error {
	trying := true
	backoff := 50

	for trying {
		ctx := context.Background()
		_, err := s3.MinioClient.PutObject(
			ctx,
			s3.Bucket.CredentialString("bucket"),
			destination_key,
			reader,
			size,
			minio.PutObjectOptions{
				ContentType: mime_type,
				// This seems to set the *minimum* partsize for multipart uploads.
				// Which... makes writing JSON objects impossible.
				// PartSize:    5000000
			},
		)
		// We might be going too fast.
		if err != nil {
			zap.L().Warn("S3JSON could not PUT object",
				zap.String("destination_key", destination_key),
				zap.String("error", err.Error()))

			// Resource requested is unwritable, please reduce your request rate
			if containsAll(err.Error(), []string{"reduce", "rate"}) || containsAll(err.Error(), []string{"not", "store"}) {
				zap.L().Warn("reducing request rate")
				//nolint:gosec
				sleepyTime := time.Duration((rand.IntN(BACKUOFF_MS) + backoff) * int(time.Millisecond))

				backoff += rand.IntN(BACKUOFF_MS) + BACKOFF_OFFSET

				time.Sleep(sleepyTime)

				continue
			} else {
				zap.L().Error("s3 storage error", zap.String("err", err.Error()))

				return err
			}
		} else {
			trying = false
		}
	}

	return nil
}
