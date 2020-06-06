package handlers

import "gopkg.in/go-playground/validator.v9"

var (
	v = validator.New()
)

//ProductValidator a product validator
type ProductValidator struct {
	validator *validator.Validate
}

//Validate validates a product
func (p *ProductValidator) Validate(i interface{}) error {
	return p.validator.Struct(i)
}

type userValidator struct {
	validator *validator.Validate
}

func (u *userValidator) Validate(i interface{}) error {
	return u.validator.Struct(i)
}
