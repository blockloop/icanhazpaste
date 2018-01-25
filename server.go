package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/apex/log"
	"github.com/blockloop/pbpaste/rand"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-redis/redis"
	"github.com/pressly/chi/render"
)

var (
	ErrNotFound = fmt.Errorf("paste not found")
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

func (h *Handler) getPaste(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	text, ttl, err := h.store.Get(name)
	if err != nil {
		sendError(w, 500, err)
		return
	}

	if len(text) == 0 {
		sendError(w, 404, ErrNotFound)
		return
	}

	w.Header().Set("Expires", ttl.Format(time.RFC1123))

	render.PlainText(w, r, text)
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
	uri := newURL(r, fname)

	data := map[string]interface{}{
		"Name": fname,
		"URL":  uri,
	}

	switch render.GetAcceptedContentType(r) {
	case render.ContentTypeHTML:
		if err := HTMLCreatedTemplate.Execute(w, data); err != nil {
			sendError(w, 500, err)
		}
	case render.ContentTypeJSON:
		render.JSON(w, r, &data)
	default:
		render.PlainText(w, r, uri)
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

func newURL(r *http.Request, name string) string {
	u, _ := url.ParseRequestURI(r.RequestURI)
	u.Scheme, u.Host, u.Path = "http", r.Host, "/x/"+name
	if r.TLS != nil {
		u.Scheme = "https"
	}
	return u.String()
}
