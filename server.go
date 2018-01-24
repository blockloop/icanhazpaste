package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/apex/log"
	"github.com/blockloop/pbpaste/rand"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-redis/redis"
	"github.com/pressly/chi/render"
)

var (
	ErrNotFound = fmt.Errorf("entity not found")
)

type Handler struct {
	redis *redis.Client
	store *Store
}

func NewHandler(redisClient *redis.Client) *Handler {
	return &Handler{
		redis: redisClient,
		store: NewStore(redisClient),
	}
}

func (h *Handler) RegisterRoutes(mux chi.Router) {
	mux.With(
		middleware.AllowContentType("application/x-www-form-urlencoded", "text/plain"),
		ipRateLimiter(h.redis),
	).Post("/", h.postForm)

	mux.Get("/", h.getForm)

	mux.Get("/x/{name}", h.getPaste)
}

func (h *Handler) getForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, HTMLForm)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func (h *Handler) FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(fs.ServeHTTP))
}

func (h *Handler) getPaste(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	b, err := h.store.Get(name)
	if err != nil {
		sendError(w, 500, err)
		return
	}

	if len(b) == 0 {
		sendError(w, 404, ErrNotFound)
		return
	}

	render.PlainText(w, r, string(b))
}

func (h *Handler) postForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendError(w, 400, err)
		return
	}

	clip := r.FormValue("clip")
	if clip == "" {
		byt, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sendError(w, 500, err)
			return
		}
		clip = string(byt)
	}

	fname := rand.String(20)
	if err := h.store.Put(fname, clip); err != nil {
		sendError(w, 500, err)
		return
	}
	furl := "/x/" + fname

	data := map[string]interface{}{
		"Name": fname,
		"URL":  furl,
	}

	switch render.GetAcceptedContentType(r) {
	case render.ContentTypeHTML:
		if err := HTMLCreatedTemplate.Execute(w, data); err != nil {
			sendError(w, 500, err)
		}
	case render.ContentTypeJSON:
		render.JSON(w, r, &data)
	default:
		u := &(*r.URL)
		u.Host = r.Host
		u.Scheme = "http"
		if r.TLS != nil {
			u.Scheme = "https"
		}
		u.Path = "/x/" + fname
		render.PlainText(w, r, u.String())
	}

}

func sendError(w http.ResponseWriter, status int, err error) {
	msg := err.Error()
	if status > 499 {
		log.WithError(err).Error(msg)
	}
	w.WriteHeader(status)
	fmt.Fprint(w, msg)
}
