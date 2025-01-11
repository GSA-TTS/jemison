//nolint:testpackage,lll
package filter

import (
	"log"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test URLs should be *UNENCODED*. We will parse
// them with url.Parse(), which will handle the encoding.
var tests = []struct {
	url      string
	fun      func(*url.URL) error
	expected bool
}{
	{"http://insecure.gov/", insecureGov, true},
	{"", isTooShort(1), true},
	{"/", isTooShort(1), false},
	{"https://tooLong.gov/", exceedsLength(5), true},
	{"https://tooLong.gov/", exceedsLength(200), false},
	{"https://nasa.gov/", hasSlashHttp, false},
	{"https://nasa.gov/something/http://", hasSlashHttp, true},
	// Spaces become %20 once encoded
	{"https://blog1a.nasa.gov/right here", hasRightHere, true},
	{"https://blog1b.nasa.gov/right here", IsReject, true},
	{"https://blog2.nasa.gov/right here", IsReject, true},
	{"http://blog3.nasa.gov/right here", IsReject, true},
	{"https://short.gov/", IsReject, false},
	{"https://dot.gov/file.docx", hasSkippableExtension, true},
	{"https://dot.gov/file.jpg", hasSkippableExtension, true},
	{"https://dot.gov/file.png", hasSkippableExtension, true},
	{"https://dot.gov/file.xlsx", hasSkippableExtension, true},
	{"#file", hasSkippablePrefixRelative, true},
	{"https://blog1a.nasa.gov/right here/category/1/right here/category/2/right here/category/3/right here/category/4/right here", IsReject, true},
	// Becomes %5C once encoded
	{`https://foo.gov/this.pdf\`, endsWithWrongSlash, true},
	{`https://bar.gov/this.pdf\`, IsReject, true},
}

func TestAll(t *testing.T) {
	for _, tt := range tests {
		u, err := url.Parse(tt.url)
		if err != nil {
			t.Error(err)
		}

		if tt.expected {
			log.Println(tt.url)
			assert.Error(t, tt.fun(u))
		} else {
			assert.Equal(t, nil, tt.fun(u))
		}
	}
}
