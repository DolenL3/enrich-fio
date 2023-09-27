package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"enrich-fio/internal/config"
	enrichfio "enrich-fio/internal/enrich-fio"
	"enrich-fio/internal/models"
)

// requestPOST is a structure of expected POST request.
type requestPOST struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic"`
}

// requestPUT is a structure of expected PUT request.
type requestPUT struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Surname     string        `json:"surname"`
	Patronymic  string        `json:"patronymic"`
	Age         int           `json:"age"`
	Gender      models.Gender `json:"gender"`
	Nationality string        `json:"nationality"`
}

// HTTPHandler is http request handler.
type HTTPHandler struct {
	router  *gin.Engine
	service *enrichfio.Service
	config  *config.RestConfig
}

// NewHTTPHandler returns HTTPHandler.
func NewHTTPHandler(router *gin.Engine, service *enrichfio.Service, config *config.RestConfig) *HTTPHandler {
	return &HTTPHandler{
		router:  router,
		service: service,
		config:  config,
	}
}

// Start starts http handler.
func (h *HTTPHandler) Start() error {
	h.router.GET("/people", h.getPeople)
	h.router.GET("people/:id", h.getPerson)
	h.router.POST("/people", h.addPerson)
	h.router.DELETE("/people/:id", h.deletePerson)
	h.router.PUT("/people/:id", h.changePerson)
	logger := zap.L()
	logger.Info(fmt.Sprintf("http server is up and running on %s", h.config.Host))
	err := h.router.Run(h.config.Host)
	if err != nil {
		return errors.Wrap(err, "run router")
	}
	return nil
}

// getPerson gets a single person by id.
func (h *HTTPHandler) getPerson(c *gin.Context) {
	idURL := c.Param("id")
	if idURL != "" {
		id, err := uuid.Parse(idURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		person, err := h.service.Storage.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, person)
		return
	}
}

// getPeople gets people list with filters, described in URL query.
// localhost:8080/people | localhost:8080/people?name=Name&age=min:max
func (h *HTTPHandler) getPeople(c *gin.Context) {
	idQuery := c.Request.URL.Query().Get("id")
	page, err := strconv.Atoi(c.Request.URL.Query().Get("page"))
	if err != nil {
		page = 0
	}
	var id uuid.UUID
	if idQuery != "" {
		parsedID, err := uuid.Parse(idQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id = parsedID
	}
	name := c.Request.URL.Query().Get("name")
	surname := c.Request.URL.Query().Get("surname")
	patronymic := c.Request.URL.Query().Get("patronymic")

	age := c.Request.URL.Query().Get("age")
	ageFilter := models.FilterAge{}
	if age != "" {
		if strings.Index(age, ":") != -1 {
			split := strings.Split(age, ":")
			ageMin, err := strconv.Atoi(split[0])
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ageFilter.Min = ageMin
			ageMax, err := strconv.Atoi(split[1])
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ageFilter.Max = ageMax
		} else {
			ageEqual, err := strconv.Atoi(age)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ageFilter.Min = ageEqual
			ageFilter.Max = ageEqual
		}
	}
	genderQuery := c.Request.URL.Query().Get("gender")
	var gender models.Gender
	if genderQuery != "" {
		switch genderQuery {
		case "male":
			gender = models.GenderMale
		case "female":
			gender = models.GenderFemale
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "unrecognized gender query parameter: " + genderQuery})
			return
		}
	}
	nationality := c.Request.URL.Query().Get("nationality")

	filter := models.FilterConfig{
		ID:          id,
		Name:        name,
		Surname:     surname,
		Patronymic:  patronymic,
		Age:         ageFilter,
		Gender:      gender,
		Nationality: nationality,
	}

	people, err := h.service.Storage.GetWithFilter(c.Request.Context(), filter, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, people)
}

// addPerson adds a new person with name, surname, patronymic from request's body.
func (h *HTTPHandler) addPerson(c *gin.Context) {
	person := requestPOST{}
	err := c.ShouldBind(&person)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if person.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	if person.Surname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "surname required"})
		return
	}
	err = h.service.AddPerson(c.Request.Context(), person.Name, person.Surname, person.Patronymic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

}

// deletePerson deletes person by id from URL parameters.
func (h *HTTPHandler) deletePerson(c *gin.Context) {
	idURL := c.Param("id")
	if idURL != "" {
		id, err := uuid.Parse(idURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = h.service.Storage.DeleteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "no id parameter found in URL"})
	return
}

// changePerson changes person's data with data from request's body.
// WARNING: Doesn't support deleting data.
func (h *HTTPHandler) changePerson(c *gin.Context) {
	idURL := c.Param("id")
	if idURL != "" {
		id, err := uuid.Parse(idURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		request := requestPUT{}
		err = c.ShouldBind(&request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		changes := models.ChangeConfig{
			ID:          request.ID,
			Name:        request.Name,
			Surname:     request.Surname,
			Patronymic:  request.Patronymic,
			Age:         request.Age,
			Gender:      request.Gender,
			Nationality: request.Nationality,
		}
		err = h.service.Storage.ChangeByID(c.Request.Context(), id, changes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "no id parameter found in URL"})
	return
}
