// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/jtolds/webhelp"
	"github.com/jtolds/webhelp/sessions"
)

var (
	listenAddr   = flag.String("addr", ":8080", "address to listen on")
	cookieSecret = flag.String("cookie_secret", "abcdef0123456789",
		"the secret for securing cookie information")

	projectId = webhelp.NewIntArgMux()
)

func main() {
	flag.Parse()
	loadOAuth2()
	secret, err := hex.DecodeString(*cookieSecret)
	if err != nil {
		panic(err)
	}

	renderer, err := NewRenderer()
	if err != nil {
		panic(err)
	}

	app, err := NewApp()
	if err != nil {
		panic(err)
	}
	defer app.Close()

	routes := webhelp.LoggingHandler(
		sessions.HandlerWithStore(sessions.NewCookieStore(secret),
			webhelp.OverlayMux{Fallback: oauth2.LoginRequired(webhelp.DirMux{
				"": renderer.Render(app.ProjectList),
				"project": projectId.ShiftIf(webhelp.MethodMux{
					"GET": renderer.Render(app.Project),
				}, webhelp.ExactPath(webhelp.MethodMux{
					"GET":  webhelp.RedirectHandler("/"),
					"POST": renderer.Process(app.NewProject),
				}))}),
				Overlay: webhelp.DirMux{
					"auth": oauth2}}))

	switch flag.Arg(0) {
	case "migrate":
		err := app.Migrate()
		if err != nil {
			panic(err)
		}
	case "serve":
		panic(webhelp.ListenAndServe(*listenAddr, routes))
	case "routes":
		webhelp.PrintRoutes(os.Stdout, routes)
	default:
		fmt.Printf("Usage: %s <serve|migrate>\n", os.Args[0])
	}
}
