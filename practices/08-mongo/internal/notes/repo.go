package notes

import (
	"context"
	"errors"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNotFound  = errors.New("note not found")
	ErrDuplicate = errors.New("note with this title already exists")
)

type Repo struct {
	col *mongo.Collection
}

// NewRepo возвращает репозиторий коллекции notes и создаёт уникальный индекс по title.
func NewRepo(ctx context.Context, db *mongo.Database) (*Repo, error) {
	col := db.Collection("notes")
	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}
	return &Repo{col: col}, nil
}

func (r *Repo) Create(ctx context.Context, title, content string) (Note, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	n := Note{Title: title, Content: content, CreatedAt: now, UpdatedAt: now}
	res, err := r.col.InsertOne(ctx, n)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return Note{}, ErrDuplicate
		}
		return Note{}, err
	}
	n.ID = res.InsertedID.(primitive.ObjectID)
	return n, nil
}

func (r *Repo) ByID(ctx context.Context, idHex string) (Note, error) {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return Note{}, ErrNotFound
	}
	var n Note
	if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&n); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Note{}, ErrNotFound
		}
		return Note{}, err
	}
	return n, nil
}

func (r *Repo) List(ctx context.Context, p ListParams) ([]Note, error) {
	filter := bson.M{}
	if p.Q != "" {
		// QuoteMeta: поисковая строка пользователя не должна быть регулярным выражением.
		filter["title"] = bson.M{"$regex": regexp.QuoteMeta(p.Q), "$options": "i"}
	}
	opts := options.Find().SetLimit(p.Limit)
	if p.After != "" {
		oid, err := primitive.ObjectIDFromHex(p.After)
		if err != nil {
			return nil, ErrNotFound
		}
		filter["_id"] = bson.M{"$lt": oid}
		opts.SetSort(bson.D{{Key: "_id", Value: -1}})
	} else {
		opts.SetSkip(p.Skip).SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}})
	}
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []Note{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Update частично обновляет заметку: nil-поля не изменяются.
func (r *Repo) Update(ctx context.Context, idHex string, title, content *string) (Note, error) {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return Note{}, ErrNotFound
	}
	set := bson.M{"updatedAt": time.Now().UTC().Truncate(time.Millisecond)}
	if title != nil {
		set["title"] = *title
	}
	if content != nil {
		set["content"] = *content
	}
	after := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated Note
	err = r.col.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": set}, after).Decode(&updated)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Note{}, ErrNotFound
		}
		if mongo.IsDuplicateKeyError(err) {
			return Note{}, ErrDuplicate
		}
		return Note{}, err
	}
	return updated, nil
}

func (r *Repo) Delete(ctx context.Context, idHex string) error {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return ErrNotFound
	}
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// Stats считает количество заметок и среднюю длину content через aggregation pipeline.
func (r *Repo) Stats(ctx context.Context) (Stats, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "avgContentLength", Value: bson.D{{Key: "$avg", Value: bson.D{{Key: "$strLenCP", Value: "$content"}}}}},
		}}},
	}
	cur, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return Stats{}, err
	}
	defer cur.Close(ctx)
	var res []Stats
	if err := cur.All(ctx, &res); err != nil {
		return Stats{}, err
	}
	if len(res) == 0 {
		return Stats{}, nil
	}
	return res[0], nil
}

var _ Store = (*Repo)(nil)

// Store описывает операции, нужные HTTP-слою; позволяет тестировать обработчики без MongoDB.
type Store interface {
	Create(ctx context.Context, title, content string) (Note, error)
	ByID(ctx context.Context, idHex string) (Note, error)
	List(ctx context.Context, p ListParams) ([]Note, error)
	Update(ctx context.Context, idHex string, title, content *string) (Note, error)
	Delete(ctx context.Context, idHex string) error
	Stats(ctx context.Context) (Stats, error)
}
