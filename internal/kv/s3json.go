// kv provides an interface to key/value work in S3
// It is specialized to the `jemison` architecture.
//
//nolint:godox,godot
package kv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/GSA-TTS/jemison/internal/util"
	minio "github.com/minio/minio-go/v7"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

var DEBUG_S3JSON = false

// NewFromBytes(bucket_name string, host string, path string, m []byte) *S3JSON
// NewEmptyS3JSON(bucket_name string, host string, path string) *S3JSON
// (s3json *S3JSON) IsEmpty() bool
// (s3json *S3JSON) Save() error
// (s3json *S3JSON) Load() error

// Only open any given bucket once.
// FIXME: get rid of these long-lived globals.
// Open and close things, until it becomes a performance concern.
// It would be safer if...
// Load() does an open and a close
// Save() does an open and a close
// Then, every object is self-contained. Slower, but self-contained.
// The sync... is hell waiting to happen in terms of debugging.

// S3JSON structs are JSON documents stored in S3.
// This is because `jemison` shuttles JSON documents in-and-out of S3, and
// we want to be able to find a document representing a host/path in
// multiple, different buckets.
type S3JSON struct {
	Key   *util.Key
	raw   []byte
	S3    S3
	empty bool
}

func NewS3JSON(bucket_name string) *S3JSON {
	s3 := newS3FromBucketName(bucket_name)

	return &S3JSON{
		Key:   &util.Key{},
		raw:   nil,
		S3:    s3,
		empty: true,
	}
}

// NewFromBytes takes a []byte representation of a JSON document and constructs
// a S3JSON document from it.
// Inserts _key
func NewFromBytes(bucket_name string, scheme util.Scheme, host string, path string, m []byte) *S3JSON {
	s3 := newS3FromBucketName(bucket_name)
	key := util.CreateS3Key(scheme, host, path, util.JSON)
	w_key, _ := sjson.SetBytes(m, "_key", key.Render())

	return &S3JSON{
		Key:   key,
		raw:   w_key,
		S3:    s3,
		empty: false,
	}
}

// Inserts _key
func NewFromMap(bucket_name string, scheme util.Scheme, host string, path string, m map[string]string) *S3JSON {
	s3 := newS3FromBucketName(bucket_name)
	key := util.CreateS3Key(scheme, host, path, util.JSON)
	m["_key"] = key.Render()

	b, _ := json.Marshal(m)

	return &S3JSON{
		Key:   key,
		raw:   b,
		S3:    s3,
		empty: false,
	}
}

// Creates a new, empty S3JSON struct, setting it as `empty`.
// `Load()` must be called on it before we can use it.
func NewEmptyS3JSON(bucket_name string, scheme util.Scheme, host string, path string) *S3JSON {
	s3 := newS3FromBucketName(bucket_name)
	key := util.CreateS3Key(scheme, host, path, util.JSON)

	return &S3JSON{
		Key:   key,
		raw:   nil,
		S3:    s3,
		empty: true,
	}
}

// IsEmpty() Checks if the S3JSON struct is empty.
// Should be `true` before a call to `Load()`, `false` after.
func (s3json *S3JSON) IsEmpty() bool {
	return s3json.empty
}

func (s3json *S3JSON) URL() *url.URL {
	return &url.URL{
		Scheme: s3json.Key.Scheme.String(),
		Host:   s3json.Key.Host,
		Path:   s3json.Key.Path,
	}
}

// Save() will do a `Put` of the JSON to S3.
// BUG(jadudm): handle errors in store gracefully
func (s3json *S3JSON) Save() error {
	if s3json.IsEmpty() {
		return fmt.Errorf("cannot save invalid S3JSON object bucket[%s] host[%s] path[%s]",
			s3json.S3.Bucket.Name, s3json.Key.Host, s3json.Key.Path)
	}

	r := bytes.NewReader(s3json.raw)
	size := int64(len(s3json.raw))

	err := store(&s3json.S3, s3json.Key.Render(), size, r, util.JSON.String())
	if err != nil {
		zap.L().Fatal("could not store S3JSON",
			zap.String("bucket_name", s3json.S3.Bucket.Name),
			zap.String("key", s3json.Key.Render()),
			zap.String("err", err.Error()))

		return err
	}

	return nil
}

