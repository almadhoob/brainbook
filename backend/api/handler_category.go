package api

import (
	"net/http"

	"brainbook-api/internal/response"
)

func (app *Application) fetchCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := app.DB.GetAllCategories()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = response.JSON(w, http.StatusOK, categories)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

}
