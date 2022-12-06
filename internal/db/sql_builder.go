package db

import (
	"fmt"
	"strings"
)

const (
	logicOpAnd = "and"
	logicOpOr  = "or"
)

type whereExpression struct {
	op        string
	children  []fmt.Stringer
	statement string
}

func (c *whereExpression) String() string {
	if len(c.children) == 0 {
		if c.statement == "" {
			panic("empty cond expr")
		}
		return c.statement
	}
	var stmt []string
	for _, child := range c.children {
		stmt = append(stmt, child.String())
	}
	spacedOp := fmt.Sprintf(" %s ", c.op)
	return "(" + strings.Join(stmt, spacedOp) + ")"
}

func (c *whereExpression) AddChild(s fmt.Stringer) *whereExpression {
	c.children = append(c.children, s)
	return c
}

func (c *whereExpression) AddString(s string) {
	c.children = append(c.children, &whereExpression{statement: s})
}

func newWhereExpression(op string) *whereExpression {
	return &whereExpression{op: op}
}

func newRawWhereExpression(s string) *whereExpression {
	return &whereExpression{statement: s}
}

// newTextSearchWhereExpression: query should be placeholder, or string wrap with ' like 'query'
func newTextSearchWhereExpression(tokenizer, query, colName string) *whereExpression {
	s := fmt.Sprintf("to_tsvector('%s', %s) @@ websearch_to_tsquery('%s', %s)",
		tokenizer, colName,
		tokenizer, query)
	return &whereExpression{statement: s}
}

func newSQLBuilder() *sqlBuilder {
	return &sqlBuilder{
		statements: []string{},
		orders:     []string{},
	}
}

type sqlBuilder struct {
	statements     []string
	defaultLogicOp string // and or
	orders         []string
	where          *whereExpression
}

func (b *sqlBuilder) ToSQL() string {
	final := []string{strings.Join(b.statements, " ")}
	if b.where != nil {
		final = append(final, "where", b.where.String())
	}

	final = append(final, b.orders...)
	return strings.Join(final, " ")
}

func (b *sqlBuilder) Statement(s string) *sqlBuilder {
	b.statements = append(b.statements, s)
	return b
}

func (b *sqlBuilder) Statementf(format string, a ...interface{}) *sqlBuilder {
	b.statements = append(b.statements, fmt.Sprintf(format, a...))
	return b
}

func (b *sqlBuilder) SetWhere(cond *whereExpression) *sqlBuilder {
	b.where = cond
	return b
}

func (b *sqlBuilder) Where(op string, cond *whereExpression) *sqlBuilder {
	parent := b.where
	if parent == nil {
		parent = &whereExpression{op: op}
	}

	if parent.op != op {
		parent.AddChild(&whereExpression{op: op, children: []fmt.Stringer{cond}})
	} else {
		parent.AddChild(cond)
	}

	if b.where == nil {
		b.where = parent
	}
	return b
}

func (b *sqlBuilder) WhereRaw(op string, format string, a ...interface{}) *sqlBuilder {
	s := fmt.Sprintf(format, a...)
	return b.Where(op, newRawWhereExpression(s))
}

func (b *sqlBuilder) Order(s string) *sqlBuilder {
	b.orders = append(b.orders, s)
	return b
}
