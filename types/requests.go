package types

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 创建或者更新
type ReqData struct {
	Appid                   string `json:"appId" bson:"appId" comment:"应用名称" binding:"required"`
	Developers              string `json:"developers" bson:"developers" comment:"开发人员" binding:"required"`
	Releasetime             string `json:"releaseTime" bson:"releaseTime" comment:"应用投产时间" binding:"required"`
	Mode                    string `json:"mode" bson:"mode" comment:"应用发布方式" binding:"required"`
	Issql                   bool   `json:"isSql" bson:"isSql" comment:"是否需要sql" `
	Dependapps              string `json:"dependApps" bson:"dependApps" comment:"依赖的应用"`
	Desc                    string `json:"desc" bson:"desc" comment:"应用发布需求描述" binding:"required"`
	Currentupdatecollection string `json:"currentUpdateCollection" bson:"currentUpdateCollection" comment:"当前需要更新的集合"`
	Createtimestamp         string `json:"createTimeStamp" bson:"createTimeStamp" comment:"创建时间"`
	Updatatimestamp         string `json:"updateTimeStamp" bson:"updateTimeStamp" comment:"更新时间"`
}

// 删除
type DelReqData struct {
	Appid       string `json:"appId" bson:"appId" comment:"应用名称" binding:"required"`
	Releasetime string `json:"releaseTime" bson:"releaseTime" comment:"应用投产时间" binding:"required"`
}

// 批量删除
type DelBatchReqData struct {
	Appid       []string `json:"appId" bson:"appId" comment:"应用名称" binding:"required"`
	Releasetime string   `json:"releaseTime" bson:"releaseTime" comment:"应用投产时间"`
}

// 查询
type SearchReqData struct {
	Appid       string   `json:"appId" bson:"appId" comment:"应用名称" `
	Developers  []string `json:"developers" bson:"developers" comment:"开发人员" `
	Releasetime string   `json:"releaseTime" bson:"releaseTime" comment:"应用投产时间" `
}

// jenkins请求
type JobReqData struct {
	Appid string `json:"appId" bson:"appId" comment:"应用名称" binding:"required"`
}

// 批量删除
func (d *DelBatchReqData) DeleteBatch(db *mongo.Database) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var filter bson.M
	if d.Appid != nil {
		var regexPatterns []primitive.Regex
		for _, item := range d.Appid {
			regexPatterns = append(regexPatterns, primitive.Regex{Pattern: item, Options: "i"})
		}
		filter = bson.M{"app.appId": bson.M{"$in": regexPatterns}}
	} else {
		return 0, errors.New("appId is null")
	}

	result, err := db.Collection(d.Releasetime).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}
