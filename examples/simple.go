// TODO do a test suite out of this.
package main

import (
	"fmt"
	"github.com/opesun/resolver"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type m map[string]interface{}

func main() {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	db := session.DB("resolver_test")
	coll1 := db.C("coll1")
	a := m{
		"a": 1,
		"b": 2,
	}
	a["_id"] = bson.NewObjectId()
	b := m{
		"d": 1,
		"f": 2,
	}
	b["_id"] = bson.NewObjectId()
	err = coll1.Insert(a)
	if err != nil {
		panic(err)
	}
	err = coll1.Insert(b)
	if err != nil {
		panic(err)
	}
	res := []interface{}{
		map[string]interface{}{
			"_coll1_single":   a["_id"],
			"_coll1_multiple": []interface{}{b["_id"], a["_id"], a["_id"]},
			"lol":             20,
		},
		map[string]interface{}{
			"_coll1_single":   b["_id"],
			"_coll1_multiple": []interface{}{b["_id"], b["_id"]},
			"lol":             21,
		},
	}
	resolver.ResolveAll(db, res, nil)
	for _, v := range res {
		fmt.Println(v, "\n")
	}
}
