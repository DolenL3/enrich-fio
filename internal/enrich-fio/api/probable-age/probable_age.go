package probableage

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"enrich-fio/internal/models"
)

const (
	_agifyURL = "https://api.agify.io/"
	_nameKey  = "name"
)

// ProbableAge is a part of service buisness logic, for getting probable age of a person.
type ProbableAge struct {
	client *http.Client
}

// New returns ProbableAge, for getting probable age of a person.
func New(client *http.Client) *ProbableAge {
	return &ProbableAge{
		client: client,
	}
}

// responce is API's responce object.
type responce struct {
	Age int `json:"age"`
}

// Get returns the most likely age for a given person.
func (p *ProbableAge) Get(ctx context.Context, name string, surname string, patronymic string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, _agifyURL, nil)
	if err != nil {
		return 0, errors.Wrap(err, "make request")
	}
	q := req.URL.Query()
	q.Add(_nameKey, name)
	req.URL.RawQuery = q.Encode()

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "send request")
	}
	defer resp.Body.Close()

	r := responce{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return 0, errors.Wrap(err, "decode responce")
	}
	if r.Age == 0 {
		return 0, models.ErrCouldNotEnrich
	}
	return r.Age, nil
}
