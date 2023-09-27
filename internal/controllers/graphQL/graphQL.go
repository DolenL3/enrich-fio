package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"enrich-fio/internal/config"
	enrichfio "enrich-fio/internal/enrich-fio"
	"enrich-fio/internal/models"
)

// GraphQLHandler is a mess...
type GraphQLHandler struct {
	service *enrichfio.Service
	config  *config.GraphQLConfig
}

// NewGraphQLHandler returns GraphQLHandler.
func NewGraphQLHandler(service *enrichfio.Service, config *config.GraphQLConfig) *GraphQLHandler {
	return &GraphQLHandler{
		service: service,
		config:  config,
	}
}

// Start starts GraphQL handler.
func (h *GraphQLHandler) Start(ctx context.Context) error {
	logger := zap.L()
	schema, err := h.createSchema(ctx)
	if err != nil {
		return errors.Wrap(err, "creating graphQL schema")
	}

	http.HandleFunc("/person", func(w http.ResponseWriter, r *http.Request) {
		result := executeQuery(r.URL.Query().Get("query"), schema)
		json.NewEncoder(w).Encode(result)
	})

	logger.Info(fmt.Sprintf("GraphQL is up and running on %s", h.config.Host))
	http.ListenAndServe(h.config.Host, nil)

	return nil
}

// executeQuery executes GraphQL query.
func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("errors: %v", result.Errors)
	}
	return result
}

