package server

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"openradar/internal/api"

	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"openradar/app"
)

type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*visitorLimiter
}

type visitorLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newIPLimiter() *ipLimiter {
	ipl := &ipLimiter{
		limiters: make(map[string]*visitorLimiter),
	}
	go ipl.cleanup()
	return ipl
}

func (ipl *ipLimiter) getLimiter(ip string) *rate.Limiter {
	ipl.mu.Lock()
	defer ipl.mu.Unlock()

	v, exists := ipl.limiters[ip]
	if !exists {
		v = &visitorLimiter{
			limiter: rate.NewLimiter(10, 15),
		}
		ipl.limiters[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (ipl *ipLimiter) cleanup() {
	for {
		time.Sleep(3 * time.Minute)
		ipl.mu.Lock()
		for ip, v := range ipl.limiters {
			if time.Since(v.lastSeen) > 5*time.Minute {
				delete(ipl.limiters, ip)
			}
		}
		ipl.mu.Unlock()
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func StartServer(db *gorm.DB) {
	router := chi.NewRouter()

	ipl := newIPLimiter()
	rateLimitMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limiter := ipl.getLimiter(r.RemoteAddr)
			if !limiter.Allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	router.Use(middleware.Logger)
	router.Use(corsMiddleware)
	router.Use(rateLimitMiddleware)

	publicFS, err := fs.Sub(app.Dist, "public")
	if err != nil {
		log.Fatal(err)
	}
	router.Mount("/public", http.StripPrefix("/public/", http.FileServer(http.FS(publicFS))))

	distFS, err := fs.Sub(app.Dist, "dist")
	if err != nil {
		log.Fatal(err)
	}

	router.Get("/findings", func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		page := 1
		if pageStr != "" {
			if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
				page = val
			}
		}

		pageSizeStr := r.URL.Query().Get("page_size")
		pageSize := 25
		if pageSizeStr != "" {
			if val, err := strconv.Atoi(pageSizeStr); err == nil && val > 0 && val <= 100 {
				pageSize = val
			}
		}

		provider := r.URL.Query().Get("provider")
		if provider == "" {
			provider = "*"
		}

		minAge := r.URL.Query().Get("min_age")
		if minAge == "" {
			minAge = "24h"
		}

		paginatedFindings, err := api.GetLatestFindings(
			page,
			pageSize,
			provider,
			minAge,
			db,
		)
		if err != nil {
			log.Printf("GET /findings error: %v", err)
			http.Error(w, "failed to fetch findings", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, paginatedFindings)
	})

	router.Get("/findings/count", func(w http.ResponseWriter, r *http.Request) {
		count, err := api.GetFindingsCount(db)
		if err != nil {
			log.Printf("GET /findings/count error: %v", err)
			http.Error(w, "failed to count findings", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]int64{"total_count": count})
	})

	router.Get("/repository", func(w http.ResponseWriter, r *http.Request) {
		repoUrl := r.URL.Query().Get("repo_url")
		if repoUrl == "" {
			http.Error(w, "repo_url parameter is required", http.StatusBadRequest)
			return
		}

		repositories, err := api.GetRepositoryInfo(repoUrl, db)
		if err != nil {
			log.Printf("GET /repository error: %v", err)
			http.Error(w, "failed to fetch repository", http.StatusInternalServerError)
			return
		}

		if len(repositories) == 0 {
			http.Error(w, "Repository not found", http.StatusNotFound)
			return
		}

		writeJSON(w, http.StatusOK, repositories[0])
	})

	router.Get("/repository/findings", func(w http.ResponseWriter, r *http.Request) {
		repoUrl := r.URL.Query().Get("repo_url")
		if repoUrl == "" {
			http.Error(w, "repo_url parameter is required", http.StatusBadRequest)
			return
		}

		pageStr := r.URL.Query().Get("page")
		page := 1
		if pageStr != "" {
			if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
				page = val
			}
		}

		pageSizeStr := r.URL.Query().Get("page_size")
		pageSize := 25
		if pageSizeStr != "" {
			if val, err := strconv.Atoi(pageSizeStr); err == nil && val > 0 && val <= 100 {
				pageSize = val
			}
		}

		paginatedFindings, err := api.GetFindingsFromRepository(repoUrl, page, pageSize, db)
		if err != nil {
			log.Printf("GET /repository/findings error: %v", err)
			http.Error(w, "failed to fetch findings", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, paginatedFindings)
	})

	router.Get("/repositories", func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		page := 1
		if pageStr != "" {
			if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
				page = val
			}
		}

		pageSizeStr := r.URL.Query().Get("page_size")
		pageSize := 25
		if pageSizeStr != "" {
			if val, err := strconv.Atoi(pageSizeStr); err == nil && val > 0 && val <= 100 {
				pageSize = val
			}
		}

		paginatedRepos, err := api.GetAllRepositories(page, pageSize, db)
		if err != nil {
			log.Printf("GET /repositories error: %v", err)
			http.Error(w, "failed to fetch repositories", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, paginatedRepos)
	})

	fileServer := http.FileServer(http.FS(distFS))

	router.Handle("/*", fileServer)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	srv.ListenAndServe()
}
