package db

import (
	"fmt"
	"strings"
)

const (
	logicOpAnd = "and"
	// logicOpOr  = "or"
)

func newSQLBuilder(defaultLogicOp string) *sqlBuilder {
	return &sqlBuilder{
		statements:     []string{},
		defaultLogicOp: defaultLogicOp,
		filters:        []string{},
		orders:         []string{},
	}
}

type sqlBuilder struct {
	statements     []string
	defaultLogicOp string   // and or
	filters        []string // where xx = y
	orders         []string
}

func (b *sqlBuilder) ToSQL() string {
	filters := strings.Join(b.filters, fmt.Sprintf(" %s ", b.defaultLogicOp))
	final := []string{strings.Join(b.statements, " ")}
	if filters != "" {
		final = append(final, "where", filters)
	}

	final = append(final, b.orders...)
	return strings.Join(final, " ")
}

func (b *sqlBuilder) Statement(s string) *sqlBuilder {
	b.statements = append(b.statements, s)
	return b
}

func (b *sqlBuilder) Filter(s string) *sqlBuilder {
	b.filters = append(b.filters, s)
	return b
}

func (b *sqlBuilder) Order(s string) *sqlBuilder {
	b.orders = append(b.orders, s)
	return b
}
