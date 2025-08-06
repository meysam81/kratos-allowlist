package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
)

func (a *AppState) respondWithInterface(w http.ResponseWriter, body interface{}) {
	err := json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
	if err != nil {
		a.logger.Error().Err(err).Msg("failed writing response body")
	}
}

func (a *AppState) Validate(w http.ResponseWriter, r *http.Request) {
	var req WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		a.respondWithInterface(w, map[string]string{"error": "invalid request body"})
		return
	}

	parts := strings.Split(req.Email, "@")
	if len(parts) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		a.respondWithInterface(w, map[string]string{"error": "invalid email format"})
		return
	}

	domain := strings.ToLower(parts[1])

	if !a.allowedDomains[domain] {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		response := ValidationResponse{
			Messages: []FieldMessage{
				{
					InstancePtr: "#/traits/email",
					Messages: []ValidationMessage{
						{
							ID:   1001,
							Text: fmt.Sprintf("Registration is restricted to authorized domains. Domain '%s' is not allowed.", domain),
							Type: "error",
							Context: map[string]interface{}{
								"domain": domain,
							},
						},
					},
				},
			},
		}

		a.respondWithInterface(w, response)
		return
	}

	w.WriteHeader(http.StatusOK)
}
