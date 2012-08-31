package resolver_test

import(
	"testing"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"encoding/json"
	"github.com/opesun/resolver"
)

type m map[string]interface{}

func enc(a interface{}) string {
	res_b, err := json.Marshal(a)
	if err != nil { panic(err) }
	return string(res_b)
}

func TestSingle(t *testing.T) {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil { panic(err) }
	db := session.DB("resolver_test")
	coll1 := db.C("coll1")
	a := m{
		"name": "a",
		"a": 1,
		"b": 2,
	}
	a_id := bson.NewObjectId()
	a["_id"] = a_id
	if err != nil { panic(err) }
	unresolved := []interface{}{
		map[string]interface{}{
			"_coll1_single": a_id,
			"dummy_field": 42,
		},
	}
	resolved := []interface{}{
		map[string]interface{}{
			"_coll1_single": a,
			"dummy_field": 42,
		},
	}
	err = coll1.Insert(a)
	if err != nil { panic(err) }
	if enc(resolved) != enc(resolved) { t.Fatal("The test is inherently flawed.") }
	resolver.ResolveAll(db, unresolved, nil)
	if enc(unresolved) != enc(resolved) {
		t.Fatal("Resolved and unresolved values does not match.")
	}
}

//func TestMultiple(t *testing.T) {
//	
//}

func TestAll(t *testing.T) {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil { panic(err) }
	db := session.DB("test")
	coll1 := db.C("coll1")
	a := m{
		"name": "a",
		"a": 1,
		"b": 2,
	}
	a_id := bson.NewObjectId()
	a["_id"] = a_id
	b := m{
		"name": "b",
		"c": 3,
		"d": 4,
	}
	b_id := bson.NewObjectId()
	b["_id"] = b_id
	err = coll1.Insert(a)
	if err != nil { panic(err) }
	err = coll1.Insert(b)
	if err != nil { panic(err) }
	coll2 := db.C("coll2")
	c := m{
		"name": "c",
		"e": 4,
		"f": 6,
	}
	c_id := bson.NewObjectId()
	c["_id"] = c_id
	err = coll2.Insert(c)
	if err != nil { panic(err) }
	// Simulating a query result here.
	unresolved := []interface{}{
		map[string]interface{}{
			"_coll1_single": a_id,
			"_coll1_multiple": []interface{}{a_id, b_id},
			"dummy_field": 42,
		},
		map[string]interface{}{
			"_coll1_single": b_id,
			"_coll1_multiple": []interface{}{a_id, b_id, b_id},
			"dummy_field": 43,
		},
		map[string]interface{}{
			"_coll2_single": c_id,
			"_coll1_varied": []interface{}{a_id, 44, m{}, b_id, 45, a_id, 46},
		},
	}
	resolved := []interface{}{
		map[string]interface{}{
			"_coll1_single": a,
			"_coll1_multiple": []interface{}{a, b},
			"dummy_field": 42,
		},
		map[string]interface{}{
			"_coll1_single": b,
			"_coll1_multiple": []interface{}{a, b, b},
			"dummy_field": 43,
		},
		map[string]interface{}{
			"_coll2_single": c,
			"_coll1_varied": []interface{}{a, 44, m{}, b, 45, a, 46},
		},
	}
	// Testing if we can use JSON encryption to equality testing.
	if enc(resolved) != enc(resolved) { t.Fatal("The test is inherently flawed.") }
	resolver.ResolveAll(db, unresolved, nil)
	if enc(unresolved) != enc(resolved) {
		t.Fatal("Resolved and unresolved values does not match.")
	}
}