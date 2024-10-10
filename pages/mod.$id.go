package pages

import (
	"iavlweb/partials"

	"github.com/go-chi/chi/v5"
	"github.com/maddalax/htmgo/framework/h"
)

func ModPage(ctx *h.RequestContext) *h.Page {
	id := chi.URLParam(ctx.Request, "id")
	return h.NewPage(
		RootPage(
			h.Div(
				h.GetPartialWithQs(
					partials.ListKeyValuesPartial,
					h.NewQs("module", id),
					"load",
				),
			),
		),
	)
}
