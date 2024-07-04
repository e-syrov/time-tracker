package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time-tracker/internal/config"
	"time-tracker/internal/database"
	"time-tracker/internal/handlers"
	"time-tracker/internal/logger"
)

func Run() error {

	logger.InitLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	err = database.InitDB(cfg)
	if err != nil {
		return err
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/users", func(r chi.Router) {

		r.Get("/", handlers.GetUsers)
		r.Get("/{id}/worklog", handlers.GetWorkLog)

		r.Post("/add", handlers.AddUser)
		r.Post("/{id}/task/start", handlers.StartTask)
		r.Post("/task/{id}/stop", handlers.StopTask)

		r.Delete("/{id}", handlers.DeleteUser)

		r.Put("/{id}", handlers.UpdateUser)

	})

	err = http.ListenAndServe("localhost:8080", r)
	if err != nil {
		return err
	}
	return nil
}
