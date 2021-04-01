package mongostore

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/kataras/iris/v12/sessions"

	"github.com/kataras/golog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var errDatabaseNameMissing = errors.New("database name is required")

// Database the BoltDB(file-based) session storage.
type Database struct {
	// mongo *mongo.Database
	Service *mongo.Database
	logger  *golog.Logger
}

// New creates and returns a new MongoDB(file-based) storage with custom client options.
// Database and collection names should be included.
//
// It will remove any old session files.
func New(clientOpts *options.ClientOptions, database string) (*Database, error) {
	if database == "" {
		return nil, errDatabaseNameMissing
	}

	ctx := context.Background()
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	mongo := client.Database(database)
	return &Database{Service: mongo}, nil
}

// SetLogger sets the logger once before server ran.
// By default the Iris one is injected.
func (db *Database) SetLogger(logger *golog.Logger) {
	db.logger = logger
}

// Acquire receives a session's lifetime from the database,
// if the return value is LifeTime{} then the session manager sets the life time based on the expiration duration lives in configuration.
func (db *Database) Acquire(sid string, expires time.Duration) sessions.LifeTime {
	var result bson.Raw
	ctx := context.TODO()
	res := db.Service.Collection(sid).FindOne(ctx, bson.D{{"key", sid}})

	// not found, create an entry and return an empty lifetime, session manager will do its job.
	if err := res.Err(); err != nil {
		expirationTime := time.Now().Add(expires)
		timeBytes, _ := sessions.DefaultTranscoder.Marshal(expirationTime)
		timeBase := base64.StdEncoding.EncodeToString(timeBytes)
		db.Service.Collection(sid).InsertOne(
			context.TODO(),
			bson.D{{"$set", bson.D{{"key", sid}, {"value", timeBase}}}},
		)

		return sessions.LifeTime{Time: sessions.CookieExpireDelete}
	}

	// found, return the expiration.
	res.Decode(&result)
	result.Validate()
	val := result.Lookup("value")
	var expirationTime time.Time
	valueBase, _ := base64.StdEncoding.DecodeString(val.StringValue())
	sessions.DefaultTranscoder.Unmarshal(valueBase, &expirationTime)
	return sessions.LifeTime{Time: expirationTime}
}

// OnUpdateExpiration not implemented here, yet.
// Note that this error will not be logged, callers should catch it manually.
func (db *Database) OnUpdateExpiration(sid string, newExpires time.Duration) error {
	return sessions.ErrNotImplemented
}

// Set sets a key value of a specific session.
// Ignore the "immutable".
func (db *Database) Set(sid string, lifetime sessions.LifeTime, key string, value interface{}, immutable bool) {
	valueBytes, err := sessions.DefaultTranscoder.Marshal(value)
	if err != nil {
		return
	}

	// convert []byte slice to base64 string
	valueBase := base64.StdEncoding.EncodeToString(valueBytes)

	db.Service.Collection(sid).UpdateOne(
		context.Background(),
		// filter
		bson.D{{"key", key}},
		// update
		bson.D{{"$set", bson.D{{"key", key}, {"value", valueBase}}}},
		// options
		options.Update().SetUpsert(true),
	)
}

// Get retrieves a session value based on the key.
func (db *Database) Get(sid string, key string) (value interface{}) {
	if err := db.Decode(sid, key, &value); err == nil {
		return value
	}

	return nil
}

// Decode binds the "outPtr" to the value associated to the provided "key".
func (db *Database) Decode(sid, key string, outPtr interface{}) error {
	var result bson.Raw
	ctx := context.TODO()
	res := db.Service.Collection(sid).FindOne(ctx, bson.D{{"key", key}})

	err := res.Decode(&result)
	if err != nil {
		return err
	}

	err = result.Validate()
	if err != nil {
		return err
	}

	val := result.Lookup("value")
	valueBase, _ := base64.StdEncoding.DecodeString(val.StringValue())
	sessions.DefaultTranscoder.Unmarshal(valueBase, outPtr)
	return nil
}

// Visit loops through all session keys and values.
func (db *Database) Visit(sid string, cb func(key string, value interface{})) {
	ctx := context.TODO()
	res, err := db.Service.Collection(sid).Find(ctx, bson.D{})
	if err != nil {
		return
	}

	for res.Next(context.TODO()) {
		var result bson.Raw
		if err := res.Decode(&result); err != nil {
			log.Fatal(err)
		}

		k := result.Lookup("key")
		v := result.Lookup("value")
		var val interface{}
		valueBase, _ := base64.StdEncoding.DecodeString(v.StringValue())
		sessions.DefaultTranscoder.Unmarshal(valueBase, &val)
		cb(k.String(), val)
	}
	if err := res.Err(); err != nil {
		log.Fatal(err)
	}
}

// Len returns the length of the session's entries (keys).
func (db *Database) Len(sid string) (n int) {
	ctx := context.TODO()
	number, err := db.Service.Collection(sid).CountDocuments(ctx, bson.D{})
	if err == nil {
		n = int(number)
	}

	return
}

// Delete removes a session key value based on its key.
func (db *Database) Delete(sid string, key string) (deleted bool) {
	ctx := context.TODO()
	_, err := db.Service.Collection(sid).DeleteOne(ctx, bson.D{{"key", key}})
	if err != nil {
		deleted = false
		return
	}
	deleted = true
	return
}

// Clear removes all session key values but it keeps the session entry.
func (db *Database) Clear(sid string) {
	db.Service.Collection(sid).DeleteMany(context.TODO(), bson.D{{"key", bson.D{{"$ne", sid}}}})
}

// Release destroys the session, it clears and removes the session entry,
// session manager will create a new session ID on the next request after this call.
func (db *Database) Release(sid string) {
	db.Service.Collection(sid).Drop(context.TODO())
}

// Close terminates Dgraph's gRPC connection.
func (db *Database) Close() error {
	db.Service.Client().Disconnect(context.TODO())
	return nil
}