// Load() uses the bucket/path information in the underlying S3 struct
// to do a `Get` against S3 and retrieve the JSON document.
func (s3json *S3JSON) Load() error {
	if !s3json.IsEmpty() {
		return fmt.Errorf("will only load empty object bucket[%s] host[%s] path[%s]",
			s3json.S3.Bucket.Name, s3json.Key.Host, s3json.Key.Path)
	}

	key := s3json.Key.Render()
	// The object has a channel interface that we have to empty.
	ctx := context.Background()
	object, err := s3json.S3.MinioClient.GetObject(
		ctx,
		s3json.S3.Bucket.CredentialString("bucket"),
		key,
		minio.GetObjectOptions{})
	// https://rezakhademix.medium.com/defer-functions-in-golang-common-mistakes-and-best-practices-96eacdb551f0
	defer func(obj *minio.Object) {
		err := obj.Close()
		if err != nil {
			zap.L().Error("deferred close on S3 object encountered error",
				zap.String("key", key))
		}
	}(object)

	if err != nil {
		zap.L().Error("could not retrieve object",
			zap.String("bucket_name", s3json.S3.Bucket.CredentialString("bucket")),
			zap.String("key", key),
			zap.String("error", err.Error()))

		return err
	}

	if DEBUG_S3JSON {
		zap.L().Debug("retrieved S3 object", zap.String("key", key))
	}

	raw, err := io.ReadAll(object)
	if err != nil {
		zap.L().Error("could not read object bytes",
			zap.String("bucket_name", s3json.S3.Bucket.CredentialString("bucket")),
			zap.String("key", key),
			zap.String("error", err.Error()))

		return err
	}

	s3json.raw = raw
	current_mime_type := s3json.GetString("content-type")

	updated, err := sjson.SetBytes(s3json.raw, "content-type", util.CleanMimeType(current_mime_type))
	if err != nil {
		zap.L().Error("could not update s3json.raw")
	} else {
		s3json.raw = updated
	}

	s3json.empty = false

	return nil
}

func (s3json *S3JSON) GetJSON() []byte {
	return s3json.raw
}

func (s3json *S3JSON) GetString(gjson_path string) string {
	r := gjson.GetBytes(s3json.raw, gjson_path)

	return r.String()
}

func (s3json *S3JSON) GetInt64(gjson_path string) int64 {
	r := gjson.GetBytes(s3json.raw, gjson_path)

	return int64(r.Int())
}

func (s3json *S3JSON) GetBool(gjson_path string) bool {
	r := gjson.GetBytes(s3json.raw, gjson_path)

	return r.Bool()
}

func (s3json *S3JSON) Set(sjson_path string, value string) {
	b, err := sjson.SetBytes(s3json.raw, sjson_path, value)
	if err != nil {
		zap.L().Error("could not set JSON path in Set()",
			zap.String("sjson_path", sjson_path),
			zap.String("value", value))
	}

	s3json.raw = b
}

func (s3json *S3JSON) Size() int64 {
	return int64(len(s3json.raw))
}

// type Storage interface {
// 	Store(string, JSON) error
// 	List(string) ([]*ObjInfo, error)
// 	Get(string) (Object, error)
// }

// func (s3 *S3) StoreFile(destination_key string, source_filename string) error {
// 	reader, err := os.Open(source_filename)
// 	if err != nil {
// 		log.Fatal("KV cannot open file", source_filename)
// 	}
// 	fi, err := reader.Stat()
// 	if err != nil {
// 		log.Println("KV could not stat file")
// 		log.Fatal(err)
// 	}

// 	return store(s3, destination_key, fi.Size(), make(JSON, 0), reader)
// }

// ////////////////////////////
// // SUPPORT

// func mapToReader(json_map JSON) (io.Reader, int64) {
// 	b, _ := json.Marshal(json_map)
// 	r := bytes.NewReader(b)
// 	return r, int64(len(b))
// }
