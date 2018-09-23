package app

import (
	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/iheanyi/simple-canary/internal/db"
	"github.com/sirupsen/logrus"
)

// App is an instance of the dashboard for the canary.
type App struct {
	l  logrus.FieldLogger
	db db.CanaryStore
}

func New(db db.CanaryStore, r *mux.Router) *App {
	app := &App{
		l:  logrus.WithField("component", "app"),
		db: db,
	}

	r.Handle("/", handler.Playground("GraphQL Playground", "/query"))
	r.Handle("/query", handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{
		db: db,
	}})))

	// TODO: Setup GraphQL server here please.
	return app
}
