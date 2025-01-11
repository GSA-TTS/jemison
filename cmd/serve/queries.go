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
	Lhs Q
	Rhs Q
}

func (orq OrQ) String() string {
	return "(" + orq.Lhs.String() + _or + orq.Rhs.String() + ")"
}

func Or(a, b string) Q {
	return OrQ{Lhs: Qs{Str: a}, Rhs: Qs{Str: b}}
}

type AndQ struct {
	Lhs Q
	Rhs Q
}

func And(a, b string) Q {
	return AndQ{Lhs: Qs{Str: a}, Rhs: Qs{Str: b}}
}

func (andq AndQ) String() string {
	return andq.Lhs.String() + _and + andq.Rhs.String()
}

type Query struct {
	Queries []Q
}

func NewQuery() *Query {
	q := Query{}
	q.Queries = make([]Q, 0)

	return &q
}

func (q *Query) AddToQuery(new_q Q) {
	q.Queries = append(q.Queries, new_q)
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
