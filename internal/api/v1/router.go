package v1

import (
	_ "github.com/GlebMoskalev/go-pickup-point-api/docs"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/api/middleware"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(services *service.Services) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		SetupAuthRoutes(r, services.Auth)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(services.Auth))

			r.Route("/pvz", func(r chi.Router) {
				SetupPVZRoutes(r, services.PVZ, services.Product, services.Reception)
			})

			r.Route("/receptions", func(r chi.Router) {
				SetupReceptionRoutes(r, services.Reception)
			})

			r.Route("/products", func(r chi.Router) {
				SetupProductRoutes(r, services.Product)
			})
		})
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return r
}
