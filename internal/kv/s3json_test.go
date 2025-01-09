package kv

import (
	"log"
	"os"
	"testing"

	"github.com/GSA-TTS/jemison/internal/env"
	"github.com/GSA-TTS/jemison/internal/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Tests need a backend
// docker compose -f backend.yaml up

func setup( /* t *testing.T */ ) func(t *testing.T) {
	os.Setenv("ENV", "LOCALHOST")
	env.InitGlobalEnv("testing_env") // we need to pass something
	return func(t *testing.T) {
		// t.Log("teardown test case")
	}
}

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestKv(t *testing.T) {
	setup()
	log.Println(env.Env.ObjectStores)
}

func TestEmpty(t *testing.T) {
	setup()
	s3json := NewEmptyS3JSON("fetch", util.HTTPS, "search.gov", "/")
	assert.Equal(t, "fetch", s3json.S3.Bucket.Name)
}

//func NewFromBytes(bucket_name string, scheme util.Scheme, host string, path string, m []byte) *S3JSON {

func TestNewFromBytes(t *testing.T) {
	setup()
	s3json := NewFromBytes("fetch", util.HTTPS, "search.gov", "/", []byte(`{"a": 3, "b": 5}`))
	assert.Equal(t, "fetch", s3json.S3.Bucket.Name)
}

func TestGetFromBytes(t *testing.T) {
	setup()
	s3json := NewFromBytes("fetch", util.HTTPS, "search.gov", "/", []byte(`{"a": 3, "b": 5}`))
	assert.Equal(t, int64(3), s3json.GetInt64("a"))
}

func TestSave(t *testing.T) {
	setup()
	s3json := NewFromBytes("fetch", util.HTTPS, "search.gov", "/", []byte(`{"a": 3, "b": 5}`))
	//nolint:all
	s3json.Save()
	assert.Equal(t, int64(3), s3json.GetInt64("a"))
}

func TestLoad(t *testing.T) {
	setup()
	s3json := NewEmptyS3JSON("fetch", util.HTTPS, "search.gov", "/")
	err := s3json.Load()
	if err != nil {
		zap.L().Error("TestLoad", zap.String("error", err.Error()))
	}
	zap.L().Info("TestLoad", zap.ByteString("raw", s3json.raw))
	assert.Equal(t, int64(3), s3json.GetInt64("a"))
	assert.Equal(t, int64(5), s3json.GetInt64("b"))
	// This will return *something*, always. Which can be dangerous.
	assert.NotEqual(t, int64(8), s3json.GetInt64("c"))
	assert.Equal(t, int64(0), s3json.GetInt64("c"))
}

// func TestServices(t *testing.T) {
// 	setup()
// 	kv := NewKV("extract")
// 	assert.Equal(t, kv.Bucket.Name, "extract")
// 	s, err := env.Env.GetUserService("extract")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	log.Println(s)
// 	assert.Equal(t, "extract", s.Name)
// }

// func TestParams1(t *testing.T) {
// 	setup()
// 	kv := NewKV("extract")
// 	assert.Equal(t, kv.Bucket.Name, "extract")
// 	s, err := env.Env.GetUserService("extract")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	log.Println(s)
// 	assert.Equal(t, s.GetParamBool("extract_html"), true)
// }

// func TestParams2(t *testing.T) {
// 	setup()
// 	kv := NewKV("serve")
// 	assert.Equal(t, kv.Bucket.Name, "serve")
// 	s, err := env.Env.GetUserService("serve")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	log.Println(s)
// 	assert.Equal(t, "../../assets/databases", s.GetParamString("database_files_path"))
// }

// func TestParams3(t *testing.T) {
// 	setup()
// 	kv := NewKV("serve")
// 	assert.Equal(t, kv.Bucket.Name, "serve")
// 	s, err := env.Env.GetUserService("serve")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	log.Println(s)
// 	assert.Equal(t, int64(10004), s.GetParamInt64("external_port"))
// }
