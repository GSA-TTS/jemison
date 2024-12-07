// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package search_db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Body struct {
	ID     int64  `json:"id"`
	PathID int32  `json:"path_id"`
	Kind   int32  `json:"kind"`
	Tag    string `json:"tag"`
	Body   string `json:"body"`
}

type Header struct {
	ID     int64  `json:"id"`
	PathID int32  `json:"path_id"`
	Kind   int32  `json:"kind"`
	Level  int32  `json:"level"`
	Header string `json:"header"`
}

type Metadatum struct {
	ID    int64       `json:"id"`
	Col   pgtype.Text `json:"col"`
	Value pgtype.Text `json:"value"`
}

type Path struct {
	ID   int64  `json:"id"`
	Host string `json:"host"`
	Path string `json:"path"`
}

type Title struct {
	ID     int64  `json:"id"`
	PathID int32  `json:"path_id"`
	Kind   int32  `json:"kind"`
	Title  string `json:"title"`
}