// Way to resolve database references without DBRefs.
// Can only work in the same database for now.
//
// Now only resolves one level.
// Later the package will be able to handle cascading reference resolution, now only one level.
// Will detect circular references, so it won't run in an infinite loop.
//
// It will resolve any value in the seed, at any depth which is an mgo.ObjectId, and has a key of the form:
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
	return false
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
		if string(v.Key[0]) == string("_") {
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
			if slice, is_slice := v.([]interface{}); is_slice && allIsObjId(slice) {
				*acc = append(*acc, Mapper{Map: &parent,Key: i,Ids: toIdSlice(slice)})
				val[i] = []bson.M{}
			} else {
				extractIds(v, acc, val, i)
			}
		}
	case bson.ObjectId:
		*acc = append(*acc, Mapper{Map: &parent,Key: key,Ids: []bson.ObjectId{val},Single: true})
	}
}

type m map[string]interface{}

func harvest(ms []Mapper) []bson.ObjectId {
	ret := []bson.ObjectId{}
	for _, v := range ms {
		ret = append(ret, v.Ids...)
	}
	return ret
}

func burnItIn(z bson.M, acc []Mapper, ind map[string]int) {
	str_id := z["_id"].(bson.ObjectId).Hex()
	if index, has := ind[str_id]; has {
		mapper := acc[index]
		if mapper.Single {
			(*mapper.Map)[mapper.Key] = z
		} else {
			(*mapper.Map)[mapper.Key] = append((*mapper.Map)[mapper.Key].([]interface{}), z)
		}
	} else {
		panic("Unown bug in resolver.")
	}
}

func glue(db *mgo.Database, sep_accs map[string][]Mapper, acc []Mapper, ind map[string]int) {
	for i, v := range sep_accs {
		ids := harvest(v)
		var res []interface{}
		db.C(i).Find(m{"_id": m{"$in": ids}}).All(&res)
		for _, z := range res {
			burnItIn(z.(bson.M), acc, ind)
		}
	}
}

func index(accum []Mapper) map[string]int {
	ret := map[string]int{}
	for i, v := range accum {
		for _, z := range v.Ids {
			ret[z.Hex()] = i
		}
	}
	return ret
}

func ResolveOne(db *mgo.Database, seed map[string]interface{}) {
	ResolveAll(db, []map[string]interface{}{seed})
}

func ResolveAll(db *mgo.Database, seeds []map[string]interface{}) {
	acc := &[]Mapper{}
	for _, v := range seeds {
		extractIds(v, acc, v, "")
	}
	ind := index(*acc)
	sep_accs := separateByColl(*acc)
	glue(db, sep_accs, *acc, ind)
}