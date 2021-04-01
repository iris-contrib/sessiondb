# sessiondb

[![Go Report](https://goreportcard.com/badge/github.com/iris-contrib/sessiondb)](https://goreportcard.com/report/github.com/iris-contrib/sessiondb)

Dgraph &amp; MongoDB stores for `sessions` of [kataras/iris](https://github.com/kataras/go-sessions).

## How to use

Full examples can be found in the [examples](https://github.com/iris-contrib/sessiondb/tree/main/examples) folder.

### Mongo

[![Mongo Reference](https://pkg.go.dev/badge/github.com/iris-contrib/sessiondb/mongostore.svg)](https://pkg.go.dev/github.com/iris-contrib/sessiondb/mongostore) 

To include `mongostore` run 
```sh
go get github.com/iris-contrib/sessiondb/mongostore
```

#### Example

```go
	// replace with your running mongo server settings:
	cred := options.Credential{
		AuthSource: "admin",
		Username:   "user",
		Password:   "password",
	}

	clientOpts := options.Client().ApplyURI("mongodb://127.0.0.1:27017").SetAuth(cred)
	db, _ := mongostore.New(clientOpts, "sessions")

	sess := sessions.New(sessions.Config{Cookie: "sessionscookieid"})

	sess.UseDatabase(db)
```

### Dgraph

[![Dgraph Reference](https://pkg.go.dev/badge/github.com/iris-contrib/sessiondb/dgraphstore.svg)](https://pkg.go.dev/github.com/iris-contrib/sessiondb/dgraphstore) 

To include `dgraphstore` run 
```sh
go get github.com/iris-contrib/sessiondb/dgraphstore
```

#### Example 

```go
	// replace with your server settings:
	conn, _ := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	db, _ := dgraphstore.NewFromDB(conn)

	sess := sessions.New(sessions.Config{Cookie: "sessionscookieid"})

	sess.UseDatabase(db)
```

## Contribute 

Development of each store is done on branches. If you plan to work with an existing store checkout the corresponding branch. If you intent to implement a new store then create a new branch named after the DB you are using.

The repository is using [go-submodules](https://github.com/go-modules-by-example/index/tree/master/009_submodules). 

For releasing a new version of each individual store (submodule) you need to 
1. merge from corresponding branch to `master`
2. tag with appropriate store name and version ie. `git tag mongostore/v0.1.1`