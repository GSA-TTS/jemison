//nolint:all
package common

import (
	"log"
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestStats(t *testing.T) {
	s1 := NewBaseStats("foo.com")
	_ = NewBaseStats("foo.com")
	_ = NewBaseStats("bar.com")
	s1.Set("a", 1)

	m := s1.GetAll()
	log.Println(m)
}

func TestPersist(t *testing.T) {
	s1 := NewBaseStats("foo.com")
	_ = NewBaseStats("foo.com")
	_ = NewBaseStats("bar.com")
	s1.Set("b", 1)

	m := s1.GetAll()
	log.Println(m)
}
