package httpapi

import (
	"encoding/json"
	"net/http"

	"gitlab.com/ruslanty/url-checker/internal/checker"
)

type Handler struct {
	service *checker.Service
}

func New(service *checker.Service) http.Handler {
	h := &Handler{
		service: service,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /checks", h.check)

	return mux
}

type checkRequest struct {
	URLs []string `json:"urls"`
}

func (h *Handler) health(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) check(response http.ResponseWriter, request *http.Request) {
	var input checkRequest

	if err := decode(response, request, &input); err != nil || len(input.URLs) == 0 || len(input.URLs) > 20 {
		writeJSON(response, http.StatusBadRequest, map[string]string{
			"error": "urls must contain from 1 to 20 items",
		})

		return
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"results": h.service.CheckMany(request.Context(), input.URLs),
	})
}

func decode(response http.ResponseWriter, request *http.Request, target any) error {
	request.Body = http.MaxBytesReader(response, request.Body, 16<<10)

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)

	_ = json.NewEncoder(response).Encode(value)
}
