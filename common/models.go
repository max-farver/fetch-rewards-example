package common

import "time"

type Transaction struct {
	Payer string `json:"payer" validate:"required"`
	Points int `json:"points" validate:"required,gt=0"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
}

type SpendingRequest struct {
	Points int `json:"points" validate:"required,gt=0"`
}

type SpendingDetail struct {
	Payer string `json:"payer"`
	Points int `json:"points"`
}

