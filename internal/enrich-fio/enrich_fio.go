package enrichfio

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"enrich-fio/internal/models"
)

// Service is a service with buisness logic.
type Service struct {
	Storage             Storage
	ProbableAge         ProbableAge
	ProbableGender      ProbableGender
	ProbableNationality ProbableNationality
}

// New returns Service service.
func New(storage Storage, probableAge ProbableAge, probableGender ProbableGender, probableNationality ProbableNationality) *Service {
	return &Service{
		Storage:             storage,
		ProbableAge:         probableAge,
		ProbableGender:      probableGender,
		ProbableNationality: probableNationality,
	}
}

func (s *Service) AddPerson(ctx context.Context, name string, surname string, patronymic string) error {
	person, err := s.enrich(ctx, name, surname, patronymic)
	if err != nil {
		return errors.Wrap(err, "enrich")
	}

	err = s.Storage.Save(ctx, person)
	if err != nil {
		return errors.Wrap(err, "save person in storage")
	}
	return nil
}

func (s *Service) enrich(ctx context.Context, name string, surname string, patronymic string) (models.Person, error) {
	gender, err := s.ProbableGender.Get(ctx, name, surname, patronymic)
	if err != nil {
		return models.Person{}, errors.Wrap(err, "get probable gender")
	}

	age, err := s.ProbableAge.Get(ctx, name, surname, patronymic)
	if err != nil {
		return models.Person{}, errors.Wrap(err, "get probable gender")
	}

	nationality, err := s.ProbableNationality.Get(ctx, name, surname, patronymic)
	if err != nil {
		return models.Person{}, errors.Wrap(err, "get probable nationality")
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return models.Person{}, errors.Wrap(err, "generate id random")
	}

	person := models.Person{
		ID:          id,
		Name:        name,
		Surname:     surname,
		Patronymic:  patronymic,
		Gender:      gender,
		Age:         age,
		Nationality: nationality,
	}

	return person, nil
}
