package types

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"k8s.io/klog/v2"
)

const (
	INQUEUE string = "INQUEUE"
	RUNNING string = "RUNNING"
	DONE    string = "DONE"
	FAILED  string = "FAILED"
)

type Releases struct {
	//Timestamp string `json:"timestamp" bson:"timestamp" comment:"发布单日期"`
	List  []Release `json:"lists" bson:"total" comment:"发布单列表"`
	Total int       `json:"total,omitempty" bson:"total,omitempty" comment:"发布单应用总数"`
}

type Release struct {
	App    App    `json:"app" bson:"app"`
	Status Status `json:"status,omitempty" bson:"status,omitempty"`
}

type App struct {
	Appid       string   `json:"appId" bson:"appId" comment:"应用名称" binding:"required"`
	Developers  []string `json:"developers" bson:"developers" comment:"开发人员" binding:"required"`
	Releasetime string   `json:"releaseTime" bson:"releaseTime" comment:"应用投产时间" binding:"required"`
	Mode        string   `json:"mode,omitempty" bson:"mode,omitempty" comment:"应用发布方式" default:"jenkins"`
	Issql       bool     `json:"isSql,omitempty" bson:"isSql,omitempty" comment:"是否需要sql" default:"false"`
	Dependapps  []string `json:"dependApps,omitempty" bson:"dependApps,omitempty" comment:"依赖的应用"`
	Desc        string   `json:"desc,omitempty" bson:"desc,omitempty" comment:"应用发布需求描述"`
}

type Status struct {
	Createtimestamp string       `json:"createTimeStamp,omitempty" bson:"createTimeStamp,omitempty" comment:"创建时间"`
	Dev             StatusOption `json:"dev,omitempty" bson:"dev,omitempty"`
	Test            StatusOption `json:"test,omitempty" bson:"test,omitempty"`
	Pre             StatusOption `json:"pre,omitempty" bson:"pre,omitempty"`
	Prd             StatusOption `json:"prd,omitempty" bson:"prd,omitempty"`
}

type StatusOption struct {
	Updatestatus []UpdateStatus `json:"updateStatus,omitempty" bson:"updateStatus,omitempty" comment:"更新状态"`
	Inspector    string         `json:"inspector,omitempty" bson:"inspector,omitempty" comment:"检查人"`

	IsPackage bool `json:"isPackage,omitempty" bson:"isPackage,omitempty" comment:"是否运行打包"`
	IsDeploy  bool `json:"isPass,omitempty" bson:"isPass,omitempty" comment:"是否允许部署"`

	Packagestatus JobQueue `json:"packageStatus,omitempty" bson:"packageStatus,omitempty" comment:"队列id"`
	PackageCounts int64    `json:"packageCounts,omitempty" bson:"packageCounts,omitempty" comment:"打包次数"`

	Deploystatus DeployQueue `json:"deployStatus,omitempty" bson:"deployStatus,omitempty" comment:"部署队列"`
}

type JobQueue struct {
	Queueid     int64  `json:"queueId,omitempty" bson:"queueId,omitempty" comment:"队列id"`
	Buildnumber int64  `json:"buildNumber,omitempty" bson:"buildNumber,omitempty" comment:"打包编号"`
	Packagetime string `json:"packageTime,omitempty" bson:"packageTime,omitempty" comment:"打包时间"`
	Status      string `json:"status" bson:"status" comment:"package状态"`
}

type DeployQueue struct {
	Id         int64  `json:"id,omitempty" bson:"id,omitempty" comment:"构建的id=buildnumber"`
	Deploytime string `json:"deployTime,omitempty" bson:"deployTime,omitempty" comment:"部署时间"`
	Status     string `json:"status" bson:"status" comment:"deploy状态"`
}

type UpdateStatus struct {
	Timestamp string `json:"timeStamp,omitempty" bson:"timeStamp,omitempty" comment:"时间"`
	Histroy   App    `json:"histroy,omitempty" bson:"histroy,omitempty" comment:"历史"`
}

func NewApp(appId, releaseTime string, developers []string) *App {
	return &App{
		Appid:       appId,
		Releasetime: releaseTime,
		Developers:  developers,
	}
}

// 新增/插入发布单
func (r *Release) Add(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.Collection(r.App.Releasetime).InsertOne(ctx, r)

	return err
}

func FindMany(db *mongo.Database, collection string, key string, args ...string) ([]Release, error) {
	var results []Release

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var regexPatterns []primitive.Regex
	for _, trim := range args {
		regexPatterns = append(regexPatterns, primitive.Regex{Pattern: trim, Options: "i"})
	}
	filter := bson.M{"$in": regexPatterns}
	cur, err := db.Collection(collection).Find(ctx, bson.M{key: filter})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// 单个查询
func (r *Release) Search(db *mongo.Database) ([]Release, error) {
	var results []Release

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"app.appId": r.App.Appid, "app.releaseTime": r.App.Releasetime}
	klog.V(3).Infof("search filter: %v, collection: %s", filter, r.App.Releasetime)
	cur, err := db.Collection(r.App.Releasetime).Find(ctx, filter)

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, err

}

func (a *App) SearchByDeveloper(db *mongo.Database, collection string) ([]App, error) {
	var results []App

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	searchTerm := a.Developers
	var regexPatterns []primitive.Regex
	for _, trim := range searchTerm {
		regexPatterns = append(regexPatterns, primitive.Regex{Pattern: trim, Options: "i"})
	}
	filter := bson.M{"app.developers": bson.M{"$in": regexPatterns}}

	klog.V(3).InfoS("mongodb filter", "filter", filter)
	cur, err := db.Collection(collection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, err
}

// 列出所有数据
func (rs *Releases) RelasesList(db *mongo.Database, collection string, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var filter = bson.M{}
	klog.V(3).Infof("collection: %s, args: %v", collection, args)
	if args == nil {
		var regexPatterns []primitive.Regex
		for _, item := range args {
			regexPatterns = append(regexPatterns, primitive.Regex{Pattern: item, Options: "i"})
		}
		filter = bson.M{"appId": bson.M{"$in": regexPatterns}}
	}

	klog.V(3).InfoS("mongodb filter", "filter", filter)
	cur, err := db.Collection(collection).Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &rs.List); err != nil {
		return err
	}

	rs.Total = len(rs.List)
	return err
}

// 指定发布单删除
func (r *Release) Delete(db *mongo.Database) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"app.appId": r.App.Appid, "app.releaseTime": r.App.Releasetime}
	klog.V(3).Infof("delete filter: %v, collection: %s", filter, r.App.Releasetime)
	result, err := db.Collection(r.App.Releasetime).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// 发布单更新
func (r *Release) Update(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"app.appId": r.App.Appid, "app.releaseTime": r.App.Releasetime}

	klog.V(3).Infoln(r.App)
	result, err := db.Collection(r.App.Releasetime).UpdateMany(ctx, filter, bson.M{"$set": bson.M{"app": r.App, "status.dev.updateStatus": r.Status.Dev.Updatestatus}}, options.MergeUpdateOptions())
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("not found")
	}
	if result.ModifiedCount == 0 {
		return errors.New("not modified")
	}
	return nil
}

// 更新发布单状态
func (r *Release) UpdateStatus(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"app.appId": r.App.Appid, "app.releaseTime": r.App.Releasetime}
	if _, err := db.Collection(r.App.Releasetime).UpdateMany(ctx, filter, bson.M{"$set": bson.M{"status": r.Status}}, options.MergeUpdateOptions()); err != nil {
		return err
	}

	return nil
}