// createSchema creates GraphQL schema.
func (h *GraphQLHandler) createSchema(ctx context.Context) (graphql.Schema, error) {
	var personType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Person",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.String,
				},
				"name": &graphql.Field{
					Type: graphql.String,
				},
				"surname": &graphql.Field{
					Type: graphql.String,
				},
				"patronymic": &graphql.Field{
					Type: graphql.String,
				},
				"age": &graphql.Field{
					Type: graphql.Int,
				},
				"gender": &graphql.Field{
					Type: graphql.String,
				},
				"nationality": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)

	var queryType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				/* Get (read) single person by id
				   http://localhost:4000/person?query={person(id:"id"){name, nationality, age}}
				*/
				"person": &graphql.Field{
					Type:        personType,
					Description: "Get person by id",
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						id, ok := p.Args["id"].(string)
						if ok {
							uuid, err := uuid.Parse(id)
							if err != nil {
								return models.Person{}, errors.Wrap(err, "parsing id into uuid")
							}
							person, err := h.service.Storage.GetByID(ctx, uuid)
							if err != nil {
								return models.Person{}, errors.Wrap(err, "get person by id")
							}
							return person, nil
						}
						return nil, nil
					},
				},
				/* Get (read) person list
				http://localhost:4000/person?query={list(page:page){id,name,surname, patronymic, age}}
				*/
				"list": &graphql.Field{
					Type:        graphql.NewList(personType),
					Description: "Get All people",
					Args: graphql.FieldConfigArgument{
						"page": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
					},
					Resolve: func(params graphql.ResolveParams) (interface{}, error) {
						page := params.Args["page"].(int)
						people, err := h.service.Storage.GetWithFilter(ctx, models.FilterConfig{}, page)
						if err != nil {
							return models.Person{}, errors.Wrap(err, "GetWithFilter")
						}
						return people, nil
					},
				},
				/* Get (read) person list with filter
				   http://localhost:4000/person?query={filter{page, name,gender, ageMin, ageMax}{id, age}}
				*/
				// Library doesn't support operators, I don't have time to rewrite or think of anything, so age filter is ugly.
				"filter": &graphql.Field{
					Type:        graphql.NewList(personType),
					Description: "Get people, that match the given filter",
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"name": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"surname": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"patronymic": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"age": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
						"ageMin": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
						"ageMax": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
						"gender": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"nationality": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"page": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
					},
					Resolve: func(params graphql.ResolveParams) (interface{}, error) {
						page := params.Args["page"].(int)
						filter := models.FilterConfig{}
						id, idOK := params.Args["id"].(string)
						if idOK {
							uuid, err := uuid.Parse(id)
							if err != nil {
								return models.Person{}, errors.Wrap(err, "parsing id into uuid")
							}
							filter.ID = uuid
						}
						name, nameOK := params.Args["name"].(string)
						if nameOK {
							filter.Name = name
						}
						surname, surnameOK := params.Args["surname"].(string)
						if surnameOK {
							filter.Surname = surname
						}
						patronymic, patronymicOK := params.Args["patronymic"].(string)
						if patronymicOK {
							filter.Patronymic = patronymic
						}
						age, ageOK := params.Args["age"].(int)
						if ageOK {
							filter.Age.Min = age
							filter.Age.Max = age
						}
						ageMin, AgeMinOk := params.Args["ageMin"].(int)
						if AgeMinOk {
							filter.Age.Min = ageMin
						}
						ageMax, AgeMaxOk := params.Args["ageMax"].(int)
						if AgeMaxOk {
							filter.Age.Max = ageMax
						}
						gender, genderOK := params.Args["gender"].(string)
						if genderOK {
							switch gender {
							case "male":
								filter.Gender = models.GenderMale
							case "female":
								filter.Gender = models.GenderFemale
							}
						}
						nationality, nationalityOK := params.Args["nationality"].(string)
						if nationalityOK {
							filter.Nationality = nationality
						}
						people, err := h.service.Storage.GetWithFilter(ctx, filter, page)
						if err != nil {
							return models.Person{}, errors.Wrap(err, "GetWithFilter")
						}
						return people, nil
					},
				},
			},
		})

	var mutationType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			/* Create new person
			http://localhost:4000/person?query=mutation{create(name:"Name",surname:"Surname",patronymic:"Patronymic"){name,surname,patronymic}}
			*/
			"create": &graphql.Field{
				Type:        personType,
				Description: "Create new person",
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"surname": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"patronymic": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					err := h.service.AddPerson(ctx, params.Args["name"].(string),
						params.Args["surname"].(string), params.Args["patronymic"].(string))
					if err != nil {
						return models.Person{}, errors.Wrap(err, "add person to storage")
					}
					return models.Person{Name: params.Args["name"].(string),
						Surname: params.Args["surname"].(string), Patronymic: params.Args["patronymic"].(string)}, nil
				},
			},

			/* Update person by id
			http://localhost:4000/person?query=mutation{update(id:"id",age:69){id,age}}
			*/
			"update": &graphql.Field{
				Type:        personType,
				Description: "Update person by id",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"newID": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"name": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"surname": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"patronymic": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"age": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"gender": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"nationality": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id, _ := params.Args["id"].(string)
					currentUUID, err := uuid.Parse(id)
					if err != nil {
						return models.Person{}, errors.Wrap(err, "parsing id into uuid")
					}
					changes := models.ChangeConfig{}
					newID, newIDOk := params.Args["newID"].(string)
					if newIDOk {
						newUUID, err := uuid.Parse(newID)
						if err != nil {
							return models.Person{}, errors.Wrap(err, "parsing newID into newUUID")
						}
						changes.ID = newUUID
					}
					name, nameOk := params.Args["name"].(string)
					if nameOk {
						changes.Name = name
					}
					surname, surnameOk := params.Args["surname"].(string)
					if surnameOk {
						changes.Surname = surname
					}
					patronymic, patronymicOk := params.Args["patronymic"].(string)
					if patronymicOk {
						changes.Patronymic = patronymic
					}
					age, ageOk := params.Args["age"].(int)
					if ageOk {
						changes.Age = age
					}
					gender, genderOk := params.Args["gender"].(string)
					if genderOk {
						switch gender {
						case "male":
							changes.Gender = models.GenderMale
						case "female":
							changes.Gender = models.GenderFemale
						}
					}
					nationality, nationalityOk := params.Args["nationality"].(string)
					if nationalityOk {
						changes.Nationality = nationality
					}
					err = h.service.Storage.ChangeByID(ctx, currentUUID, changes)
					if err != nil {
						return models.Person{}, errors.Wrap(err, "changing person by id")
					}
					return changes, nil
				},
			},

			/* Delete person by id
			   http://localhost:4000/person?query=mutation{delete(id:"id"){id}}
			*/
			"delete": &graphql.Field{
				Type:        personType,
				Description: "Delete person by id",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id, _ := params.Args["id"].(string)
					uuid, err := uuid.Parse(id)
					if err != nil {
						return models.Person{}, errors.Wrap(err, "parsing id into uuid")
					}
					err = h.service.Storage.DeleteByID(ctx, uuid)
					if err != nil {
						return models.Person{}, errors.Wrap(err, "deleting person by id")
					}
					return models.Person{ID: uuid}, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    queryType,
			Mutation: mutationType,
		},
	)
	if err != nil {
		return graphql.Schema{}, err
	}
	return schema, nil
}
