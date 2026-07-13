package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Server is the HTTP layer. Handlers only parse requests, call the Service and
// write responses; they hold no business rules.
type Server struct {
	app     *App
	handler http.Handler
}

// NewServer builds the router, wires the middleware chain and returns a Server
// usable as an http.Handler.
func NewServer(app *App) *Server {
	s := &Server{app: app}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)

	mux.HandleFunc("POST /api/users", s.handleCreateUser)
	mux.HandleFunc("GET /api/users/{id}", s.handleGetUser)
	mux.HandleFunc("PUT /api/users/{id}", s.handleUpdateUser)
	mux.HandleFunc("GET /api/users/{id}/skills", s.handleGetUserSkills)
	mux.HandleFunc("PUT /api/users/{id}/skills", s.handleSetUserSkills)

	mux.HandleFunc("GET /api/services", s.handleListServices)
	mux.HandleFunc("POST /api/services", s.handleCreateService)
	mux.HandleFunc("GET /api/services/{id}", s.handleGetService)
	mux.HandleFunc("PUT /api/services/{id}", s.handleUpdateService)
	mux.HandleFunc("DELETE /api/services/{id}", s.handleDeleteService)

	// Outermost first: log wraps everything, recover catches handler panics,
	// cors sets the headers.
	s.handler = chain(mux, logMiddleware, recoverMiddleware, corsMiddleware)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Middlewares ---

type middleware func(http.Handler) http.Handler

// chain applies middlewares so that the first one is the outermost.
func chain(h http.Handler, mws ...middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// statusRecorder captures the response status for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic récupérée: %v", rec)
				writeError(w, http.StatusInternalServerError, "erreur interne")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Response / request helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			log.Printf("encodage réponse: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// respondError maps a business error to its HTTP response, hiding internal
// details behind a generic 500 message.
func respondError(w http.ResponseWriter, err error) {
	status := httpStatus(err)
	if status == http.StatusInternalServerError {
		log.Printf("erreur interne: %v", err)
		writeError(w, status, "erreur interne")
		return
	}
	writeError(w, status, err.Error())
}

func decodeJSON(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return &ValidationError{Message: "corps de requête JSON invalide"}
	}
	return nil
}

// pathID extracts and validates the {id} path parameter.
func pathID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, &ValidationError{Field: "id", Message: "identifiant invalide"}
	}
	return id, nil
}

// authUserID reads the caller identity from the X-User-ID header.
func authUserID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.Header.Get("X-User-ID"))
	if err != nil || id <= 0 {
		return 0, ErrUnauthorized
	}
	return id, nil
}
