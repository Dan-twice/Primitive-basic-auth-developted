package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func newRouter(handler *chi.Mux) *http.Server {
	// Pre-connetion params for Server
	return &http.Server{
		Addr:         "",
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}
}

func initRouter(handler *chi.Mux, address string,
	readTime, writeTime time.Duration) *http.Server {
	srv := newRouter(handler)
	srv.Addr = address
	srv.ReadTimeout = time.Duration(readTime)
	srv.WriteTimeout = time.Duration(writeTime)
	return srv
}

func chiHandlers(chiRouter *chi.Mux) *chi.Mux {
	chiRouter.Use(middleware.RequestID)
	chiRouter.Use(middleware.Logger)
	chiRouter.Use(middleware.Recoverer)

	chiRouter.Get("/", RegistrationHandler)
	chiRouter.Post("/post", PostHandler)
	chiRouter.Get("/un", Unprotected)
	chiRouter.Get("/protected", BasicAuth(ProtectedHandler))
	chiRouter.Get("/show", ShowUsers)
	chiRouter.Get("/log", LogOut)
	return chiRouter
}

func ServerDealer() {
	chiRouter := chi.NewRouter()
	chiRouter = chiHandlers(chiRouter)

	srv := initRouter(chiRouter, ":4000", 10*time.Second, 30*time.Second)

	// Use https with «SSL certificate». mkcert localhost creates these files
	err = srv.ListenAndServeTLS("localhost.pem",
		"localhost-key.pem")
	if err != nil {
		log.Println("ListenAndServeTLS problem")
		log.Fatal(err)
	}
	// log.Printf("starting server on %s", srv.Addr)
	log.Println("Shouldn't been riched this point")
}
