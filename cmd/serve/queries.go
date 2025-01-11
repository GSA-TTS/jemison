package main

var _or = " | "

var _and = " & "

type Q interface {
	String() string
}

type Qs struct {
	Str string
}

func (qs Qs) String() string {
	return qs.Str
}

type OrQ struct {
	LHS Q
	RHS Q
}

func (orq OrQ) String() string {
	return "(" + orq.LHS.String() + _or + orq.RHS.String() + ")"
}

func Or(a, b string) Q {
	return OrQ{LHS: Qs{Str: a}, RHS: Qs{Str: b}}
}

type AndQ struct {
	LHS Q
	RHS Q
}

func And(a, b string) Q {
	return AndQ{LHS: Qs{Str: a}, RHS: Qs{Str: b}}
}

func (andq AndQ) String() string {
	return andq.LHS.String() + _and + andq.RHS.String()
}

type Query struct {
	Queries []Q
}

func NewQuery() *Query {
	q := Query{}
	q.Queries = make([]Q, 0)

	return &q
}

func (q *Query) AddToQuery(newQ Q) {
	q.Queries = append(q.Queries, newQ)
}

func (q *Query) ToString() string {
	qs := ""

	for ndx, current := range q.Queries {
		qs = qs + current.String()
		if ndx != len(q.Queries)-1 {
			qs += _and
		}
	}

	return qs
}
