//go:build dev
// +build dev

package view

import (
	"net/http"
)

var Assets = http.Dir("internal/view/public")
