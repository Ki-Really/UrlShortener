package redirect

import (
	"UrlShortener/UrlShortener/internal/lib/api/response"
	"UrlShortener/UrlShortener/internal/lib/logger/sl"
	"UrlShortener/UrlShortener/internal/storage"
	"errors"
	"golang.org/x/exp/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, response.Error("not found"))
			return
		}

		resUrl, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Info("Url not found", "alias", alias)
			render.JSON(w, r, response.Error("not found"))
			return
		}

		if err != nil{
			log.Error("failed to get url",sl.Err(err))
			render.JSON(w,r,response.Error("internal"))
		}

		log.Info("Got url",slog.String("url",resUrl))
		http.Redirect(w,r,resUrl,http.StatusFound )
	}
}
