package grifts

import (
  "github.com/gobuffalo/buffalo"
	"github.com/go-saloon/saloon/actions"
)

func init() {
  buffalo.Grifts(actions.App())
}
