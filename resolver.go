// Way to resolve database references without DBRefs.
// Can only work in the same database for now.
//
// Now only resolves one level.
// Later the package will be able to handle cascading reference resolution.
// Will detect circular references, so it won't run into an infinite loop.
//
// It will resolve any value in the seed, at any depth which is a bson.ObjectId or []bson.ObjectId, and has a key of the form:
// _collectionName or optionally _collectionName_fieldName (if there are more than one references in to the same collection)
package resolver

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

type Mapper struct{
	Map 		*map[string]interface{}
	Key 		string
	Ids			[]bson.ObjectId
	Single 		bool
}

func allIsObjId(sl []interface{}) bool {
	for _, v := range sl {
		_, ok := v.(bson.ObjectId)
		if !ok {
			return false
		}
	}
	return true
}

func toIdSlice(from []interface{}) []bson.ObjectId {
	ret := []bson.ObjectId{}
	for _, v := range from {
		ret = append(ret, v.(bson.ObjectId))
	}
	return ret
}

func separateByColl(accum []Mapper) map[string][]Mapper {
	ret := map[string][]Mapper{}
	for _, v := range accum {
		if string(v.Key[0]) == "_" {
			collname := strings.Split(v.Key, "_")[1]
			_, has := ret[collname]
			if !has {
				ret[collname] = []Mapper{}
			}
			ret[collname] = append(ret[collname], v)
		}
	}
	return ret
}

func extractIds(dat interface{}, acc *[]Mapper, parent map[string]interface{}, key string) {
	switch val := dat.(type) {
	case map[string]interface{}:
		for i, v := range val {
			if i == "_id" { continue }
			if slice, is_slice := v.([]interface{}); is_slice && allIsObjId(slice) && string(i[0]) == "_" {
				m := Mapper{Map: &val,Key: i,Ids: toIdSlice(slice)}
				*acc = append(*acc, m)
				val[i] = make([]interface{}, len(slice))
			} else {
				extractIds(v, acc, val, i)
			}
		}
	case bson.ObjectId:
		*acc = append(*acc, Mapper{Map: &parent,Key: key,Ids: []bson.ObjectId{val},Single: true})
	}
}

type m map[string]interface{}

func collectIds(ms []Mapper) []bson.ObjectId {
	ret := []bson.ObjectId{}
	for _, v := range ms {
		ret = append(ret, v.Ids...)
	}
	return ret
}

func burnItIn(z bson.M, acc []Mapper, ind map[string][][2]int) {
	str_id := z["_id"].(bson.ObjectId).Hex()
	if index, has := ind[str_id]; has {
		for _, v := range index {
			mapper := acc[v[0]]
			if mapper.Single {
				(*mapper.Map)[mapper.Key] = z
			} else {
				(*mapper.Map)[mapper.Key].([]interface{})[v[1]] = z
			}
		}
	} else {
		panic("Unown bug in resolver.")
	}
}

func queryAndSet(db *mgo.Database, acc []Mapper, keys map[string]interface{}) {
	ind := index(acc)
	sep_accs := separateByColl(acc)
	for i, v := range sep_accs {
		ids := collectIds(v)
		var res []interface{}
		q := db.C(i).Find(m{"_id": m{"$in": ids}})
		if keys != nil {
			q.Select(keys)
		}
		q.All(&res)
		for _, z := range res {
			burnItIn(z.(bson.M), acc, ind)
		}
	}
}

func index(accum []Mapper) map[string][][2]int {
	ret := map[string][][2]int{}
	for i, v := range accum {
		for j, z := range v.Ids {
			_, has := ret[z.Hex()]
			if !has {
				ret[z.Hex()] = [][2]int{}
			}
			ret[z.Hex()] = append(ret[z.Hex()], [2]int{i, j})
		}
	}
	return ret
}

func ResolveOne(db *mgo.Database, seed interface{}, keys map[string]interface{}) {
	ResolveAll(db, []interface{}{seed}, keys)
}

func ResolveAll(db *mgo.Database, seeds []interface{}, keys map[string]interface{}) {
	acc := &[]Mapper{}
	for _, v := range seeds {
		extractIds(v, acc, v.(map[string]interface{}), "")
	}
	queryAndSet(db, *acc, keys)
}