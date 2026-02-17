package server

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"openradar/internal/api"

	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"openradar/app"
)

func StartServer(db *gorm.DB) {
	router := chi.NewRouter()

	var limiter = rate.NewLimiter(10, 15) // 10 r/s, burst of 15

	// Load frontend (app)
	distFS, err := fs.Sub(app.Dist, "dist")
	if err != nil {
		log.Fatal(err)
	}

	router.Use(middleware.Logger)

	rateLimitMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	router.Use(rateLimitMiddleware)

	// GET /findings?page=1&page_size=25&provider=google&min_age=24h
	router.Get("/findings", func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		page := 1 // default
		if pageStr != "" {
			if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
				page = val
			}
		}

		pageSizeStr := r.URL.Query().Get("page_size")
		pageSize := 25 // default
		if pageSizeStr != "" {
			if val, err := strconv.Atoi(pageSizeStr); err == nil && val > 0 && val <= 100 {
				pageSize = val
			}
		}

		provider := r.URL.Query().Get("provider")
		if provider == "" {
			provider = "*" // default
		}

		minAge := r.URL.Query().Get("min_age")
		if minAge == "" {
			minAge = "24h" // default
		}

		paginatedFindings, err := api.GetLatestFindings(
			page,
			pageSize,
			provider,
			minAge,
			db,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(paginatedFindings); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	// GET /repository?repo_url=<url>
	router.Get("/repository", func(w http.ResponseWriter, r *http.Request) {
		repoUrl := r.URL.Query().Get("repo_url")
		if repoUrl == "" {
			http.Error(w, "repo_url parameter is required", http.StatusBadRequest)
			return
		}

		repositories, err := api.GetRepositoryInfo(repoUrl, db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(repositories) == 0 {
			http.Error(w, "Repository not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(repositories[0]); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	// GET /repository/findings?repo_url=<url>&page=1&page_size=25
	router.Get("/repository/findings", func(w http.ResponseWriter, r *http.Request) {
		repoUrl := r.URL.Query().Get("repo_url")
		if repoUrl == "" {
			http.Error(w, "repo_url parameter is required", http.StatusBadRequest)
			return
		}

		pageStr := r.URL.Query().Get("page")
		page := 1 // default
		if pageStr != "" {
			if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
				page = val
			}
		}

		pageSizeStr := r.URL.Query().Get("page_size")
		pageSize := 25 // default
		if pageSizeStr != "" {
			if val, err := strconv.Atoi(pageSizeStr); err == nil && val > 0 && val <= 100 {
				pageSize = val
			}
		}

		paginatedFindings, err := api.GetFindingsFromRepository(repoUrl, page, pageSize, db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(paginatedFindings); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	// GET /repositories?page=1&page_size=25
	router.Get("/repositories", func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		page := 1 // default
		if pageStr != "" {
			if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
				page = val
			}
		}

		pageSizeStr := r.URL.Query().Get("page_size")
		pageSize := 25 // default
		if pageSizeStr != "" {
			if val, err := strconv.Atoi(pageSizeStr); err == nil && val > 0 && val <= 100 {
				pageSize = val
			}
		}

		paginatedRepos, err := api.GetAllRepositories(page, pageSize, db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(paginatedRepos); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	fileServer := http.FileServer(http.FS(distFS))

	// GET / ROOT
	router.Handle("/*", fileServer)

	http.ListenAndServe(":8080", router)
}
