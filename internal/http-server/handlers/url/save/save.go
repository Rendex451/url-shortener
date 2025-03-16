package save

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	"url-shortener/internal/lib/random"
	"url-shortener/internal/lib/response"
	"url-shortener/internal/storage"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Alias  string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlSaver URLSaver, randomAliasLength int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		var req Request

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", slog.String("error", err.Error()))
			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", slog.String("error", err.Error()))
			render.JSON(w, r, response.Error("invalid request"))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		alias := req.Alias

		if alias == "" {
			isValidAlias := false
			for !isValidAlias {
				alias = random.GetRandomString(randomAliasLength)
				_, err := urlSaver.GetURL(alias)
				if errors.Is(err, storage.ErrUrlNotFound) {
					isValidAlias = true
				}
			}
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrUrlExists) {
			log.Info("url already exists", slog.String("url", req.URL))
			render.JSON(w, r, response.Error("url already exists"))
			return
		}

		if err != nil {
			log.Error("failed to save url", slog.String("error", err.Error()))
			render.JSON(w, r, response.Error("failed to save url"))
			return
		}

		log.Info("url saved", slog.Int64("id", id))

		render.JSON(w, r, response.OK())
	}
}
