package save

import (
	resp "UrlShortener/UrlShortener/internal/lib/api/response"
	"UrlShortener/UrlShortener/internal/lib/logger/sl"
	"UrlShortener/UrlShortener/internal/lib/random"
	"UrlShortener/UrlShortener/internal/storage"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slog"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

const aliasLength = 8

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			log.Error("Failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("Failed to decode request"))
			return
		}

		log.Info("Request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("Invalid request", sl.Err(err))
			validateErr := err.(validator.ValidationErrors)
			render.JSON(w, r, resp.Error("Invalid request"))
			render.JSON(w, r, resp.ValidationError(validateErr))
		}

		alias := req.Alias

		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id,err:=urlSaver.SaveURL(req.URL,alias)

		if errors.Is(err,storage.ErrUrlExists){
			log.Info("Url already exists",slog.String("url",req.URL))
			render.JSON(w,r,resp.Error("Url already exists"))
			return 
		}

		log.Info("Url added",slog.Int64("id",id))
		
		render.JSON(w,r,Response{
			Response: resp.OK(),
			Alias: alias,
		})
	}
}
