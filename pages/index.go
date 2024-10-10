package pages

import (
	"iavlweb/partials"

	"github.com/maddalax/htmgo/framework/h"
)

func IndexPage(ctx *h.RequestContext) *h.Page {
	return h.NewPage(
		RootPage(
			h.Div(
				partials.ModulesListPartial(ctx),
			),
		),
	)
}
