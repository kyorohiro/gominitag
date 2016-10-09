package gominitag

import (
	//	"encoding/json"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

type GaeObjectTag struct {
	ProjectId string
	MainTag   string
	SubTag    string
	TargetId  string
	Info      string `datastore:",noindex"`
	Created   time.Time
	Priority  int
	Type      string
}

type MiniTagManager struct {
	kind      string
	projectId string
}

type MiniTag struct {
	gaeObject    *GaeObjectTag
	gaeObjectKey *datastore.Key
	kind         string
}

type TagSource struct {
	MainTag string
	SubTag  string
	Type    string
}

const (
	TypeProjectId = "ProjectId"
	TypeMainTag   = "MainTag"
	TypeSubTag    = "SubTag"
	TypeTargetId  = "TargetId"
	TypeInfo      = "Info"
	TypeCreated   = "Created"
	TypePriority  = "Priority"
	TypeType      = "Type"
)

func NewMiniTagManager(kind string, projectId string) *MiniTagManager {
	ret := new(MiniTagManager)
	ret.kind = kind
	ret.projectId = projectId
	return ret
}

func (obj *MiniTagManager) DeleteTagsFromTargetId(ctx context.Context, targetId string, cursor string) (string, error) {
	q := datastore.NewQuery(obj.kind)
	q = q.Filter("ProjectId =", obj.projectId)
	q = q.Filter("TargetId =", targetId)
	q = q.Order("-Created")
	r, _, eCursor := obj.FindTagKeyFromQuery(ctx, q, cursor)
	for _, v := range r {
		datastore.Delete(ctx, v)
	}
	return eCursor, nil
}

func (obj *MiniTagManager) AddPairTags(ctx context.Context, tagList []string, info string, targetId string) error {
	vs := obj.MakeTags(ctx, tagList)
	for _, v := range vs {
		log.Infof(ctx, ">>"+v.MainTag+" : "+v.SubTag)
		tag := obj.NewTag(ctx, v.MainTag, v.SubTag, targetId, v.Type)
		err := tag.SaveOnDB(ctx)
		if err != nil {
			log.Infof(ctx, ">>"+err.Error())
		}
	}
	return nil
}

func (obj *MiniTagManager) DeletePairTags(ctx context.Context, tagList []string, info string, targetId string) error {
	vs := obj.MakeTags(ctx, tagList)
	for _, v := range vs {
		key := obj.NewTagKey(ctx, v.MainTag, v.SubTag, targetId, v.Type)
		datastore.Delete(ctx, key)
	}
	return nil
}

func (obj *MiniTagManager) AddMainTag(ctx context.Context, tag1 string, tag2 string, info string, targetId string) error {
	return obj.AddTag(ctx, tag1, tag2, info, targetId, "main")
}

func (obj *MiniTagManager) AddSubTag(ctx context.Context, tag1 string, tag2 string, info string, targetId string) error {
	return obj.AddTag(ctx, tag1, tag2, info, targetId, "sub")
}

//
//
//
func (obj *MiniTagManager) AddTag(ctx context.Context, tag1 string, tag2 string, info string, targetId string, Type string) error {
	mainTag := tag1
	subTag := tag2
	if subTag != "" && strings.Compare(tag1, tag2) <= 0 {
		mainTag = tag2
		subTag = tag1
	}
	tag := obj.NewTag(ctx, mainTag, subTag, targetId, Type)
	tag.gaeObject.Info = info
	return tag.SaveOnDB(ctx)
}

func (obj *MiniTagManager) DeleteMainTag(ctx context.Context, MainTag string, SubTag string, TargetId string) error {
	return obj.DeleteTag(ctx, MainTag, SubTag, TargetId, "main")
}

func (obj *MiniTagManager) DeleteSubTag(ctx context.Context, MainTag string, SubTag string, TargetId string) error {
	return obj.DeleteTag(ctx, MainTag, SubTag, TargetId, "sub")
}

func (obj *MiniTagManager) DeleteTag(ctx context.Context, MainTag string, SubTag string, TargetId string, Type string) error {
	key := obj.NewTagKey(ctx, MainTag, SubTag, TargetId, Type)
	datastore.Delete(ctx, key)
	return nil
}

func (obj *MiniTagManager) MakeTags(ctx context.Context, tagList []string) []TagSource {
	ret := make([]TagSource, 0)
	for _, x := range tagList {
		isSave := false
		for _, y := range tagList {
			if strings.Compare(x, y) > 0 {
				t := "sub"
				if isSave == false {
					t = "main"
				}
				ret = append(ret, TagSource{
					MainTag: x,
					SubTag:  y,
					Type:    t,
				})
				isSave = true
			}
		}
		if isSave == false {
			ret = append(ret, TagSource{
				MainTag: x,
				SubTag:  "",
				Type:    "main",
			})
		}
	}
	return ret
}

func (obj *MiniTagManager) NewTag(ctx context.Context, mainTag string, //
	subTag string, targetId string, t string) *MiniTag {
	ret := new(MiniTag)
	ret.gaeObject = new(GaeObjectTag)
	ret.gaeObject.ProjectId = obj.projectId
	ret.gaeObject.MainTag = mainTag
	ret.gaeObject.SubTag = subTag
	ret.gaeObject.TargetId = targetId
	ret.gaeObjectKey = obj.NewTagKey(ctx, mainTag, subTag, targetId, t)
	ret.gaeObject.Created = time.Now()
	ret.gaeObject.Type = t
	return ret
}

func (obj *MiniTagManager) NewTagKey(ctx context.Context, mainTag string, //
	subTag string, targetId string, ttype string) *datastore.Key {
	ret := datastore.NewKey(ctx, obj.kind, ""+obj.projectId+"::"+targetId+"::"+mainTag+"::"+subTag+"::"+ttype, 0, nil)
	return ret
}

func (obj *MiniTagManager) NewTagFromKey(ctx context.Context, gaeKey *datastore.Key) (*MiniTag, error) {

	ret := new(MiniTag)
	ret.kind = obj.kind
	ret.gaeObject = new(GaeObjectTag)
	ret.gaeObjectKey = gaeKey
	//
	//
	memObjSrc, memObjErr := memcache.Get(ctx, gaeKey.StringID())
	if memObjErr == nil {
		err := ret.SetParamFromsJson(ctx, memObjSrc.Value)
		if err == nil {
			return ret, nil
		}
	}
	//
	//
	err := datastore.Get(ctx, gaeKey, ret.gaeObject)
	if err != nil {
		return nil, err
	}
	//
	//	ret.SetParamFromsJson(ctx)

	return ret, nil
}

/*

- kind: ArticleTag
  properties:
  - name: ProjectId
  - name: MainTag
  - name: SubTag
  - name: Created
    direction: desc

- kind: ArticleTag
  properties:
  - name: ProjectId
  - name: MainTag
  - name: Type
  - name: Created
    direction: desc

https://cloud.google.com/appengine/docs/go/config/indexconfig#updating_indexes
*/
func (obj *MiniTagManager) FindTags(ctx context.Context, mainTag string, subTag string, cursorSrc string) ([]*MiniTag, string, string) {
	q := datastore.NewQuery(obj.kind)
	q = q.Filter("ProjectId =", obj.projectId)
	q = q.Filter("MainTag =", mainTag)

	if subTag != "" {
		q = q.Filter("SubTag =", subTag)
	} else {
		q = q.Filter("Type =", "main")
	}

	q = q.Order("-Created").Limit(10)
	return obj.FindTagFromQuery(ctx, q, cursorSrc)
}

/*
- kind: ArticleTag
  properties:
  - name: ProjectId
  - name: TargetId
  - name: Created
    direction: desc
https://cloud.google.com/appengine/docs/go/config/indexconfig#updating_indexes
*/
func (obj *MiniTagManager) FindTagFromTargetId(ctx context.Context, targetTag string, cursorSrc string) ([]*MiniTag, string, string) {
	q := datastore.NewQuery(obj.kind)
	q = q.Filter("ProjectId =", obj.projectId)
	q = q.Filter("TargetId =", targetTag)
	q = q.Order("-Created").Limit(10)
	return obj.FindTagFromQuery(ctx, q, cursorSrc)
}

func (obj *MiniTagManager) FindTagKeyFromQuery(ctx context.Context, q *datastore.Query, cursorSrc string) ([]*datastore.Key, string, string) {
	cursor := obj.newCursorFromSrc(cursorSrc)
	if cursor != nil {
		q = q.Start(*cursor)
	}
	q = q.KeysOnly()
	founds := q.Run(ctx)

	var retUser []*datastore.Key

	var cursorNext string = ""
	var cursorOne string = ""

	for i := 0; ; i++ {
		var d GaeObjectTag
		key, err := founds.Next(&d)
		if err != nil || err == datastore.Done {
			break
		} else {
			retUser = append(retUser, key)
		}
		if i == 0 {
			cursorOne = obj.makeCursorSrc(founds)
		}
	}
	cursorNext = obj.makeCursorSrc(founds)
	return retUser, cursorOne, cursorNext
}

func (obj *MiniTagManager) FindTagFromQuery(ctx context.Context, q *datastore.Query, cursorSrc string) ([]*MiniTag, string, string) {
	cursor := obj.newCursorFromSrc(cursorSrc)
	if cursor != nil {
		q = q.Start(*cursor)
	}
	q = q.KeysOnly()
	founds := q.Run(ctx)

	var retUser []*MiniTag

	var cursorNext string = ""
	var cursorOne string = ""

	for i := 0; ; i++ {
		var d GaeObjectTag
		key, err := founds.Next(&d)
		if err != nil || err == datastore.Done {
			break
		} else {
			dobj, derr := obj.NewTagFromKey(ctx, key)
			if derr == nil {
				retUser = append(retUser, dobj)
			}
		}
		if i == 0 {
			cursorOne = obj.makeCursorSrc(founds)
		}
	}
	cursorNext = obj.makeCursorSrc(founds)
	return retUser, cursorOne, cursorNext
}

func (obj *MiniTagManager) newCursorFromSrc(cursorSrc string) *datastore.Cursor {
	c1, e := datastore.DecodeCursor(cursorSrc)
	if e != nil {
		return nil
	} else {
		return &c1
	}
}

func (obj *MiniTagManager) makeCursorSrc(founds *datastore.Iterator) string {
	c, e := founds.Cursor()
	if e == nil {
		return c.String()
	} else {
		return ""
	}
}
