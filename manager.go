package gominitag

import (
	//	"encoding/json"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
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

//
// TODO
//
func (obj *MiniTagManager) AddTags(ctx context.Context, tagList []string, targetId string) error {
	for _, x := range tagList {
		isSave := false
		for _, y := range tagList {
			if strings.Compare(x, y) > 0 {
				t := "sub"
				if isSave == false {
					t = "main"
				}
				tag := obj.NewTag(ctx, x, y, targetId, t)
				tag.SaveOnDB(ctx)
				isSave = true
			}
		}
		if isSave == false {
			tag := obj.NewTag(ctx, x, "", targetId, "main")
			tag.SaveOnDB(ctx)
		}
	}
	return nil
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
func (obj *MiniTagManager) FindTagFromTagPlus(ctx context.Context, mainTag string, subTag string, cursorSrc string) ([]*MiniTag, string, string) {
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
