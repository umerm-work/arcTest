package db

import (
	"context"
	"fmt"
	"github.com/umerm-work/arcTest/config"
	"github.com/umerm-work/arcTest/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type Repository interface {
	CreateUser(ctx context.Context, in data.User) error
	Login(ctx context.Context, in *data.User) error
	GetUser(ctx context.Context, in *data.User) error
	GetUserByEmail(ctx context.Context, in *data.User) error
	UpdateToken(ctx context.Context, in *data.User) error
	CreateIdea(ctx context.Context, in data.Idea) error
	UpdateIdea(ctx context.Context, in data.Idea) error
	FindIdeas(ctx context.Context, page int64) (idea []*data.Idea, err error)
	FindIdea(ctx context.Context, in *data.Idea) error
	DeleteIdea(ctx context.Context, id string) (err error)
}

type repository struct {
	client         *mongo.Client
	database       *mongo.Database
	userCollection string
	ideaCollection string
}

func New(setting config.Config) Repository {

	opt := options.Client()
	opt.ApplyURI(fmt.Sprintf("mongodb://%v/", setting.DBHost))
	client, err := mongo.NewClient(opt)
	if err != nil {
		log.Fatalf("Client error: %v", err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Connect error: %v", err)
	}
	//defer client.Disconnect(ctx)

	db := client.Database("arc")
	// return  repository
	return &repository{
		client:         client,
		database:       db,
		userCollection: setting.DbUsersCollection,
		ideaCollection: setting.DbIdeasCollection,
	}
}

func (r *repository) CreateUser(ctx context.Context, in data.User) (err error) {

	c := r.database.Collection(r.userCollection)

	res, err := c.InsertOne(ctx, in)
	if err != nil {
		return fmt.Errorf("cannot persist new data with error:%v", err)
	}
	log.Printf("created new data with ID %v", res.InsertedID)
	return nil
}

func (r *repository) Login(ctx context.Context, in *data.User) (err error) {
	c := r.database.Collection(r.userCollection)

	if err = c.FindOne(ctx, bson.M{"email": in.Email, "password": in.Password}).Decode(in); err != nil {
		return fmt.Errorf("cannot find data with error : %v %v", err.Error(), err != nil)
	}
	return nil
}
func (r *repository) GetUser(ctx context.Context, in *data.User) error {

	c := r.database.Collection(r.userCollection)

	if err := c.FindOne(ctx, bson.M{"id": in.ID}).Decode(&in); err != nil {
		return fmt.Errorf("cannot find data with error:%v", err)
	}

	return nil
}
func (r *repository) GetUserByEmail(ctx context.Context, in *data.User) error {

	c := r.database.Collection(r.userCollection)

	if err := c.FindOne(ctx, bson.M{"email": in.Email}).Decode(&in); err != nil {
		return fmt.Errorf("cannot find data with error:%v", err)
	}

	return nil
}

func (r *repository) UpdateToken(ctx context.Context, in *data.User) (err error) {

	c := r.database.Collection(r.userCollection)
	if err = c.FindOne(ctx, bson.M{"id": in.ID}).Decode(&in); err != nil {
		return fmt.Errorf("cannot find data with error:%v", err)
	}
	res, err := c.UpdateOne(ctx,
		bson.M{"id": in.ID},
		bson.M{"$set": &in})

	if err != nil {
		return fmt.Errorf("cannot persist new data with error:%v", err)
	}
	log.Printf("updated data with count %d", res.UpsertedCount)
	return nil
}

func (r *repository) UpdateIdea(ctx context.Context, in data.Idea) (err error) {

	c := r.database.Collection(r.ideaCollection)

	res, err := c.UpdateOne(ctx,
		bson.M{"id": in.ID},
		bson.M{"$set": in})

	if err != nil {
		return fmt.Errorf("cannot persist new data with error:%v", err)
	}
	log.Printf("updated data with count %d", res.UpsertedCount)
	return nil
}

func (r *repository) CreateIdea(ctx context.Context, in data.Idea) error {
	c := r.database.Collection(r.ideaCollection)

	in.CreatedAt = time.Now()
	in.TimeStamp = in.CreatedAt.Unix()
	res, err := c.InsertOne(ctx, in)
	if err != nil {
		return fmt.Errorf("cannot persist new data with error:%v", err)
	}
	log.Printf("created new data with ID %v", res.InsertedID)
	return nil
}

func (r *repository) FindIdeas(ctx context.Context, page int64) (idea []*data.Idea, err error) {

	c := r.database.Collection(r.ideaCollection)

	var count int64

	if count, err = c.CountDocuments(ctx, bson.M{}); err != nil {
		return nil, fmt.Errorf("cannot find data with error:%v", err)
	}
	var limit int64
	limit = 10
	offset := (page - 1) * limit
	if count > 0 {
		var list []*data.Idea

		findOpts := options.FindOptions{}
		findOpts.Limit = &limit
		findOpts.Skip = &offset
		//project := bson.M{
		//	"$project": bson.M{
		//		"id":         1,
		//		"content":    1,
		//		"impact":     1,
		//		"ease":       1,
		//		"confidence": 1,
		//		"created_at": 1,
		//		"average_score": bson.M{
		//			"$divide": []interface{}{
		//				bson.M{"$add": []string{"impact", "ease", "confidence"}}, 3,
		//			},
		//		},
		//	},
		//	"$sort": bson.M{"average_score": -1},
		//}

		log.Printf("limit %v skip %v", limit, offset)
		dbData, err := c.Find(ctx, bson.M{}, &findOpts)
		if err != nil {
			return nil, fmt.Errorf("cannot find data with error:%v", err)
		}

		for dbData.Next(ctx) {
			var ne data.Idea

			// Decode the document
			if err := dbData.Decode(&ne); err != nil {
				return nil, fmt.Errorf("decode error:%v", err)
			}
			list = append(list, &ne)
		}
		log.Printf("Count %v", len(list))
		return list, nil

	}
	return nil, fmt.Errorf("cannot find data")

}

func (r *repository) FindIdea(ctx context.Context, in *data.Idea) error {
	c := r.database.Collection(r.ideaCollection)

	if err := c.FindOne(ctx, bson.M{"id": in.ID}).Decode(&in); err != nil {
		return fmt.Errorf("cannot find data with error:%v", err)
	}

	return nil
}
func (r *repository) DeleteIdea(ctx context.Context, in string) error {
	c := r.database.Collection(r.ideaCollection)

	dr, err := c.DeleteOne(ctx, bson.M{"id": in})
	if err != nil {
		return fmt.Errorf("cannot find data with error:%v", err)
	}
	log.Printf("delete data with count %v", dr.DeletedCount)
	return nil
}
