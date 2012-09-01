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
	"fmt"
)

type Mapper struct{
	Map 		*map[string]interface{}
	Key 		string
	Ids			[]bson.ObjectId
	Single 		bool
}

func hasObjId(sl []interface{}) bool {
	for _, v := range sl {
		_, ok := v.(bson.ObjectId)
		if ok {
			return true
		}
	}
	return false
}

// Returns a []bson.ObjectId from a []interface{}.
// Non ObjectId values are appended as empty ObjectIds (they are not valid).
func toIdSlice(from []interface{}) []bson.ObjectId {
	ret := []bson.ObjectId{}
	for _, v := range from {
		val, ok := v.(bson.ObjectId)
		if ok {
			ret = append(ret, val)
		} else {
			ret = append(ret, "")
		}
	}
	return ret
}

// Separates each Mapper in accum by the collection they belong to.
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

// Recursively traverses the whole input and extracts all ObjectIds from it (remembering their position) to further processing.
// Caution: bson.M s are intentionally not handled here.
func extractIds(dat interface{}, acc *[]Mapper, parent map[string]interface{}, key string) {
	switch val := dat.(type) {
	case []interface{}:
		// if hasObjId(slice)...		// This is a yet not handled corner case.
		for _, v := range val {
			mapval, ok := v.(map[string]interface{})
			if !ok { continue }
			extractIds(v, acc, mapval, "")
		}
	case map[string]interface{}:
		for i, v := range val {
			if i == "_id" { continue }
			if slice, is_slice := v.([]interface{}); is_slice && hasObjId(slice) && string(i[0]) == "_" {
				m := Mapper{Map: &val,Key: i,Ids: toIdSlice(slice)}
				*acc = append(*acc, m)
			} else {
				extractIds(v, acc, val, i)
			}
		}
	case bson.ObjectId:
		m := Mapper{Map: &parent,Key: key,Ids: []bson.ObjectId{val},Single: true}
		*acc = append(*acc, m)
	case bson.M:
		panic("Please convert all bson.M maps to map[string]interface{} in query results.")
	}
}

type m map[string]interface{}

// Returns all ids from a []Mapper, so they can be queried.
func collectIds(ms []Mapper) []bson.ObjectId {
	ret := []bson.ObjectId{}
	for _, v := range ms {
		for _, z := range v.Ids {
			if z != "" {
				ret = append(ret, z)		// Here to support non bson.ObjectIds too.
			}
		}
	}
	return ret
}

// Takes a doc (z) from a query and with the help of the index it replaces the ObjectID with the doc.
func burnItIn(z bson.M, acc []Mapper, ind map[string][][2]int) {
	str_id := z["_id"].(bson.ObjectId).Hex()
	if index_sl, has := ind[str_id]; has {
		for _, v := range index_sl {
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

func query(db *mgo.Database, collname string, ids []bson.ObjectId, keys map[string]interface{}) []interface{} {
	var res []interface{}
	q := db.C(collname).Find(m{"_id": m{"$in": ids}})
	if keys != nil {
		q.Select(keys)
	}
	err := q.All(&res)
	if err != nil { panic(err) }
	return res
}

// Glues everything together.
func queryAndSet(db *mgo.Database, acc []Mapper, keys map[string]interface{}) {
	ind := index(acc)
	sep_accs := separateByColl(acc)
	for i, v := range sep_accs {
		ids := collectIds(v)
		res := query(db, i, ids, keys)
		for _, z := range res {
			burnItIn(z.(bson.M), acc, ind)
		}
	}
}

// Index creates an index out of a []Mapper.
// Index key will be the string representation of the bson.ObjectId.
// The value will be the [position in accum][position in ObjectId slice of the Mapper instance]
func index(accum []Mapper) map[string][][2]int {
	ret := map[string][][2]int{}
	for i, v := range accum {
		for j, z := range v.Ids {
			if z == "" { continue }		// Added to support []interface{}s where not all members are bson.ObjectIds.
			_, has := ret[z.Hex()]
			if !has {
				ret[z.Hex()] = [][2]int{}
			}
			ret[z.Hex()] = append(ret[z.Hex()], [2]int{i, j})
		}
	}
	return ret
}

func ResolveOne(db *mgo.Database, seed interface{}, keys map[string]interface{}) error {
	return ResolveAll(db, []interface{}{seed}, keys)
}

func ResolveAll(db *mgo.Database, seeds []interface{}, keys map[string]interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(fmt.Sprint(r))
		}
	}()
	acc := &[]Mapper{}
	for _, v := range seeds {
		extractIds(v, acc, v.(map[string]interface{}), "")
	}
	queryAndSet(db, *acc, keys)
	return nil
}