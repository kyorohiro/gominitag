package gominitag

import (
	"encoding/json"
	//	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

func (obj *MiniTag) toJson() (string, error) {
	v := map[string]interface{}{
		TypeProjectId: obj.gaeObject.ProjectId,
		TypeMainTag:   obj.gaeObject.MainTag,
		TypeSubTag:    obj.gaeObject.SubTag,
		TypeTargetId:  obj.gaeObject.TargetId,
		TypeInfo:      obj.gaeObject.Info,
		TypeCreated:   obj.GetCreated().UnixNano(),
		TypePriority:  obj.gaeObject.Priority,
		TypeType:      obj.gaeObject.Type,
	}
	vv, e := json.Marshal(v)
	return string(vv), e
}

//
func (obj *MiniTag) SetParamFromsJson(ctx context.Context, source []byte) error {
	v := make(map[string]interface{})
	e := json.Unmarshal(source, &v)
	if e != nil {
		return e
	}
	//
	obj.gaeObject.ProjectId = v[TypeProjectId].(string)
	obj.gaeObject.MainTag = v[TypeMainTag].(string)
	obj.gaeObject.SubTag = v[TypeSubTag].(string)
	obj.gaeObject.TargetId = v[TypeTargetId].(string)
	obj.gaeObject.Info = v[TypeInfo].(string)
	obj.gaeObject.Created = time.Unix(0, int64(v[TypeCreated].(float64))) //srcCreated
	obj.gaeObject.Priority = int(v[TypePriority].(float64))
	obj.gaeObject.Type = v[TypeType].(string)

	return nil
}

func (obj *MiniTag) GetProjectId() string {
	return obj.gaeObject.ProjectId
}

func (obj *MiniTag) GetMainTag() string {
	return obj.gaeObject.MainTag
}

func (obj *MiniTag) GetSubTag() string {
	return obj.gaeObject.SubTag
}

func (obj *MiniTag) GetTargetId() string {
	return obj.gaeObject.TargetId
}

func (obj *MiniTag) GetCreated() time.Time {
	return obj.gaeObject.Created
}

func (obj *MiniTag) GetPriority() int {
	return obj.gaeObject.Priority
}

func (obj *MiniTag) GetGaeObjectKey() *datastore.Key {
	return obj.gaeObjectKey
}

func (obj *MiniTag) SaveOnDB(ctx context.Context) error {
	_, err := datastore.Put(ctx, obj.gaeObjectKey, obj.gaeObject)
	memSrc, errMemSrc := obj.toJson()
	if err == nil && errMemSrc == nil {
		objMem := &memcache.Item{
			Key:   obj.gaeObjectKey.StringID(),
			Value: []byte(memSrc), //
		}
		memcache.Set(ctx, objMem)
	}
	return err
}
