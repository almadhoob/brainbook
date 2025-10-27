package validator

import "time"

type Validator struct {
	Errors      []string          `json:",omitempty"`
	FieldErrors map[string]string `json:",omitempty"`
}

func (v Validator) HasErrors() bool {
	return len(v.Errors) != 0 || len(v.FieldErrors) != 0
}

func (v *Validator) AddError(message string) {
	if v.Errors == nil {
		v.Errors = []string{}
	}

	v.Errors = append(v.Errors, message)
}

func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = map[string]string{}
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) Check(ok bool, message string) {
	if !ok {
		v.AddError(message)
	}
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func AgeFromDOB(dob, ref time.Time) int {
	if dob.IsZero() || !dob.Before(ref) {
		return 0
	}
	age := ref.Year() - dob.Year()
	if ref.Month() < dob.Month() || (ref.Month() == dob.Month() && ref.Day() < dob.Day()) {
		age--
	}
	return age
}

// ValidDOB checks presence, not-in-future, and age range [minYears, maxYears].
func ValidDOB(dob time.Time, minYears, maxYears int) bool {
	if dob.IsZero() {
		return false
	}
	now := time.Now()
	if !dob.Before(now) {
		return false
	}
	age := AgeFromDOB(dob, now)
	return age >= minYears && age <= maxYears
}
