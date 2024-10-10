package partials

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"iavlweb/internal/iavl"

	"github.com/maddalax/htmgo/framework/h"
	"github.com/maddalax/htmgo/framework/service"
)

func ModulesListPartial(ctx *h.RequestContext) *h.Partial {
	iavlPair := service.Get[iavl.DBPair](ctx.ServiceLocator())

	modulesLeft := iavl.ListModules(iavlPair.Left)
	modulesRight := iavl.ListModules(iavlPair.Left)

	var modulesInCommon []string
	for _, m := range modulesLeft {
		for _, m2 := range modulesRight {
			if m == m2 {
				modulesInCommon = append(modulesInCommon, m)
			}
		}
	}

	statsLeft, statsRight := iavlPair.Stats()

	return h.NewPartial(
		h.Fragment(
			h.Div(
				h.Div(
					h.P(h.TextF("Left database has %d keys", statsLeft.Count)),
					h.P(h.TextF("Right database has %d keys", statsRight.Count)),
				),
			),
			h.P(h.Text("Select a module to compare:")),
			h.Ul(
				h.List(
					modulesInCommon,
					func(module string, i int) *h.Element {
						return h.Li(h.A(
							h.Class("text-blue-800 underline"),
							h.Href("/mod/"+module),
							h.Text(module),
						))
					},
				),
			),
		),
	)
}

func ListKeyValuesPartial(ctx *h.RequestContext) *h.Partial {
	module := ctx.QueryParam("module")
	dbPair := service.Get[iavl.DBPair](ctx.ServiceLocator())

	iavlPair, err := dbPair.LoadTrees(0, []byte(fmt.Sprintf("s/k:%s/", module)))
	if err != nil {
		return h.NewPartial(ErrorMessage(err))
	}

	v, err := iavlPair.LoadHigherCommonVersion()
	if err != nil {
		return h.NewPartial(ErrorMessage(err))
	}

	hashLeft, hashRight := iavlPair.Hash()

	if bytes.Equal(hashLeft, hashRight) {
		return h.NewPartial(
			h.H1(
				h.Class("font-bold bg-green-500 p-1 rounded"),
				h.Text(fmt.Sprintf("No differences found in %s", module)),
			),
		)
	}

	var keysDifferentValues [][]byte
	var keysOnlyLeft [][]byte
	iavlPair.Left.Iterate(func(key []byte, value []byte) bool { //nolint:errcheck
		valueRight, err := iavlPair.Right.Get(key)
		if err != nil {
			panic(err)
		}

		if valueRight == nil {
			keysOnlyLeft = append(keysOnlyLeft, key)
			return false
		}

		if !bytes.Equal(value, valueRight) {
			keysDifferentValues = append(keysDifferentValues, key)
		}

		return false
	})

	var keysOnlyRight [][]byte
	iavlPair.Right.Iterate(func(key []byte, value []byte) bool { //nolint:errcheck
		found, err := iavlPair.Left.Has(key)
		if err != nil {
			panic(err)
		}

		if !found {
			keysOnlyRight = append(keysOnlyRight, key)
		}

		return false
	})

	if len(keysDifferentValues) == 0 && len(keysOnlyLeft) == 0 && len(keysOnlyRight) == 0 {
		return h.NewPartial(ErrorMessage(fmt.Errorf("Keys and values are the same, but hash differs. This can happen when the same set/delete operations were made the IAVL tree but in a different order.")))
	}

	return h.NewPartial(
		h.Div(
			h.H1(
				h.Class("font-bold bg-red-500 p-1 rounded"),
				h.Text(fmt.Sprintf("Found differences in %s (loaded version %d)", module, v)),
			),

			h.Table(
				h.Class("font-mono w-full"),

				h.THead(
					h.Tr(
						h.Td(h.Text("Left")),
						h.Td(h.Text("Right")),
					),
				),

				h.TBody(
					h.If(len(keysOnlyLeft) > 0,
						h.Fragment(
							h.Tr(
								h.Td(
									h.Class("font-bold"),
									h.Text("only in left"),
								),
							),
							h.List(keysOnlyLeft, func(key []byte, i int) *h.Element {
								return h.Tr(
									h.Td(IavlKey(key)),
								)
							}),
						),
					),

					h.If(len(keysOnlyRight) > 0,
						h.Fragment(
							h.Tr(
								h.Td(
									h.Class("font-bold"),
									h.Text("only in right"),
								),
							),
							h.List(keysOnlyRight, func(key []byte, i int) *h.Element {
								return h.Tr(
									h.Td(),
									h.Td(IavlKey(key)),
								)
							}),
						),
					),

					h.If(len(keysDifferentValues) > 0,
						h.Fragment(
							h.Tr(
								h.Td(
									h.Class("font-bold"),
									h.Text("different values"),
								),
							),
							h.List(keysDifferentValues, func(key []byte, i int) *h.Element {
								return h.Tr(
									h.Td(IavlKey(key)),
									h.Td(IavlKey(key)),
								)
							}),
						),
					),
				),
			),
		),
	)
}

func IavlKey(k []byte) *h.Element {
	hex := h.Text(hex.EncodeToString(k))
	pretty := prettyKey(k)
	if pretty == "" {
		return h.P(hex)
	}

	return h.Div(
		h.P(h.Text(pretty)),
		h.P(
			h.Class("text-xs text-gray-500"),
			hex,
		),
	)
}

func prettyKey(k []byte) string {
	cut := bytes.IndexRune(k, ':')
	if cut == -1 {
		return encodeID(k)
	}
	prefix := k[:cut]
	id := k[cut+1:]
	return fmt.Sprintf("%s:%s", encodeID(prefix), encodeID(id))
}

func encodeID(id []byte) string {
	for _, b := range id {
		if b < 0x20 || b >= 0x80 {
			return ""
		}
	}
	return string(id)
}

func ErrorMessage(err error) *h.Element {
	return h.Div(
		h.Pre(
			h.Class("bg-red-500 p-1 rounded overflow-x-auto"),
			h.Text(err.Error()),
		),
	)
}
