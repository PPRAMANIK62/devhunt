package models

import "time"

type Company struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Website     string    `json:"website,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateCompanyRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=100"`
	Slug        string `json:"slug"        validate:"required,min=2,max=100,alphanum"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	Website     string `json:"website"     validate:"omitempty,url"`
}

type UpdateCompanyRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=2,max=100"`
	Slug        *string `json:"slug"        validate:"omitempty,min=2,max=100,alphanum"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Website     *string `json:"website"     validate:"omitempty,url"`
}
