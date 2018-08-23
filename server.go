package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/blockloop/icanhazpaste/rand"
	"github.com/apex/log"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/pressly/chi/render"
)

const (
	Byte = 1.0 << (10 * iota)
	Kilobyte
	Megabyte
	Gigabyte
	Terabyte

	helpText = `
 # send a file
 curl --data-binary @./notes.txt icanhazpaste.com

 # send some raw text
 curl --data-binary 'Hello, there!' icanhazpaste.com

 # send from stdin
 journalctl -xe -u dnsmasq | curl --data-binary @- icanhazpaste.com
`
)

var (
	// ErrNotFound is an error indicating a paste was not found
	ErrNotFound = fmt.Errorf("paste not found")
	// ErrInternal is an internal error
	ErrInternal = fmt.Errorf("internal error")
)

// Handler is an HTTP handler
type Handler struct {
	redis *redis.Client
	store *Store
}

// NewHandler constructs a new handler with the given client
func NewHandler(redisClient *redis.Client) *Handler {
	return &Handler{
		redis: redisClient,
		store: NewStore(redisClient),
	}
}

// RegisterRoutes registers the HTTP routes with the given router
func (h *Handler) RegisterRoutes(mux chi.Router) {
	mux.With(
		middleware.AllowContentType("application/x-www-form-urlencoded", "text/plain"),
		ipRateLimiter(h.redis),
	).Post("/", h.postForm)

	mux.Get("/styles.css", h.getStyles)
	mux.Get("/", h.getForm)
	mux.Get("/help", h.getHelp)
	mux.Get("/x/{name}", h.getPaste)
}

func (h *Handler) getStyles(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "styles.css")
}

func (h *Handler) getForm(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "form.html")
}

func (h *Handler) getHelp(w http.ResponseWriter, r *http.Request) {
	switch render.GetAcceptedContentType(r) {
	case render.ContentTypeHTML:
		data := map[string]interface{}{"Text": helpText}
		if err := HTMLHelpTemplate.Execute(w, data); err != nil {
			sendError(w, 500, err)
		}
	default:
		fmt.Fprintf(w, helpText)
	}
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

// postForm handles all posts of pastes
//
// when using curl the default content-type is form-urlencoded even if the intention
// was to submit plain text. If the field 'clip' does not exist (specified in the
// html form served from us) we will assume that the user was submitting a plaintext form
func (h *Handler) postForm(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > 1*Megabyte {
		http.Error(w, "Request is too large. Must be < 1MB.", http.StatusRequestEntityTooLarge)
		return
	}

	rawBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(w, 500, errors.Wrap(err, "failed to read body"))
		return
	}
	body := string(rawBody)

	if render.GetRequestContentType(r) == render.ContentTypeForm {
		form, err := url.ParseQuery(body)
		if err == nil {
			if clip := form["clip"]; len(clip) > 0 {
				body = clip[0]
			}
		}
	}

	// nothing was submitted so just render the form again
	if len(body) == 0 {
		h.getForm(w, r)
		return
	}

	fname := rand.String(20)
	if err := h.store.Put(fname, body); err != nil {
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
	w.WriteHeader(status)
	if status > 499 {
		log.WithError(err).Error(msg)
		if !debug {
			msg = ErrInternal.Error()
		}
	}
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
