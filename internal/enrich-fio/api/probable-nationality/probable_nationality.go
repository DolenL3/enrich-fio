package probablenationality

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"enrich-fio/internal/models"
)

const (
	_nameKey        = "name"
	_nationalizeURL = "https://api.nationalize.io/"
)

// ProbableNationality is a part of service buisness logic, for getting probable nationality of a person.
type ProbableNationality struct {
	client *http.Client
}

// New returns ProbableNationality, for getting probable nationality of a person.
func New(client *http.Client) *ProbableNationality {
	return &ProbableNationality{
		client: client,
	}
}

// responce is API's responce object.
type responce struct {
	Country []country `json:"country"`
}

// country is a segment of API's responce object.
type country struct {
	CountryID   string  `json:"country_id"`
	Probability float64 `json:"probability"`
}

// Get returns the most likely nationality for a given person
func (p *ProbableNationality) Get(ctx context.Context, name string, surname string, patronymic string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, _nationalizeURL, nil)
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

	responce := responce{}
	err = json.NewDecoder(resp.Body).Decode(&responce)
	if err != nil {
		return "", errors.Wrap(err, "decode responce")
	}
	nationality := ""
	probability := 0.0
	for _, country := range responce.Country {
		if country.Probability > probability {
			probability = country.Probability
			nationality = country.CountryID
		}
	}
	if nationality == "" {
		return "", models.ErrCouldNotEnrich
	}
	return nationality, nil
}
