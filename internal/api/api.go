package api

import (
	"context"
	"encoding/json"
	"enrich-fio/internal/models"
	"net/http"

	"github.com/pkg/errors"
)

func enrichWithAge(ctx context.Context, person *models.Person) error {
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, _agifyURL, nil)
	if err != nil {
		return errors.Wrap(err, "make request")
	}
	q := req.URL.Query()
	q.Add(_nameKey, person.Name)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "send request")
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(person)
	if err != nil {
		return errors.Wrap(err, "decode responce")
	}
	return nil
}

// func enrichPerson(ctx context.Context, person *models.Person, ) error {
// 	return nil
// }
