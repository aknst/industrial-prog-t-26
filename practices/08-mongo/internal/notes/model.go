package notes

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// ListParams описывает фильтр и пагинацию списка.
// After (hex ObjectID) включает пагинацию по курсору вместо Skip.
type ListParams struct {
	Q     string
	Limit int64
	Skip  int64
	After string
}

// Stats — результат aggregation pipeline.
type Stats struct {
	Count            int     `bson:"count" json:"count"`
	AvgContentLength float64 `bson:"avgContentLength" json:"avgContentLength"`
}
