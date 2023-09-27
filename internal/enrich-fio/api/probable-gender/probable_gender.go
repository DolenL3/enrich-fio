package probablegender

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"enrich-fio/internal/models"
)

const (
	_genderizeURL = "https://api.genderize.io/"
	_nameKey      = "name"
)

// ProbableGender is a part of service buisness logic, for getting probable gender of a person.
type ProbableGender struct {
	client *http.Client
}

// New returns ProbableGender, for getting probable gedner of a person.
func New(client *http.Client) *ProbableGender {
	return &ProbableGender{
		client: client,
	}
}

// responce is API's responce object.
type responce struct {
	Gender string `json:"gender"`
}

// Get returns the most likely gender for a given person.
func (p *ProbableGender) Get(ctx context.Context, name string, surname string, patronymic string) (models.Gender, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, _genderizeURL, nil)
	if err != nil {
		return "", errors.Wrap(err, "make request")
	}
	q := req.URL.Query()
	q.Add(_nameKey, name)
	req.URL.RawQuery = q.Encode()

	resp, err := p.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "send request")
	}
	defer resp.Body.Close()

	r := responce{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return "", errors.Wrap(err, "decode responce")
	}
	switch r.Gender {
	case "male":
		return models.GenderMale, nil
	case "female":
		return models.GenderFemale, nil
	}
	return "", models.ErrCouldNotEnrich
}
