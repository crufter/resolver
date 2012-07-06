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
	err = cont.Insert(ex)
	if err != nil {
		panic(err)
	}
	res := m{"_content_parent":ex["_id"], "_content_multiple": []interface{}{ex["_id"], ex["_id"], ex["_id"]},  "lol": 20}
	fmt.Println(res)
	resolver.ResolveOne(db, res)
	fmt.Println(res)
}