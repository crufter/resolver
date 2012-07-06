// Way to resolve database references without DBRefs.
// Can only work in the same database for now.
//
// Now only resolves one level.
// Later the package will be able to handle cascading reference resolution, now only one level.
// Can detect circular references, so it won't run in an infinite loop.
//
// It will resolve any value in the seed, at any depth which is an mgo.ObjectId, and has a key of the form:
// _collectionName or optionally _collectionName_fieldName (if there are more than one references in to the same collection)
package resolver

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"fmt"
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
			} else {
				extractIds(v, acc, val, i)
			}
		}
	case bson.ObjectId:
		*acc = append(*acc, Mapper{Map: &parent,Key: key,Ids: []bson.ObjectId{val},Single: true})
		fmt.Println("lol")
	}
}



func burnThemIn(db *mgo.Database, sep_accs map[string][]Mapper) {
	for i, v := range sep_accs {
		db.C(i).Find
	}
}

func ResolveOne(db *mgo.Database, seed map[string]interface{}) {
	ResolveAll(db, []map[string]interface{}{seed})
}

func ResolveAll(db *mgo.Database, seeds []map[string]interface{}) {
	acc := &[]Mapper{}
	for _, v := range seeds {
		extractIds(v, acc, v, "")
	}
	sep_accs := separateByColl(*acc)
	fmt.Println(sep_accs)
	burnThemIn(db, sep_accs)
}