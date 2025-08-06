package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	ev "github.com/AfterShip/email-verifier"
	"github.com/goccy/go-json"
)

var (
	evPool sync.Pool = sync.Pool{
		New: func() any {
			return ev.NewVerifier().EnableGravatarCheck()
		},
	}

	successResponse = &SuccessfulResponse{
		Identity: &Identity{
			MetadataAdmin: map[string]string{
				"domain_authorized": "true",
			},
		},
	}
)

func (a *AppState) respondWithInterface(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		a.logger.Error().Err(err).Msg("failed writing response body")
	}
}

func NewValidationResponse(text string, context map[string]interface{}) *ValidationResponse {
	return &ValidationResponse{
		Messages: []FieldMessage{
			{
				InstancePtr: "#/traits/email",
				Messages: []ValidationMessage{
					{
						ID:      1001,
						Text:    text,
						Type:    "error",
						Context: context,
					},
				},
			},
		},
	}

}

func (a *AppState) Validate(w http.ResponseWriter, r *http.Request) {
	var req WebhookRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		a.logger.Error().Err(err).Msg("failed deserializing request body")
		w.WriteHeader(http.StatusBadRequest)
		a.respondWithInterface(w, map[string]string{"error": "invalid request body"})
		return
	}

	verifier := evPool.Get().(*ev.Verifier)
	defer evPool.Put(verifier)

	verifResult, err := verifier.Verify(req.Email)
	if err != nil {
		response := NewValidationResponse("Failed processing your request. Please try again or contact the administrator.", nil)
		a.respondWithInterface(w, response)
		return
	}

	a.logger.Info().Interface("verification_result", verifResult).Msg("verification result ready")

	if !verifResult.Syntax.Valid || verifResult.Disposable || verifResult.RoleAccount || verifResult.Free || !verifResult.HasMxRecords {
		response := NewValidationResponse("Provided email is invalid. Please provide a valid business email.", nil)
		a.respondWithInterface(w, response)
		return
	}

	domain := strings.ToLower(verifResult.Syntax.Domain)

	if !a.allowedDomains[domain] {
		w.WriteHeader(http.StatusBadRequest)

		response := NewValidationResponse(fmt.Sprintf("Registration is restricted to authorized domains. Domain '%s' is not allowed.", domain), nil)

		a.respondWithInterface(w, response)
		return
	}

	w.WriteHeader(http.StatusOK)
	a.respondWithInterface(w, successResponse)
}
