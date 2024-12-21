package save

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"syap_3/internal/lib/api/response"
	"syap_3/internal/lib/logger/sl"
	"syap_3/internal/lib/random"
	"syap_3/internal/lib/validate"
	"syap_3/internal/storage"
)

const aliasLength = 20

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.url.save"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode request body ", sl.Err(err))

			render.JSON(w, r, response.Error("Failed to decode request body"))

			return
		}

		log.Info("Request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			log.Error("Failed to validate request ", sl.Err(err))

			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}

		alias := req.Alias

		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		if len(alias) != aliasLength {
			errorString := fmt.Sprintf("Alias %s is not an acceptable length of %d", req.Alias, aliasLength)

			log.Info(errorString)

			render.JSON(w, r, response.Error(errorString))

			return
		}

		if !validate.IsSubset(alias) {
			errorString := fmt.Sprintf("Alias %s is not a valid", req.Alias)

			log.Info(errorString)

			render.JSON(w, r, response.Error(errorString))
			  
			return
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("URL already exists", slog.String("url", req.URL))

			render.JSON(w, r, response.Error("URL Already exists"))

			return
		}

		if err != nil {
			log.Error("Failed to save URL ", sl.Err(err))

			render.JSON(w, r, response.Error("Failed to save URL"))

			return
		}

		log.Info("URL saved", slog.Int64("id", id))

		render.JSON(w, r, Response{
			Response: response.OK(),
			Alias:    alias,
		})
	}
}
