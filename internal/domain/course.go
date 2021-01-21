package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type CourseEntity struct {
	Name      string `json:"name" bson:"name"`
	Position  int    `json:"position" bson:"position"`
	Published bool   `json:"published"`
}

type Course struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	ImageURL    string             `json:"imageUrl" bson:"imageUrl"`
	CreatedAt   int64              `json:"createdAt" bson:"createdAt"`
	UpdatedAt   int64              `json:"updatedAt" bson:"updatedAt"`
	Published   bool               `json:"published" bson:"published"`
}

type Module struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	CourseID  primitive.ObjectID `json:"courseId" bson:"courseId"`
	PackageID primitive.ObjectID `json:"packageId" bson:"packageId"`
	Lessons   []Lesson           `json:"lessons" bson:"lessons"`
	CourseEntity
}

type Lesson struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`
	CourseEntity
}

type CourseContent struct {
	LessonID primitive.ObjectID `json:"lessonId" bson:"lessonId"`
	Content  string             `json:"content" bson:"content"`
}

type CoursePackages struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	CourseID    primitive.ObjectID `json:"courseId" bson:"courseId"`
}