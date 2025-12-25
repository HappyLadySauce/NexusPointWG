package model

import (
	customvalidator "github.com/HappyLadySauce/NexusPointWG/pkg/utils/validator"
	"github.com/go-playground/validator/v10"
	"github.com/marmotedu/component-base/pkg/validation/field"
)

var (
	// modelValidator is a shared validator instance with custom validators registered
	modelValidator *validator.Validate
)

func init() {
	// Create a new validator instance and register custom validators
	modelValidator = validator.New()
	if err := customvalidator.RegisterCustomValidators(modelValidator); err != nil {
		panic("Failed to register custom validators for model validation: " + err.Error())
	}
}

func (u *User) Validate() field.ErrorList {
	var allErrs field.ErrorList

	// Use the custom validator instance which has all custom validators registered
	if err := modelValidator.Struct(u); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, validationError := range validationErrors {
				allErrs = append(allErrs, field.Invalid(
					field.NewPath(validationError.Field()),
					validationError.Value(),
					validationError.Tag(),
				))
			}
		}
	}

	return allErrs
}

func (p *WGPeer) Validate() field.ErrorList {
	var allErrs field.ErrorList
	if err := modelValidator.Struct(p); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, validationError := range validationErrors {
				allErrs = append(allErrs, field.Invalid(
					field.NewPath(validationError.Field()),
					validationError.Value(),
					validationError.Tag(),
				))
			}
		}
	}
	return allErrs
}
