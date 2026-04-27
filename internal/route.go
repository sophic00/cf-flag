package flagapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
)

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.handleHealth)
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("POST /createflag", s.handleCreateFlag)
	mux.HandleFunc("GET /listflag", s.handleListFlags)
	mux.HandleFunc("POST /checkflag", s.handleCheckFlag)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
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

func (s *Server) handleListFlags(w http.ResponseWriter, req *http.Request) {
	flags, err := s.listFlags(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list flags")
		return
	}

	writeJSON(w, http.StatusOK, listFlagsResponse{Flags: flags})
}

func (s *Server) handleCheckFlag(w http.ResponseWriter, req *http.Request) {
	var in checkFlagRequest
	if err := decodeJSON(req.Body, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	flagID := strings.TrimSpace(in.FlagID)
	userID := strings.TrimSpace(in.UserID)
	if flagID == "" || userID == "" {
		writeError(w, http.StatusBadRequest, "flagId and userId are required")
		return
	}
	userCountry := strings.ToUpper(strings.TrimSpace(in.UserCountry))

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
	case RuleTypeCountryPercentage:
		active = strings.EqualFold(userCountry, rule.Country) &&
			PercentageEnabled(flagID, userID, rule.Percentage, s.hashKey)
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
