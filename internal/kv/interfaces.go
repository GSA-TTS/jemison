package kv

import (
	"encoding/json"
	"log"

	"github.com/GSA-TTS/jemison/internal/util"
)

type JSON map[string]string

type Object interface {
	GetKey() string
	GetJson() JSON
	GetValue(key string) string
	GetSize() int64
	GetMimeType() string
}

type ObjInfo struct {
	Key  string
	Size int64
	Mime string
}

func NewObjInfo(key string, size int64) *ObjInfo {
	return &ObjInfo{
		Key:  key,
		Size: size,
	}
}

type Obj struct {
	value JSON
	info  *ObjInfo
	// bytes []byte
}

func NewObject(key string, value JSON) *Obj {
	b, err := json.Marshal(value)
	if err != nil {
		log.Fatal("ENV could not marshal", key)
	}

	size := int64(len(b))
	mime := ""

	if good, ok := value["content-type"]; !ok {
		mime = "octet/binary"
	} else {
		// Clean the mime type before we instert it.
		mime = util.CleanMimeType(good)
	}

	return &Obj{
		info: &ObjInfo{
			Key:  key,
			Size: size,
			Mime: mime,
		},
		value: value,
	}
}

func (o Obj) GetKey() string {
	return o.info.Key
}

func (o Obj) GetValue(key string) string {
	return o.value[key]
}

func (o Obj) GetJson() JSON {
	return o.value
}

func (o Obj) GetSize() int64 {
	return o.info.Size
}

func (o Obj) GetMimeType() string {
	return o.info.Mime
}
