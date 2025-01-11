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
func newS3FromBucketName(bucketName string) S3 {
	if !env.IsValidBucketName(bucketName) {
		log.Fatal("KV INVALID BUCKET NAME ", bucketName)
	}

	if v, ok := s3cache.Load(bucketName); ok {
		cast, ok := v.(S3)
		if !ok {
			zap.L().Error("could not cast to s3 struct")
		}

		return cast
	}

	s3 := S3{}

	// Grab a reference to our bucket from the config.
	b, err := env.Env.GetObjectStore(bucketName)
	if err != nil {
		zap.L().Error("could not get bucket from config", zap.String("bucket_name", bucketName))
		os.Exit(1)
	}

	if DebugS3 {
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
			zap.String("bucket_name", bucketName),
			zap.String("err", err.Error()))
	}

	if found {
		if DebugS3 {
			zap.L().Debug("pre-existing bucket in S3",
				zap.String("bucket_name", bucketName))
		}
		// Make sure to insert the metadata into the sync.Map
		// when we find a bucket that already exists.
		// buckets.Store(bucket_name, s3)
		s3cache.Store(bucketName, s3)

		return s3
	}

	if env.IsContainerEnv() {
		log.Println("KV creating new bucket ", bucketName)
		// Try and make the bucket; if we're local, this is necessary.
		ctx := context.Background()
		err = minioClient.MakeBucket(
			ctx,
			s3.Bucket.CredentialString("bucket"),
			minio.MakeBucketOptions{Region: b.CredentialString("region")})

		if err != nil && !strings.Contains(err.Error(), "succeeded") {
			log.Println(err)
			log.Fatal("KV could not create bucket ", bucketName)
		}
	} // Skip container creation in CF

	s3cache.Store(bucketName, s3)

	return s3
}

func containsAll(target string, pieces []string) bool {
	allExist := true
	for _, s := range pieces {
		allExist = allExist && strings.Contains(target, s)
	}

	return allExist
}

const BackoffMillis = 50

const BackoffOffset = 25

func store(s3 *S3, destinationKey string, size int64, reader io.Reader, mimeType string) error {
	trying := true
	backoff := 50

	for trying {
		ctx := context.Background()
		_, err := s3.MinioClient.PutObject(
			ctx,
			s3.Bucket.CredentialString("bucket"),
			destinationKey,
			reader,
			size,
			minio.PutObjectOptions{
				ContentType: mimeType,
				// This seems to set the *minimum* partsize for multipart uploads.
				// Which... makes writing JSON objects impossible.
				// PartSize:    5000000
			},
		)
		// We might be going too fast.
		if err != nil {
			zap.L().Warn("S3JSON could not PUT object",
				zap.String("destination_key", destinationKey),
				zap.String("error", err.Error()))

			// Resource requested is unwritable, please reduce your request rate
			if containsAll(err.Error(), []string{"reduce", "rate"}) || containsAll(err.Error(), []string{"not", "store"}) {
				zap.L().Warn("reducing request rate")
				//nolint:gosec
				sleepyTime := time.Duration((rand.IntN(BackoffMillis) + backoff) * int(time.Millisecond))

				//nolint:gosec
				backoff += rand.IntN(BackoffMillis) + BackoffOffset

				time.Sleep(sleepyTime)

				continue
			}

			zap.L().Error("s3 storage error", zap.String("err", err.Error()))

			//nolint:wrapcheck
			return err
		}

		trying = false
	}

	return nil
}
