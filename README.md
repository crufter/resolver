resolver
========

Mongodb database references without DBRefs, and automatic resolution of them.

What does it do?
========
You pass in a *mgo.Database (db) and a query result (v) to
```
resolver.Resolve(db, v, nil)
```

And this package will resolve all "foreign key references" which value is an instance of bson.ObjectId, and which key has the form of:
```
"_collName"
// or
"_collName_customFieldName"	// In case of there are more than one reference to a given collection in a document.
```

The value can be a single bson.ObjectId or a []bson.ObjectId.
