package common

import "net/http"

type BalanceArgs struct {
	Key  string
	Size int64
}

func (BalanceArgs) Kind() string {
	return "balance"
}

type SpaceAvailArgs struct {
	Host     string `json:"host"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

func (SpaceAvailArgs) Kind() string {
	return "balance"
}

type EntreeArgs struct {
	Scheme    string `json:"scheme"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	FullCrawl bool   `json:"fullCrawl"`
	HallPass  bool   `json:"hallpass"`
}

func (EntreeArgs) Kind() string {
	return "entree"
}

type ExtractArgs struct {
	Scheme      string `json:"scheme"`
	Host        string `json:"host"`
	Path        string `json:"path"`
	GuestbookID int64  `json:"gb_id"`
}

func (ExtractArgs) Kind() string {
	return "extract"
}

type FetchArgs struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
}

func (FetchArgs) Kind() string {
	return "fetch"
}

type PackArgs struct {
	Scheme      string `json:"scheme"`
	Host        string `json:"host"`
	Path        string `json:"path"`
	GuestbookID int64  `json:"gb_id"`
}

func (PackArgs) Kind() string {
	return "pack"
}

type ServeArgs struct {
	Filename string `json:"filename"`
}

func (ServeArgs) Kind() string {
	return "serve"
}

type WalkArgs struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
}

func (WalkArgs) Kind() string {
	return "walk"
}

type HTTPResponse func(w http.ResponseWriter, r *http.Request)

// VALIDATOR TYPES.
var ValidateFetchQueue = "validate_fetch"

type ValidateFetchArgs struct {
	Fetch FetchArgs
}

func (ValidateFetchArgs) Kind() string {
	return ValidateFetchQueue
}
