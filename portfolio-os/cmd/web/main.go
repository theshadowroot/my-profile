package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"

	"portfolio-os/internal/handlers"
	"portfolio-os/internal/middleware"
	"portfolio-os/internal/renderer"
	"portfolio-os/internal/services"
	"portfolio-os/internal/web"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// 1. Static Assets & File Servers
	staticFS, err := fs.Sub(web.Files, "static")
	if err != nil {
		log.Fatal(err)
	}
	staticHandler := http.FileServer(http.FS(staticFS))
	uploadsHandler := http.FileServer(http.Dir("uploads"))

	// 2. Services & Renderer
	renderer, err := renderer.New(web.Files)
	if err != nil {
		log.Fatal(err)
	}

	portfolioService, err := services.NewPortfolioService("data/portfolio.json")
	if err != nil {
		log.Fatal(err)
	}

	analyticsService, err := services.NewAnalyticsService()
	if err != nil {
		log.Fatal(err)
	}

	// 3. Handlers
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		log.Fatal(err)
	}

	adminHandler := handlers.NewAdminHandler(
		renderer,
		portfolioService,
		cloudinaryService,
	)
	homeHandler := handlers.NewHomeHandler(renderer, portfolioService)

	// 4. Mux Routes
	mux := http.NewServeMux()

	// Static & Upload File Handlers
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticHandler))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", uploadsHandler))

	// Page Views
	mux.HandleFunc("GET /{$}", homeHandler.HandleHome)
	mux.HandleFunc("GET /client", homeHandler.HandleClient)
	mux.HandleFunc("GET /developer", homeHandler.HandleDeveloper)

	// Auth Endpoints
	mux.HandleFunc("GET /admin/login", adminHandler.Login)
	mux.HandleFunc("POST /admin/login", adminHandler.Login)
	mux.HandleFunc("GET /admin/logout", adminHandler.Logout)

	// Analytics Endpoints
	mux.HandleFunc("POST /api/visits", analyticsHandler.StartVisit)
	mux.HandleFunc("PATCH /api/visits/{id}", analyticsHandler.UpdateDuration)
	mux.Handle("GET /api/admin/analytics", middleware.AdminAuth(http.HandlerFunc(analyticsHandler.GetStats)))

	// Admin CRUD Routes
	mux.Handle("GET /admin", middleware.AdminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if err := renderer.Render(w, "admin", portfolioService.GetPortfolio()); err != nil {
			log.Printf("admin render error: %v", err)
			http.Error(w, "Failed to render admin page", http.StatusInternalServerError)
			return
		}
	})))

	mux.Handle("POST /admin/profile", middleware.AdminAuth(http.HandlerFunc(adminHandler.UpdateProfile)))

	mux.Handle("POST /admin/certificates", middleware.AdminAuth(http.HandlerFunc(adminHandler.AddCertificate)))
	mux.Handle("POST /admin/certificates/{index}", middleware.AdminAuth(http.HandlerFunc(adminHandler.UpdateCertificate)))
	mux.Handle("POST /admin/certificates/{index}/delete", middleware.AdminAuth(http.HandlerFunc(adminHandler.DeleteCertificate)))

	mux.Handle("POST /admin/skills", middleware.AdminAuth(http.HandlerFunc(adminHandler.AddSkill)))
	mux.Handle("POST /admin/skills/{index}", middleware.AdminAuth(http.HandlerFunc(adminHandler.UpdateSkill)))
	mux.Handle("POST /admin/skills/{index}/delete", middleware.AdminAuth(http.HandlerFunc(adminHandler.DeleteSkill)))

	mux.Handle("POST /admin/statistics", middleware.AdminAuth(http.HandlerFunc(adminHandler.AddStatistic)))
	mux.Handle("POST /admin/statistics/{index}", middleware.AdminAuth(http.HandlerFunc(adminHandler.UpdateStatistic)))
	mux.Handle("POST /admin/statistics/{index}/delete", middleware.AdminAuth(http.HandlerFunc(adminHandler.DeleteStatistic)))

	mux.Handle("POST /admin/services", middleware.AdminAuth(http.HandlerFunc(adminHandler.AddService)))
	mux.Handle("POST /admin/services/{index}", middleware.AdminAuth(http.HandlerFunc(adminHandler.UpdateService)))
	mux.Handle("POST /admin/services/{index}/delete", middleware.AdminAuth(http.HandlerFunc(adminHandler.DeleteService)))

	mux.Handle("POST /admin/education", middleware.AdminAuth(http.HandlerFunc(adminHandler.AddEducation)))
	mux.Handle("POST /admin/education/{index}", middleware.AdminAuth(http.HandlerFunc(adminHandler.UpdateEducation)))
	mux.Handle("POST /admin/education/{index}/delete", middleware.AdminAuth(http.HandlerFunc(adminHandler.DeleteEducation)))

	mux.Handle("POST /admin/projects", middleware.AdminAuth(http.HandlerFunc(adminHandler.AddProject)))
	mux.Handle("POST /admin/projects/{index}", middleware.AdminAuth(http.HandlerFunc(adminHandler.UpdateProject)))
	mux.Handle("POST /admin/projects/{index}/delete", middleware.AdminAuth(http.HandlerFunc(adminHandler.DeleteProject)))

	mux.Handle("POST /admin/contact", middleware.AdminAuth(http.HandlerFunc(adminHandler.UpdateContact)))

	mux.Handle("POST /admin/social-links", middleware.AdminAuth(http.HandlerFunc(adminHandler.AddSocialLink)))
	mux.Handle("POST /admin/social-links/{index}", middleware.AdminAuth(http.HandlerFunc(adminHandler.UpdateSocialLink)))
	mux.Handle("POST /admin/social-links/{index}/delete", middleware.AdminAuth(http.HandlerFunc(adminHandler.DeleteSocialLink)))

	mux.Handle("POST /admin/resume", middleware.AdminAuth(http.HandlerFunc(adminHandler.UploadResume)))
	mux.Handle("GET /resume", http.HandlerFunc(adminHandler.DownloadResume))

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
