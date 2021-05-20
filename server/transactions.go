package server

import (
	"encoding/json"
	"fetch-rewards/common"
	"net/http"
)

// RegisterTransactionRoutes attaches routes to the given router.
func (s *Server) RegisterTransactionRoutes() {
	s.Router.HandleFunc("/balance", s.balance).Methods("GET")
	s.Router.HandleFunc("/transactions", s.addTransaction).Methods("POST")
	s.Router.HandleFunc("/spend", s.spendPoints).Methods("POST")
}

func (s *Server) balance(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	balances, err := s.PointsService.Balance()
	if err != nil {
		s.Logger.Error(err)
		httpError := common.GetHttpError(err)
		http.Error(w, httpError.Message, httpError.StatusCode)
	}

	err = json.NewEncoder(w).Encode(balances)
	if err != nil {
		s.Logger.Error(err)
		httpError := common.GetHttpError(err)
		http.Error(w, httpError.Message, httpError.StatusCode)
		return
	}
}

func (s *Server) addTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var transaction common.Transaction
	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		s.Logger.Error(err)
		httpError := common.GetHttpError(err)
		http.Error(w, httpError.Message, httpError.StatusCode)
		return
	}

	err = s.Validator.Struct(transaction)
	if err != nil {
		s.Logger.Error(err)
		httpError := common.HttpError{StatusCode: 400, Message: err.Error()}
		http.Error(w, httpError.Message, httpError.StatusCode)
		return
	}

	err = s.PointsService.Add(transaction)
	if err != nil {
		s.Logger.Error(err)
		httpError := common.GetHttpError(err)
		http.Error(w, httpError.Message, httpError.StatusCode)
		return
	}

	err = json.NewEncoder(w).Encode(transaction)
	if err != nil {
		s.Logger.Error(err)
		httpError := common.GetHttpError(err)
		http.Error(w, httpError.Message, httpError.StatusCode)
		return
	}
}

func (s *Server) spendPoints(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var body common.SpendingRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		s.Logger.Error(err)
		httpError := common.HttpError{StatusCode: 400, Message: "Please provide a valid points value."}
		http.Error(w, httpError.Message, httpError.StatusCode)
		return
	}

	err = s.Validator.Struct(body)
	if err != nil {
		s.Logger.Error(err)
		httpError := common.HttpError{StatusCode: 400, Message: "Please provide a valid points value."}
		http.Error(w, httpError.Message, httpError.StatusCode)
		return
	}

	spendResult, err := s.PointsService.Spend(body.Points)
	if err != nil {
		s.Logger.Error(err)
		httpError := common.GetHttpError(err)
		http.Error(w, httpError.Message, httpError.StatusCode)
		return
	}

	err = json.NewEncoder(w).Encode(spendResult)
	if err != nil {
		s.Logger.Error(err)
		httpError := common.GetHttpError(err)
		http.Error(w, httpError.Message, httpError.StatusCode)
		return
	}
}