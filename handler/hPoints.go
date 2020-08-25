package handler

import (
	"github.com/kokizzu/gotro/I"
	"github.com/kokizzu/gotro/S"
	"github.com/kokizzu/gotro/T"
	"go.mongodb.org/mongo-driver/bson"
	"tpoints/model"
)

type Points_Logs_Response struct {
	Rows  []model.Point
	Count int64
	Error string `json:",omitempty"`
}

// @Summary see user's point history
// @Description get processed logs
// @Accept json
// @Produce json
// @Param userId path string true "user id"
// @Param limit path int true "page size, default: 10"
// @Param offset path int true "page, default: 0"
// @Success 200 {object} Points_Logs_Response
// @Router /points/logs/{userId}/{limit}/{offset} [get]
func Points_Logs(ctx *Ctx) {
	userId := ctx.Param(`userId`)
	offset := I.Max(0, S.ToI(ctx.Param(`offset`)))
	limit := I.Min(10, S.ToI(ctx.Param(`limit`)))
	count, err := ctx.Points.TotalLogs_ByUser(userId)
	if ctx.ReturnError(err) {
		return
	}
	rows, err := ctx.Points.Logs_ByUser_ByOffset_ByLImit(userId, offset, limit)
	if ctx.ReturnError(err) {
		return
	}
	ctx.JSON(200, Points_Logs_Response{
		Rows:  rows,
		Count: count,
	})
}

type Points_Queue_Response struct {
	Point model.Point `json:",omitempty"`
	Error string      `json:",omitempty"`
}

// @Summary add points to specific user later 
// @Description async/enqueue point change event
// @Accept json
// @Produce json
// @Param userId path string true "user id"
// @Param delta path int true "can be positive or negative"
// @Success 200 {object} Points_Queue_Response
// @Router /points/queue/{userId}/{delta} [post]
func Points_Queue(ctx *Ctx) {
	userId := ctx.Param(`userId`)
	delta := S.ToI(ctx.Param(`delta`))
	point := model.Point{
		UserId:  userId,
		Delta:   delta,
		QueueAt: T.UnixNano(),
	}
	blob, err := bson.Marshal(point)
	if ctx.IsError(err, `failed marshall Point`) {
		return 
	}
	_, err = ctx.DiskQueue.Enqueue(blob)
	if ctx.IsError(err, `failed enqueue Point`) {
		return 
	}
	ctx.JSON(200, Points_Queue_Response{Point: point})
}

type Points_Add_Response struct {
	Point model.Point `json:",omitempty"`
	Error string      `json:",omitempty"`
}

// @Summary add points to specific user now
// @Description sync point addition/subtraction
// @Accept json
// @Produce json
// @Param userId path string true "user id"
// @Param delta path int true "can be positive or negative"
// @Success 200 {object} Points_Add_Response
// @Router /points/add/{userId}/{delta} [post]
func Points_Add(ctx *Ctx) {
	userId := ctx.Param(`userId`)
	delta := S.ToI(ctx.Param(`delta`))
	point := model.Point{
		UserId:  userId,
		Delta:   delta,
		QueueAt: T.UnixNano(),
	}
	err := ctx.Points.Process(&point)
	if ctx.ReturnError(err) {
		return
	}
	ctx.JSON(200, Points_Add_Response{
		Point: point,
	})
}
