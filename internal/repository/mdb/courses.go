package mdb

import (
	"context"
	"github.com/zhashkevych/courses-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CoursesRepo struct {
	db *mongo.Database
}

func NewCoursesRepo(db *mongo.Database) *CoursesRepo {
	return &CoursesRepo{db: db}
}

func (r *CoursesRepo) GetModules(ctx context.Context, courseId primitive.ObjectID) ([]domain.Module, error) {
	var modules []domain.Module
	cur, err := r.db.Collection(modulesCollection).Find(ctx, bson.M{"courseId": courseId, "published": true})
	if err != nil {
		return nil, err
	}

	err = cur.All(ctx, &modules)
	return modules, err
}

func (r *CoursesRepo) GetModuleWithContent(ctx context.Context, moduleId primitive.ObjectID) (domain.Module, error) {
	var module domain.Module
	err := r.db.Collection(modulesCollection).FindOne(ctx, bson.M{"_id": moduleId, "published": true}).Decode(&module)
	if err != nil {
		return module, err
	}

	lessonIds := make([]primitive.ObjectID, len(module.Lessons))
	for _, lesson := range module.Lessons {
		lessonIds = append(lessonIds, lesson.ID)
	}

	var content []domain.LessonContent
	cur, err := r.db.Collection(contentCollection).Find(ctx, bson.M{"lessonId": bson.M{"$in": lessonIds}})
	if err != nil {
		return module, err
	}

	err = cur.All(ctx, &content)
	if err != nil {
		return module, err
	}

	for i := range module.Lessons {
		for _, lessonContent := range content {
			if module.Lessons[i].ID == lessonContent.LessonID {
				module.Lessons[i].Content = lessonContent.Content
			}
		}
	}

	return module, nil
}

func (r *CoursesRepo) GetModule(ctx context.Context, moduleId primitive.ObjectID) (domain.Module, error) {
	var module domain.Module
	err := r.db.Collection(modulesCollection).FindOne(ctx, bson.M{"_id": moduleId, "published": true}).Decode(&module)
	return module, err
}

func (r *CoursesRepo) GetPackagesModules(ctx context.Context, packageIds []primitive.ObjectID) ([]domain.Module, error) {
	var modules []domain.Module
	cur, err := r.db.Collection(modulesCollection).Find(ctx, bson.M{"packageId": bson.M{"$in": packageIds}})
	if err != nil {
		return nil, err
	}

	err = cur.All(ctx, &modules)
	return modules, err
}
