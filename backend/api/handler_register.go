package api

import (
	"net/http"
	"time"

	"brainbook-api/internal/request"
	"brainbook-api/internal/security"
	"brainbook-api/internal/validator"
)

func (app *Application) createUser(w http.ResponseWriter, r *http.Request) {
	// Check if user is already authenticated
	if user := contextGetAuthenticatedUser(r); user != nil {
		app.Conflict(w, r)
		return
	}

	var input struct {
		FName     string              `json:"f_name"`
		LName     string              `json:"l_name"`
		Email     string              `json:"email"`
		Password  string              `json:"password"`
		DOB       time.Time           `json:"dob"`
		Avatar    []byte              `json:"avatar"`
		Nickname  string              `json:"nickname"`
		Bio       string              `json:"bio"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	_, emailFound, err := app.DB.UserByEmail(input.Email)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// First Name validation
	input.Validator.CheckField(validator.NotBlank(input.FName), "first-name", "First name is required")
	input.Validator.CheckField(validator.MinRunes(input.FName, 2), "first-name", "First name must be at least 2 characters")
	input.Validator.CheckField(validator.MaxRunes(input.FName, 50), "first-name", "First name limit exceeded (50 characters)")

	// Last Name validation
	input.Validator.CheckField(validator.NotBlank(input.LName), "last-name", "Last name is required")
	input.Validator.CheckField(validator.MinRunes(input.LName, 2), "last-name", "Last name must be at least 2 characters")
	input.Validator.CheckField(validator.MaxRunes(input.LName, 50), "last-name", "Last name limit exceeded (50 characters)")

	// Nickname validation
	input.Validator.CheckField(validator.MaxRunes(input.Nickname, 30), "nickname", "Nickname limit exceeded (30 characters)")

	//Bio validation
	input.Validator.CheckField(validator.MaxRunes(input.Bio, 500), "bio", "Bio limit exceeded (500 characters)")

	//Avatar validation
	input.Validator.CheckField(len(input.Avatar) <= 5_000_000, "avatar", "Avatar size limit exceeded (5MB)")
	if len(input.Avatar) > 0 {
		input.Validator.CheckField(isAllowedImage(input.Avatar), "avatar", "Avatar must be JPEG, PNG, or GIF")
	}

	// Email validation
	input.Validator.CheckField(validator.NotBlank(input.Email), "email", "Email is required")
	input.Validator.CheckField(validator.IsEmail(input.Email), "email", "Must be a valid email address")
	input.Validator.CheckField(!emailFound, "email", "Email is already in use")

	// Password validation
	input.Validator.CheckField(validator.NotBlank(input.Password), "password", "Password is required")
	input.Validator.CheckField(validator.MinRunes(input.Password, 8), "password", "Password must be at least 8 characters")
	input.Validator.CheckField(validator.MaxRunes(input.Password, 72), "password", "Password must be no more than 72 characters")
	input.Validator.CheckField(validator.NotIn(input.Password, security.CommonPasswords...), "password", "Password is too common")

	// DOB validation
	input.Validator.CheckField(validator.ValidDOB(input.DOB, 13, 120), "age", "Age must be between 13 and 120.")

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	hashedPassword, err := security.Hash(input.Password)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	_, err = app.DB.InsertUser(input.FName, input.LName, input.Email, hashedPassword, input.Nickname, input.Bio, input.DOB, input.Avatar)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
