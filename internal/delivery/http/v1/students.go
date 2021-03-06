package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/zhashkevych/courses-backend/internal/domain"
	"github.com/zhashkevych/courses-backend/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"
)

func (h *Handler) initStudentsRoutes(api *gin.RouterGroup) {
	students := api.Group("/students", h.setSchoolFromRequest)
	{
		students.POST("/sign-up", h.studentSignUp)
		students.POST("/sign-in", h.studentSignIn)
		students.POST("/auth/refresh", h.studentRefresh)
		students.POST("/verify/:hash", h.studentVerify)
		students.GET("/courses", h.studentGetAllCourses)
		students.GET("/courses/:id", h.studentGetCourseById)

		authenticated := students.Group("/", h.userIdentity)
		{
			authenticated.GET("/modules/:id/lessons", h.studentGetModuleLessons)
			authenticated.GET("/modules/:id/offers", h.studentGetModuleOffers)
			authenticated.GET("/promocodes/:code", h.studentGetPromocode)
			authenticated.POST("/order", h.studentCreateOrder)
		}
	}
}

type studentSignUpInput struct {
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email" binding:"required"`
	Password       string `json:"password" binding:"required"`
	RegisterSource string `json:"registerSource"`
}

// @Summary Student SignUp
// @Tags students
// @Description create student account
// @ID studentSignUp
// @Accept  json
// @Produce  json
// @Param input body studentSignUpInput true "sign up info"
// @Success 201 {string} string "ok"
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/sign-up [post]
func (h *Handler) studentSignUp(c *gin.Context) {
	var inp studentSignUpInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	school, err := getSchoolFromContext(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.studentsService.SignUp(c.Request.Context(), service.StudentSignUpInput{
		Name:           inp.Name,
		Email:          inp.Email,
		Password:       inp.Password,
		RegisterSource: inp.RegisterSource,
		SchoolID:       school.ID,
	}); err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusCreated)
}

type studentSignInInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// @Summary Student SignIn
// @Tags students
// @Description student sign in
// @ID studentSignIn
// @Accept  json
// @Produce  json
// @Param input body studentSignInInput true "sign up info"
// @Success 200 {object} tokenResponse
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/sign-in [post]
func (h *Handler) studentSignIn(c *gin.Context) {
	var inp studentSignInInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	school, err := getSchoolFromContext(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := h.studentsService.SignIn(c.Request.Context(), service.StudentSignInInput{
		SchoolID: school.ID,
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}

type refreshInput struct {
	Token string `json:"token" binding:"required"`
}

// @Summary Student Refresh Tokens
// @Tags students
// @Description student refresh tokens
// @Accept  json
// @Produce  json
// @Param input body refreshInput true "sign up info"
// @Success 200 {object} tokenResponse
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/refresh [post]
func (h *Handler) studentRefresh(c *gin.Context) {
	var inp refreshInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	school, err := getSchoolFromContext(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := h.studentsService.RefreshTokens(c.Request.Context(), school.ID, inp.Token)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}

// @Summary Student Verify Registration
// @Tags students
// @Description student verify registration
// @ID studentVerify
// @Accept  json
// @Produce  json
// @Param code path string true "verification code"
// @Success 200 {object} tokenResponse
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/verify/{code} [post]
func (h *Handler) studentVerify(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		newResponse(c, http.StatusBadRequest, "code is empty")
		return
	}

	if err := h.studentsService.Verify(c.Request.Context(), code); err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

// @Summary Student Get All Courses
// @Tags students
// @Description student get all courses
// @ID studentGetAllCourses
// @Accept  json
// @Produce  json
// @Success 200 {array} domain.Course
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/courses [get]
func (h *Handler) studentGetAllCourses(c *gin.Context) {
	school, err := getSchoolFromContext(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Return only published courses
	courses := make([]domain.Course, 0)
	for _, course := range school.Courses {
		if course.Published {
			courses = append(courses, course)
		}
	}

	c.JSON(http.StatusOK, courses)
}

type getCourseByIdResponse struct {
	Course  domain.Course `json:"course"`
	Modules []module      `json:"modules"`
}

type module struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Name     string             `json:"name" bson:"name"`
	Position int                `json:"position" bson:"position"`
	Lessons  []lesson           `json:"lessons" bson:"lessons"`
}

type lesson struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Name     string             `json:"name" bson:"name"`
	Position int                `json:"position" bson:"position"`
}

func newGetCourseByIdResponse(course domain.Course, courseModules []domain.Module) getCourseByIdResponse {
	modules := make([]module, len(courseModules))

	for i := range courseModules {
		modules[i].ID = courseModules[i].ID
		modules[i].Name = courseModules[i].Name
		modules[i].Position = courseModules[i].Position
		modules[i].Lessons = toLessons(courseModules[i].Lessons)
	}

	return getCourseByIdResponse{
		Course:  course,
		Modules: modules,
	}
}

func toLessons(lessons []domain.Lesson) []lesson {
	out := make([]lesson, 0)
	for _, l := range lessons {
		if l.Published {
			out = append(out, lesson{
				ID:       l.ID,
				Name:     l.Name,
				Position: l.Position,
			})
		}
	}
	return out
}

// @Summary Student Get Course By ID
// @Tags students
// @Description student get course by id
// @ID studentGetCourseById
// @Accept  json
// @Produce  json
// @Param id path string true "course id"
// @Success 200 {object} domain.Course
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/courses/{id} [get]
func (h *Handler) studentGetCourseById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		newResponse(c, http.StatusBadRequest, "empty id param")
		return
	}

	school, err := getSchoolFromContext(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var searchedCourse domain.Course
	for _, course := range school.Courses {
		if course.Published && course.ID.Hex() == id {
			searchedCourse = course
		}
	}

	if searchedCourse.ID.IsZero() {
		newResponse(c, http.StatusBadRequest, "not found")
		return
	}

	modules, err := h.coursesService.GetCourseModules(c.Request.Context(), searchedCourse.ID)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, newGetCourseByIdResponse(searchedCourse, modules))
}

type studentGetModuleLessonsResponse struct {
	Lessons []domain.Lesson `json:"lessons"`
}

// @Summary Student Get Lessons By Module ID
// @Security StudentsAuth
// @Tags students
// @Description student get lessons by module id
// @ID studentGetModuleLessons
// @Accept  json
// @Produce  json
// @Param id path string true "module id"
// @Success 200 {object} studentGetModuleLessonsResponse
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/modules/{id}/lessons [get]
func (h *Handler) studentGetModuleLessons(c *gin.Context) {
	moduleIdParam := c.Param("id")
	if moduleIdParam == "" {
		newResponse(c, http.StatusBadRequest, "empty id param")
		return
	}

	moduleId, err := primitive.ObjectIDFromHex(moduleIdParam)
	if err != nil {
		newResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	studentId, err := getStudentId(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	school, err := getSchoolFromContext(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	lessons, err := h.studentsService.GetStudentModuleWithLessons(c.Request.Context(), school.ID, studentId, moduleId)
	if err != nil {
		if err == service.ErrModuleIsNotAvailable {
			newResponse(c, http.StatusForbidden, err.Error())
			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, studentGetModuleLessonsResponse{
		Lessons: lessons,
	})
}

type studentGetModuleOffersResponse struct {
	Offers []studentOffer `json:"offers"`
}

type studentOffer struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	Price       price              `json:"price" bson:"price"`
}

type price struct {
	Value    int    `json:"value"`
	Currency string `json:"currency"`
}

func toStudentOffers(offers []domain.Offer) []studentOffer {
	out := make([]studentOffer, len(offers))

	for i := range offers {
		out[i] = toStudentOffer(offers[i])
	}

	return out
}

func toStudentOffer(offer domain.Offer) studentOffer {
	return studentOffer{
		ID:          offer.ID,
		Name:        offer.Name,
		Description: offer.Description,
		CreatedAt:   offer.CreatedAt,
		Price: price{
			Value:    offer.Price.Value,
			Currency: offer.Price.Currency,
		},
	}
}

// @Summary Student Get Offers By Module ID
// @Security StudentsAuth
// @Tags students
// @Description student get offers by module id
// @ID studentGetModuleOffers
// @Accept  json
// @Produce  json
// @Param id path string true "module id"
// @Success 200 {string} string "ok"
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/modules/{id}/offers [get]
func (h *Handler) studentGetModuleOffers(c *gin.Context) {
	moduleIdParam := c.Param("id")
	if moduleIdParam == "" {
		newResponse(c, http.StatusBadRequest, "empty id param")
		return
	}

	moduleId, err := primitive.ObjectIDFromHex(moduleIdParam)
	if err != nil {
		newResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	school, err := getSchoolFromContext(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	offers, err := h.coursesService.GetModuleOffers(c.Request.Context(), school.ID, moduleId)
	if err != nil {
		if err == service.ErrModuleIsNotAvailable {
			newResponse(c, http.StatusForbidden, err.Error())
			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, studentGetModuleOffersResponse{
		Offers: toStudentOffers(offers),
	})
}

// @Summary Student Get Promocode By Code
// @Security StudentsAuth
// @Tags students
// @Description student get promocode by code
// @ID studentGetPromocode
// @Accept  json
// @Produce  json
// @Param code path string true "code"
// @Success 200 {object} domain.Promocode
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/promocodes/{code} [get]
func (h *Handler) studentGetPromocode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		newResponse(c, http.StatusBadRequest, "empty code param")
		return
	}

	school, err := getSchoolFromContext(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	promocode, err := h.coursesService.GetPromocodeByCode(c.Request.Context(), school.ID, code)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, promocode)
}

type createOrderInput struct {
	OfferId string `json:"offerId"`
	PromoId string `json:"promoId"`
}

type createOrderResponse struct {
	PaymentLink string `json:"paymentLink"`
}

// @Summary Student CreateOrder
// @Security StudentsAuth
// @Tags students
// @Description student create order
// @ID studentCreateOrder
// @Accept  json
// @Produce  json
// @Param input body createOrderInput true "order info"
// @Success 200 {object} createOrderResponse
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/order [post]
func (h *Handler) studentCreateOrder(c *gin.Context) {
	var inp createOrderInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	offerId, err := primitive.ObjectIDFromHex(inp.OfferId)
	if err != nil {
		newResponse(c, http.StatusBadRequest, "invalid offer id")
		return
	}

	var promoId primitive.ObjectID

	if inp.PromoId != "" {
		var err error
		promoId, err = primitive.ObjectIDFromHex(inp.PromoId)
		if err != nil {
			newResponse(c, http.StatusBadRequest, "invalid promo id")
			return
		}
	}

	studentId, err := getStudentId(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	paymentLink, err := h.ordersService.Create(c.Request.Context(), studentId, offerId, promoId)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, createOrderResponse{paymentLink})
}
