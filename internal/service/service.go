package service

import (
	"context"
	"github.com/zhashkevych/courses-backend/internal/domain"
	"github.com/zhashkevych/courses-backend/internal/repository"
	"github.com/zhashkevych/courses-backend/pkg/auth"
	"github.com/zhashkevych/courses-backend/pkg/cache"
	"github.com/zhashkevych/courses-backend/pkg/email"
	"github.com/zhashkevych/courses-backend/pkg/hash"
	"github.com/zhashkevych/courses-backend/pkg/payment"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Schools interface {
	GetByDomain(ctx context.Context, domainName string) (domain.School, error)
}

type StudentSignUpInput struct {
	Name           string
	Email          string
	Password       string
	SchoolID       primitive.ObjectID
	RegisterSource string
}

type StudentSignInInput struct {
	Email    string
	Password string
	SchoolID primitive.ObjectID
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type Students interface {
	SignUp(ctx context.Context, input StudentSignUpInput) error
	SignIn(ctx context.Context, input StudentSignInInput) (Tokens, error)
	RefreshTokens(ctx context.Context, schoolId primitive.ObjectID, refreshToken string) (Tokens, error)
	Verify(ctx context.Context, hash string) error
	GetStudentModuleWithLessons(ctx context.Context, schoolId, studentId, moduleId primitive.ObjectID) ([]domain.Lesson, error)
	GiveAccessToModules(ctx context.Context, studentId primitive.ObjectID, moduleIds []primitive.ObjectID) error
	GiveAccessToPackages(ctx context.Context, studentId primitive.ObjectID, packageIds []primitive.ObjectID) error
}

type AddToListInput struct {
	Email            string
	Name             string
	RegisterSource   string
	VerificationCode string
}

type Emails interface {
	AddToList(AddToListInput) error
}

type Courses interface {
	GetCourseModules(ctx context.Context, courseId primitive.ObjectID) ([]domain.Module, error)
	GetModule(ctx context.Context, moduleId primitive.ObjectID) (domain.Module, error)
	GetModuleWithContent(ctx context.Context, moduleId primitive.ObjectID) (domain.Module, error)

	GetModuleOffers(ctx context.Context, schoolId, moduleId primitive.ObjectID) ([]domain.Offer, error)

	GetPackageOffers(ctx context.Context, schoolId, packageId primitive.ObjectID) ([]domain.Offer, error)
	GetPackagesModules(ctx context.Context, packageIds []primitive.ObjectID) ([]domain.Module, error)

	GetPromocodeByCode(ctx context.Context, schoolId primitive.ObjectID, code string) (domain.Promocode, error)
	GetPromocodeById(ctx context.Context, id primitive.ObjectID) (domain.Promocode, error)

	GetOfferById(ctx context.Context, id primitive.ObjectID) (domain.Offer, error)
}

type Orders interface {
	Create(ctx context.Context, studentId, offerId, promocodeId primitive.ObjectID) (string, error)
	AddTransaction(ctx context.Context, id primitive.ObjectID, transaction domain.Transaction) (domain.Order, error)
}

type Payments interface {
	ProcessTransaction(ctx context.Context, callbackData payment.Callback) error
}

type Services struct {
	Schools  Schools
	Students Students
	Courses  Courses
	Payments Payments
	Orders   Orders
}

type ServicesDeps struct {
	Repos              *repository.Repositories
	Cache              cache.Cache
	Hasher             hash.PasswordHasher
	TokenManager       auth.TokenManager
	EmailProvider      email.Provider
	EmailListId        string
	PaymentProvider    payment.FondyProvider
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	PaymentCallbackURL string
	PaymentResponseURL string
}

func NewServices(deps ServicesDeps) *Services {
	emailsService := NewEmailsService(deps.EmailProvider, deps.EmailListId)
	coursesService := NewCoursesService(deps.Repos.Courses, deps.Repos.Offers, deps.Repos.Promocodes)
	ordersService := NewOrdersService(deps.Repos.Orders, coursesService, deps.PaymentProvider, deps.PaymentCallbackURL, deps.PaymentResponseURL)
	studentsService := NewStudentsService(deps.Repos.Students, coursesService, deps.Hasher,
		deps.TokenManager, emailsService, deps.AccessTokenTTL, deps.RefreshTokenTTL)

	return &Services{
		Schools:  NewSchoolsService(deps.Repos.Schools, deps.Cache),
		Students: studentsService,
		Courses:  coursesService,
		Payments: NewPaymentsService(deps.PaymentProvider, ordersService, coursesService, studentsService),
		Orders:   ordersService,
	}
}
