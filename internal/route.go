package flagapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
)

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("POST /users", s.handleCreateUser)
	mux.HandleFunc("POST /flags", s.handleCreateFlag)
	mux.HandleFunc("GET /flags/{flagID}/users/{userID}/active", s.handleFlagActive)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	var in createUserRequest
	if err := decodeJSON(req.Body, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := normalizeUserInput(in)
	if err != nil {
		if errors.Is(err, errIDGeneration) {
			writeError(w, http.StatusInternalServerError, "failed to generate user id")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = s.insertUser(req.Context(), user)
	if err != nil {
		if isUniqueConstraintError(err) {
			writeError(w, http.StatusConflict, "user already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, createUserResponse{User: user})
}

func (s *Server) handleCreateFlag(w http.ResponseWriter, req *http.Request) {
	var in createFlagRequest
	if err := decodeJSON(req.Body, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	flag, err := normalizeFlagInput(in)
	if err != nil {
		if errors.Is(err, errIDGeneration) {
			writeError(w, http.StatusInternalServerError, "failed to generate flag id")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = s.insertFlag(req.Context(), flag)
	if err != nil {
		if isUniqueConstraintError(err) {
			writeError(w, http.StatusConflict, "flag already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create flag")
		return
	}

	writeJSON(w, http.StatusCreated, createFlagResponse{Flag: flag})
}

func (s *Server) handleFlagActive(w http.ResponseWriter, req *http.Request) {
	flagID := strings.TrimSpace(req.PathValue("flagID"))
	userID := strings.TrimSpace(req.PathValue("userID"))
	if flagID == "" || userID == "" {
		writeError(w, http.StatusBadRequest, "flagID and userID are required")
		return
	}

	ctx := req.Context()
	rawRule, err := s.getFlagRule(ctx, flagID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load flag")
		return
	}

	userCountry, err := s.getUserCountry(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}

	rule, err := ParseFlagRule(rawRule)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "stored flag rule is invalid")
		return
	}

	active := false
	switch rule.TypeName {
	case RuleTypeCountry:
		active = strings.EqualFold(userCountry, rule.Country)
	case RuleTypePercentage:
		active = PercentageEnabled(flagID, userID, rule.Percentage, s.hashKey)
	default:
		writeError(w, http.StatusInternalServerError, "unsupported rule type")
		return
	}

	writeJSON(w, http.StatusOK, flagStatusResponse{
		FlagID: flagID,
		UserID: userID,
		Rule:   rawRule,
		Active: active,
	})
}
