package kv

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/util"
	minio "github.com/minio/minio-go/v7"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

var DEBUG_S3 = false

// S3 holds a bucket structure (containing VCAP_SERVICES information)
// and an S3 client connection from the min.io libraries.
type S3 struct {
	Bucket      env.Bucket
	MinioClient *minio.Client
}

// /////////////////////////////
// FILE INTERACTIONS

// NewS3 creates a new S3 object for the bucket given.
// Lets us copy files to/from the bucket.
func NewS3(bucket_name string) *S3 {
	s3 := newS3FromBucketName(bucket_name)

	return &s3
}

func (s3 *S3) FileToS3(key *util.Key, local_filename string, mime_type string) error {
	reader, err := os.Open(local_filename)
	if err != nil {
		log.Fatal("FileToS3 cannot open file ", local_filename)
	}

	fi, err := reader.Stat()
	if err != nil {
		log.Println("KV could not stat file")
		log.Fatal(err)
	}

	return store(s3, key.Render(), fi.Size(), reader, mime_type)
}

func (s3 *S3) FileToS3Path(key string, local_filename string, mime_type string) error {
	reader, err := os.Open(local_filename)
	if err != nil {
		log.Fatal("FileToS3Path cannot open file ", local_filename)
	}

	fi, err := reader.Stat()
	if err != nil {
		log.Println("KV could not stat file")
		log.Fatal(err)
	}

	return store(s3, key, fi.Size(), reader, mime_type)
}

func (s3 *S3) S3ToFile(key *util.Key, local_filename string) error {
	ctx := context.Background()

	err := s3.MinioClient.FGetObject(
		ctx,
		s3.Bucket.CredentialString("bucket"),
		key.Render(),
		local_filename,
		minio.GetObjectOptions{})
	if err != nil {
		zap.L().Error("could not FGetObject",
			zap.String("bucket", s3.Bucket.Name),
			zap.String("key", key.Render()),
			zap.String("local_filename", local_filename),
		)

		return err
	}

	return nil
}

func (s3 *S3) S3PathToFile(path string, local_filename string) error {
	ctx := context.Background()

	err := s3.MinioClient.FGetObject(
		ctx,
		s3.Bucket.CredentialString("bucket"),
		path,
		local_filename,
		minio.GetObjectOptions{})
	if err != nil {
		zap.Error(err)

		return err
	}

	return nil
}

func (s3 *S3) S3PathToS3JSON(key *util.Key) (*S3JSON, error) {
	// The object has a channel interface that we have to empty.
	ctx := context.Background()
	object, err := s3.MinioClient.GetObject(
		ctx,
		s3.Bucket.CredentialString("bucket"),
		key.Render(),
		minio.GetObjectOptions{})

	// https://rezakhademix.medium.com/defer-functions-in-golang-common-mistakes-and-best-practices-96eacdb551f0
	defer func(obj *minio.Object) {
		err := obj.Close()
		if err != nil {
			zap.L().Error("deferred close on S3 object encountered error",
				zap.String("key", key.Render()))
		}
	}(object)

	if err != nil {
		zap.L().Error("could not retrieve object",
			zap.String("bucket_name", s3.Bucket.CredentialString("bucket")),
			zap.String("key", key.Render()),
			zap.String("error", err.Error()))

		return nil, err
	}

	if DEBUG_S3 {
		zap.L().Debug("retrieved S3 object", zap.String("key", key.Render()))
	}

	raw, err := io.ReadAll(object)
	if err != nil {
		zap.L().Error("could not read object bytes",
			zap.String("bucket_name", s3.Bucket.CredentialString("bucket")),
			zap.String("key", key.Render()),
			zap.String("error", err.Error()))

		return nil, err
	}

	s3json := NewS3JSON(s3.Bucket.Name)
	s3json.raw = raw
	s3json.Key = key
	current_mime_type := s3json.GetString("content-type")

	updated, err := sjson.SetBytes(s3json.raw, "content-type", util.CleanMimeType(current_mime_type))
	if err != nil {
		zap.L().Error("could not update raw S3JSON")
	} else {
		s3json.raw = updated
	}

	s3json.empty = false

	return s3json, nil
}

// Lists objects in the bucket, returning keys and sizes.
func (s3 *S3) List(prefix string) ([]*ObjInfo, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectCh := s3.MinioClient.ListObjects(
		ctx,
		s3.Bucket.CredentialString("bucket"),
		minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: false,
		})

	objects := make([]*ObjInfo, 0)

	for object := range objectCh {
		if object.Err != nil {
			zap.Error(object.Err)

			return nil, object.Err
		}

		objects = append(objects, NewObjInfo(object.Key, object.Size))
	}

	return objects, nil
}
