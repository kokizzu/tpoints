package handler

import (
	"context"
	"github.com/beeker1121/goque"
	"github.com/kokizzu/gotro/L"
	"github.com/kokizzu/gotro/M"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"tpoints/model"
)

type Server struct {
	MongoClient *mongo.Client
	DiskQueue *goque.Queue
	Points *model.MPoints
}

type Ctx struct {
	echo.Context
	*Server
	HasError error
}

func (ctx *Ctx) ReturnError(err error) bool {
	if err != nil {
		ctx.JSON(200,M.SX{
			`error`: err.Error(),
		})
		ctx.HasError = err
	}
	return ctx.HasError != nil
}

func (ctx *Ctx) IsError(err error,description string) bool {
	if L.IsError(err,description) {
		ctx.JSON(200,M.SX{
			`error`: description,
		})
		ctx.HasError = err
	}
	return ctx.HasError != nil
}

func InitServer() *Server {
	server := &Server{}
	var err error
	ctx, _ := context.WithTimeout(context.Background(), 4*time.Second)
	// mongo
	server.MongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(MONGO_URL))
	L.PanicIf(err,`failed to connect to mongo`) 
	server.Points = &model.MPoints{Collection: server.MongoClient.Database(MONGO_DB).Collection(MONGO_COLLECTION)}
	server.Points.CreateIndex()
	// goque
	server.DiskQueue, err = goque.OpenQueue(GOQUE_DIR)
	L.PanicIf(err,`failed to open goque directory, forgot mkdir or still used by other process?`) 
	go server.RunDequeue()
	return server
}

func (s *Server) RunDequeue() {
	for {
		item, err := s.DiskQueue.Peek()
		if err == goque.ErrEmpty { // no item, sleep for a moment
			time.Sleep(1 * time.Second)
			continue
		} else if L.IsError(err,`failed peeking goque`) {
			time.Sleep(2 * time.Second)
			continue
		}
		point := model.Point{}
		err = bson.Unmarshal(item.Value,&point)
		if L.IsError(err,`failed decoding goque`) {
			time.Sleep(2 * time.Second)
			continue
		}
		//L.Describe(point) // processing
		s.Points.Process(&point)
		_, err = s.DiskQueue.Dequeue()
		if L.IsError(err,`failed dequeue goque`) {
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

func (s *Server) Wrap(handler func(*Ctx)) (echo.HandlerFunc){
	return func(ctx echo.Context) error {
		ctx2 := Ctx{
			Context: ctx,
			Server:  s,
		}
		handler(&ctx2)
		return ctx2.HasError
	}
}

func (s *Server) Close() {
	s.MongoClient.Disconnect(context.Background())
	s.DiskQueue.Close()
	s.MongoClient = nil
	s.DiskQueue = nil
	s.Points = nil
}
