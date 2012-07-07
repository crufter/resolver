// TODO do a test suite out of this.
package main

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/resolver"
	"fmt"
)

type m map[string]interface{}

func main() {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	db := session.DB("whatever")
	cont := db.C("content")
	ex := m{"a":1, "b":2}
	ex["_id"] = bson.NewObjectId()
	ex2 := m{"d":1, "f":2}
	ex2["_id"] = bson.NewObjectId()
	err = cont.Insert(ex)
	if err != nil {
		panic(err)
	}
	err2 := cont.Insert(ex2)
	if err2 != nil {
		panic(err2)
	}
	res := []map[string]interface{}{
		m{"_content_parent":ex["_id"], "_content_multiple": []interface{}{ex2["_id"], ex["_id"], ex["_id"]},  "lol": 20},
		m{"_content_parent":ex2["_id"], "_content_multiple": []interface{}{ex2["_id"], ex2["_id"]},  "lol": 20},
	}
	//fmt.Println(res)
	resolver.ResolveAll(db, res)
	for _, v := range res {
		fmt.Println(v, "\n")
	}
}