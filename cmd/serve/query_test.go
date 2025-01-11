//nolint:paralleltest
package main

import (
	"testing"

	"github.com/zeebo/assert"
)

func TestOr(t *testing.T) {
	q := NewQuery()
	q.AddToQuery(Or("alice", "bob"))
	assert.Equal(t, "alice OR bob", q.ToString())
}

func TestAnd(t *testing.T) {
	q := NewQuery()
	q.AddToQuery(And("alice", "bob"))
	assert.Equal(t, "alice AND bob", q.ToString())
}

func TestCompound(t *testing.T) {
	q := NewQuery()
	q.AddToQuery(Or("alice", "bob"))
	q.AddToQuery(Or("clarice", "dylan"))
	assert.Equal(t, "alice OR bob AND clarice OR dylan", q.ToString())
}
