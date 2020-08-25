package model

import (
	"context"
	"github.com/kokizzu/gotro/I"
	"github.com/kokizzu/gotro/L"
	"github.com/kokizzu/gotro/M"
	"github.com/kokizzu/gotro/T"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Point struct {
	Id        string `bson:"_id"`
	UserId    string `bson:"userId"`
	Delta     int64
	Sum       int64
	QueueAt   int64  `bson:"queueAt"`
	ProcessAt int64  `bson:"processAt"`
}
const _UserId = `userId`
const _ProcessAt = `processAt`

type MPoints struct {
	*mongo.Collection
}

func (m *MPoints) Process(newPoint *Point) error {
	client := m.Collection.Database().Client()
	
	return client.UseSession(context.Background(), func(session mongo.SessionContext) error {
		err := session.StartTransaction()
		if L.IsError(err, `failed start transaction`) {
			return err
		}
		
		// find last record
		res := m.FindOne(session, bson.M{
			_UserId: newPoint.UserId,
		}, &options.FindOneOptions{
			Sort: M.SI{_ProcessAt: -1},
		})
		
		lastPoint := Point{}
		err = res.Decode(&lastPoint)
		if err != mongo.ErrNoDocuments {
			if L.IsError(err, `error findOne/decode lastPoint`) {
				return err
			}
		}
		
		newPoint.Sum = lastPoint.Sum + newPoint.Delta
		newPoint.ProcessAt = T.UnixNano()
		newPoint.Id = newPoint.UserId + `_` + I.ToS(newPoint.ProcessAt)
		_, err = m.InsertOne(session, newPoint)
		if L.IsError(err, `failed insertOne newPoint`) {
			return err
		}
		
		return session.CommitTransaction(session)
	})
}

func (m *MPoints) TotalLogs_ByUser(userId string) (count int64, err error) {
	count, err = m.CountDocuments(context.Background(), bson.M{
		_UserId: userId,
	})
	if L.IsError(err, `failed aggregate count`) {
		return
	}
	return
}

func (m *MPoints) CreateIndex() {
	mod := mongo.IndexModel{
		Keys: bson.M{
			_UserId: 1,
		},
	}
	_, err := m.Indexes().CreateOne(context.Background(), mod)
	if L.IsError(err,`failed createing userId index`) {
		return
	}
	mod = mongo.IndexModel{
		Keys: bson.M{
			_ProcessAt: -1,
		},
	}
	_, err = m.Indexes().CreateOne(context.Background(), mod)
	if L.IsError(err,`failed creating processAt index`) {
		return
	}
}

func (m *MPoints) Logs_ByUser_ByOffset_ByLImit(userId string, offset int64, limit int64) (rows []Point, err error) {
	rows = []Point{}
	var cursor *mongo.Cursor
	cursor, err = m.Find(context.Background(), bson.M{
		_UserId: userId,
	}, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
		Sort: map[string]int{
			_ProcessAt: -1,
		},
	})
	if L.IsError(err, `failed find by userid`) {
		return
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		row := Point{}
		err = cursor.Decode(&row)
		if L.IsError(err, `failed decoding find cursor`) {
			return
		}
		rows = append(rows, row)
	}
	return
}

func (m *MPoints) DeleteAll_ByUser(userId string) (err error) {
	_, err = m.DeleteMany(context.Background(), bson.M{
		_UserId: bson.M{
			`$eq`: userId,
		},
	})
	return
}
