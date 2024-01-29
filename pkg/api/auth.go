package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
)

func (a *API) getAuth(w http.ResponseWriter, r *http.Request) {
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		a.sendError(w, err.Error(), http.StatusUnauthorized)
		return
	}
	jwt, _ := token.AsMap(r.Context())
	a.writeJSON(w, jwt)
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *API) createAuth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req AuthRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		a.sendError(w, "Failed to decode request.", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		a.sendError(w, "Invalid username or password.", http.StatusBadRequest)
		return
	}
	if req.Username != "admin" {
		a.sendError(w, "Invalid username or password.", http.StatusBadRequest)
		return
	}
	if a.cfg.WebuiAdminPasswordHash == "" {
		a.sendError(w, "Webui admin password is not set.", http.StatusBadRequest)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(a.cfg.WebuiAdminPasswordHash), []byte(req.Password))
	if err != nil {
		a.sendError(w, "Invalid username or password.", http.StatusBadRequest)
		return
	}
	claims := map[string]any{"username": req.Username}
	jwtauth.SetIssuedNow(claims)
	// set expiry to 1000 days
	jwtauth.SetExpiryIn(claims, time.Hour*24*1000)
	_, token, err := a.tokenAuth.Encode(claims)
	if err != nil {
		a.sendError(w, "Failed to create JWT.", http.StatusInternalServerError)
		return
	}
	a.writeJSON(w, map[string]string{"token": token})
}
