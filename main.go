package main

import (
	"iavlweb/__htmgo"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/maddalax/htmgo/framework/h"
	"github.com/maddalax/htmgo/framework/service"

	"iavlweb/internal/iavl"
)

func main() {
	dbPathLeft := os.Getenv("IAVL_LEFT")
	dbPathRight := os.Getenv("IAVL_RIGHT")
	log.Println("Opening IAVL Pair", dbPathLeft, dbPathRight)
	iavlPair, err := iavl.OpenDBPair(dbPathLeft, dbPathRight)
	if err != nil {
		panic(err)
	}
	defer func() {
		log.Println("Closing IAVL Pair")
		err := iavlPair.Close()
		log.Println("IAVL Pair closed", err)
	}()

	locator := service.NewLocator()
	service.Set(locator, service.Singleton, func() *iavl.DBPair {
		return iavlPair
	})

	h.Start(h.AppOpts{
		ServiceLocator: locator,
		LiveReload:     true,
		Register: func(app *h.App) {
			sub, err := fs.Sub(GetStaticAssets(), "assets/dist")

			if err != nil {
				panic(err)
			}

			http.FileServerFS(sub)

			app.Router.Handle("/public/*", http.StripPrefix("/public", http.FileServerFS(sub)))
			__htmgo.Register(app.Router)
		},
	})
}
