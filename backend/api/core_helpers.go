package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (app *Application) backgroundTask(r *http.Request, fn func() error) {
	app.WG.Add(1)

	go func() {
		defer app.WG.Done()

		defer func() {
			err := recover()
			if err != nil {
				app.reportServerError(r, fmt.Errorf("%s", err))
			}
		}()

		err := fn()
		if err != nil {
			app.reportServerError(r, err)
		}
	}()
}

func (app *Application) parseStringID(stringID string) (int, error) {

	sanitizedID := strings.TrimSpace(stringID)

	if sanitizedID == "" {
		return 0, fmt.Errorf("ID is required")
	}

	ID, err := strconv.Atoi(sanitizedID)
	if err != nil {
		return 0, fmt.Errorf("Invalid ID: %s", sanitizedID)
	}

	return ID, nil
}
