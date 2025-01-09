package common

import (
	"testing"

	"github.com/zeebo/assert"
)

var d64 string = `
{
   "com": {
      "Domain64ToFQDN": {
         "0300000100000100": "treasury-testing.alextom.com",
         "0300000100000200": "staging.arc.alextom.com",
         "0300000200000100": "my.goarmy.com",
         "0300000200000200": "www.goarmy.com"
      },
      "Domain64s": [
         "0300000100000100",
         "0300000100000200",
         "0300000200000100",
         "0300000200000200"
      ],
      "FQDNToDomain64": {
         "my.goarmy.com": "0300000200000100",
         "staging.arc.alextom.com": "0300000100000200",
         "treasury-testing.alextom.com": "0300000100000100",
         "www.goarmy.com": "0300000200000200"
      },
      "FQDNs": [
         "treasury-testing.alextom.com",
         "staging.arc.alextom.com",
         "my.goarmy.com",
         "www.goarmy.com"
      ],
      "RFQDNs": [
         "com.alextom.treasury-testing",
         "com.alextom.staging.arc",
         "com.goarmy.my",
         "com.goarmy.www"
      ]
   },
   "edu": {
      "Domain64ToFQDN": {
         "0400000100000100": "www.cdse.edu",
         "0400000200000100": "library.usda.cornell.edu",
         "0400000200000200": "mannlib.agcensus.cornell.edu",
         "0400000300000100": "wmdcenter.ndu.edu",
         "0400000400000100": "georgewbushlibrary.smu.edu",
         "0400000500000100": "bush41library.tamu.edu",
         "0400000600000100": "www.usmcu.edu",
         "0400000700000100": "www.usmma.edu"
      },
      "Domain64s": [
         "0400000100000100",
         "0400000200000100",
         "0400000200000200",
         "0400000300000100",
         "0400000400000100",
         "0400000500000100",
         "0400000600000100",
         "0400000700000100"
      ],
      "FQDNToDomain64": {
         "bush41library.tamu.edu": "0400000500000100",
         "georgewbushlibrary.smu.edu": "0400000400000100",
         "library.usda.cornell.edu": "0400000200000100",
         "mannlib.agcensus.cornell.edu": "0400000200000200",
         "wmdcenter.ndu.edu": "0400000300000100",
         "www.cdse.edu": "0400000100000100",
         "www.usmcu.edu": "0400000600000100",
         "www.usmma.edu": "0400000700000100"
      },
      "FQDNs": [
         "www.cdse.edu",
         "library.usda.cornell.edu",
         "mannlib.agcensus.cornell.edu",
         "wmdcenter.ndu.edu",
         "georgewbushlibrary.smu.edu",
         "bush41library.tamu.edu",
         "www.usmcu.edu",
         "www.usmma.edu"
      ],
      "RFQDNs": [
         "edu.cdse.www",
         "edu.cornell.library.usda",
         "edu.cornell.mannlib.agcensus",
         "edu.ndu.wmdcenter",
         "edu.smu.georgewbushlibrary",
         "edu.tamu.bush41library",
         "edu.usmcu.www",
         "edu.usmma.www"
      ]
   }
}
`

func TestUnmarshal(t *testing.T) {
	b := []byte(d64)
	//nolint:all
	NewTLD64s(b)
}

func TestCheckEdu(t *testing.T) {
	b := []byte(d64)
	//nolint:all
	d, err := NewTLD64s(b)
	if err != nil {
		t.Error(err)
		return
	}
	if _, ok := d["edu"]; !ok {
		t.Error("did not find edu in map")
	}
}

func TestCountEdu(t *testing.T) {
	b := []byte(d64)
	//nolint:all
	d, err := NewTLD64s(b)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, len(d["edu"].Domain64s), 8)
}

func TestConvH2D(t *testing.T) {
	b := []byte(d64)
	d, _ := NewTLD64s(b)
	n := D64HexToDec(d["edu"].Domain64s[0])
	assert.Equal(t, 288230380446679296, n)
}

func TestConvD2H(t *testing.T) {
	n := D64DecToHex(288230380446679296)
	assert.Equal(t, "0400000100000100", n)
}
