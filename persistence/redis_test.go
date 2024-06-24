package persistence

import (
	"encoding/gob"
	"net"
	"testing"
	"time"
)

// These tests require redis server running on localhost:6379 (the default)
const redisTestServer = "localhost:6379"

var newRedisStore = func(t *testing.T, defaultExpiration time.Duration) CacheStore {
	c, err := net.Dial("tcp", redisTestServer)
	if err == nil {
		_, err := c.Write([]byte("flush_all\r\n"))
		if err != nil {
			t.Fatalf("Unexpected error: %s", err.Error())
		}
		c.Close()
		redisCache := NewRedisCache(redisTestServer, "", defaultExpiration)
		redisCache.Flush()
		return redisCache
	}
	t.Errorf("couldn't connect to redis on %s", redisTestServer)
	t.FailNow()
	panic("")
}

func TestRedisStoreSelectDatabase(t *testing.T) {
	c, err := net.Dial("tcp", redisTestServer)
	if err != nil {
		t.Errorf("couldn't connect to redis on %s", redisTestServer)
	}
	_, err = c.Write([]byte("flush_all\r\n"))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	c.Close()
	redisCache := NewRedisCache(redisTestServer, "", 1*time.Second, WithSelectDatabase(1))
	err = redisCache.Flush()
	if err != nil {
		t.Errorf("couldn't connect to redis on %s", redisTestServer)
	}
}
func TestRedisCache_MgetTwoKeys(t *testing.T) {
	simpleMgetTwoKeys(t, newRedisStore)
}

func TestRedisCache_MsetNXTwoKeys(t *testing.T) {
	simpleMsetNXTwoKeys(t, newRedisStore)
}

func TestRedisCache_MsetNXThenMgetThreeKeys(t *testing.T) {
	msetNXThenMgetThreeKeys(t, newRedisStore)
}

func TestRedisCache_TypicalGetSet(t *testing.T) {
	typicalGetSet(t, newRedisStore)
}

func TestRedisCache_IncrDecr(t *testing.T) {
	incrDecr(t, newRedisStore)
}
func TestRedis_IncrAtomic(t *testing.T) {
	incrAtomic(t, newRawRedisStore)
}

func TestRedis_GetExpiresIn(t *testing.T) {
	getExpiresIn(t, newRawRedisStore)
}

func TestRedisCache_Expiration(t *testing.T) {
	expiration(t, newRedisStore)
}

func TestRedisCache_EmptyCache(t *testing.T) {
	emptyCache(t, newRedisStore)
}

func TestRedisCache_Replace(t *testing.T) {
	testReplace(t, newRedisStore)
}

func TestRedisCache_Add(t *testing.T) {
	testAdd(t, newRedisStore)
}

// The following tests are specific to RedisStore.
func simpleMgetTwoKeys(t *testing.T, newCache cacheFactory) {
	cache := newCache(t, time.Hour).(*RedisStore)
	// set two keys and make sure set is successful
	value := "foo1"
	err := cache.Set("test1", value, DEFAULT)
	if err != nil {
		t.Errorf("Error setting a value: %s", err)
	}
	value = ""
	err = cache.Get("test1", &value)
	if err != nil {
		t.Errorf("Error getting a value: %s", err)
	}
	if value != "foo1" {
		t.Errorf("Expected to get foo back, got %s", value)
	}
	value = "foo2"
	if err = cache.Set("test2", value, DEFAULT); err != nil {
		t.Errorf("Error setting a value: %s", err)
	}
	value = ""
	err = cache.Get("test2", &value)
	if err != nil {
		t.Errorf("Error getting a value: %s", err)
	}
	if value != "foo2" {
		t.Errorf("Expected to get foo back, got %s", value)
	}

	// mget and verify
	s1 := ""
	s2 := ""
	result := []interface{}{&s1, &s2}
	err = cache.Mget(result, "test1", "test2")
	if err != nil {
		t.Errorf("Error while doing mget: %v", err)
	}
	if s1 != "foo1" {
		t.Errorf("Expected to get foo1 for key test1, got %v", s1)
	}
	if s2 != "foo2" {
		t.Errorf("Expected to get foo2 for key test2, got %v", s2)
	}
	// shows another way to get the value
	t.Logf("test1: %v, test2: %v, err: %v", *(result[0].(*string)), *(result[1].(*string)), err)
}

func simpleMsetNXTwoKeys(t *testing.T, newCache cacheFactory) {
	cache := newCache(t, time.Hour).(*RedisStore)
	k1 := "test1"
	v1 := "value2"
	k2 := "test2"
	v2 := "value2"

	if err := cache.Delete(k1); err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	if err := cache.Delete(k2); err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	// mset two keys
	err := cache.MSetNX(time.Hour, k1, v1, k2, v2)
	if err != nil {
		t.Errorf("Error mset: %v", err)
	}

	// verify values
	value := ""
	err = cache.Get(k1, &value)
	if err != nil {
		t.Errorf("Error getting a key %s: %v", k1, err)
	}
	if value != v1 {
		t.Errorf("Error getting value for key %s. Got %v, expect %v", k1, value, v1)
	}
	value = ""
	err = cache.Get(k2, &value)
	if err != nil {
		t.Errorf("Error getting a key %s: %v", k2, err)
	}
	if value != v2 {
		t.Errorf("Error getting value for key %s. Got %v, expect %v", k2, value, v2)
	}
}

func msetNXThenMgetThreeKeys(t *testing.T, newCache cacheFactory) {
	cache := newCache(t, time.Hour).(*RedisStore)
	k1 := "test1"
	v1 := "value2"
	k2 := "test2"
	v2 := "value2"
	k3 := "test3"
	v3 := "value3"

	if err := cache.Delete(k1); err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	if err := cache.Delete(k2); err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	if err := cache.Delete(k3); err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	// mset two keys
	err := cache.MSetNX(time.Hour, k1, v1, k2, v2, k3, v3)
	if err != nil {
		t.Errorf("Error mset: %v", err)
	}

	r1 := ""
	r2 := ""
	r3 := ""
	result := []interface{}{&r1, &r2, &r3}
	err = cache.Mget(result, k1, k2, k3)
	if err != nil {
		t.Errorf("Error while doing mget: %v", err)
	}
	if r1 != v1 {
		t.Errorf("Expected to get %v for key %v, got %v", v1, k1, r1)
	}
	if r2 != v2 {
		t.Errorf("Expected to get %v for key %v, got %v", v2, k2, r2)
	}
	if r3 != v3 {
		t.Errorf("Expected to get %v for key %v, got %v", v3, k3, r3)
	}
}

// Redis Hash Sets
type aType struct {
	Count int
	Str string
}
func init() {
	gob.Register(aType{})
}
type rHash struct {
	Field1 int
	Field2 string
	Field3 aType
	field4 string
}

func TestRedisCache_HSET_HGET_HGETALL(t *testing.T) {
	var err error
	exp := time.Duration(5)*time.Second
	store := newRawRedisStore(t, exp)

	testData := rHash{
		Field1: 19,
		Field2: "f2 value",
		Field3: aType{Count: 1,Str: "hi"},
		field4: "unexported",
	}
	numFieldsCreated, err := store.HSet("hset1", exp, &testData)
	if err != nil {
		t.Errorf("Expected no error for Hset: %v", err)
	}
	if numFieldsCreated != 3 {
		t.Errorf("Expected 3 fields created, got %v", numFieldsCreated)
	}

	var intValue int
	err = store.HGet("hset1", "Field1", &intValue)
	if err != nil {
		t.Errorf("Failed to get a field hset1:Field1: %v", err)
	}
	if intValue != 19 {
		t.Errorf("Get Value: %v doesn't match what was set", intValue)
	}

	getAllHash := rHash{}
	err = store.HGetAll("hset1", &getAllHash)
	if err != nil {
		t.Errorf("Failed to get all fields for key: hset1: %v", err)
	}
	if getAllHash.Field1 != 19 {
		t.Errorf("Expected 19 got %v", getAllHash.Field1)
	}
	if getAllHash.Field2 != "f2 value" {
		t.Errorf("Expected f2 value got %v", getAllHash.Field2)
	}
	a := aType{Count: 1,Str: "hi"}
	if getAllHash.Field3 != a {
		t.Errorf("Expected a struct with values %v", getAllHash.Field3)
	}

	getAllHash2 := rHash{}
	err = store.HGetAll("notthere", &getAllHash2)
	if err != ErrCacheMiss {
		t.Errorf("Expected a cache miss of non existant key for getall got: %v", err)
	}
}

func TestRedisCache_HEXISTS_HDEL(t *testing.T) {
	var err error
	exp := time.Duration(5)*time.Second
	store := newRawRedisStore(t, exp)

	testData := rHash{
		Field1: 19,
		Field2: "f2 value",
		Field3: aType{Count: 1,Str: "hi"},
		field4: "unexported",
	}
	numFieldsCreated, err := store.HSet("hset2", exp, &testData)
	if err != nil {
		t.Errorf("Expected no error for Hset: %v", err)
	}
	if numFieldsCreated != 3 {
		t.Errorf("Expected 3 fields created, got %v", numFieldsCreated)
	}

	f1Exists, err := store.HExists("hset2", "Field1")
	if err != nil {
		t.Errorf("Expected no error from hexists: %v", err)
	}
	if !f1Exists {
		t.Errorf("Field1 Expected to exist")
	}
	numDel, err := store.HDel("hset2", "Field1", "Field2", "NotAField")
	if err != nil {
		t.Errorf("Expected no error from hdel: %v", err)
	}
	if numDel != 2 {
		t.Errorf("Hdel had incorrect number deleted: %v", err)
	}
	f1Exists, err = store.HExists("hset2", "Field1")
	if err != nil {
		t.Errorf("Expected no error from hexists: %v", err)
	}
	if f1Exists {
		t.Errorf("Field1 Expected to have been deleted")
	}
}

func TestRedisCache_HKEYS(t *testing.T) {
	var err error
	exp := time.Duration(5)*time.Second
	store := newRawRedisStore(t, exp)

	testData := rHash{
		Field1: 19,
		Field2: "f2 value",
		Field3: aType{Count: 1,Str: "hi"},
		field4: "unexported",
	}
	numFieldsCreated, err := store.HSet("hset3", exp, &testData)
	if err != nil {
		t.Errorf("Expected no error for Hset: %v", err)
	}
	if numFieldsCreated != 3 {
		t.Errorf("Expected 3 fields created, got %v", numFieldsCreated)
	}

	fieldNames, err := store.HKeys("hset3")
	if err != nil {
		t.Errorf("expected hkeys to succeed: %v", err)
	}
	if len(fieldNames) != 3 {
		t.Errorf("Expected 3 fields, got: %v", len(fieldNames))
	}
	for _, field := range fieldNames {
		if !stringInSlice(field, []string{"Field1", "Field2", "Field3"}) {
			t.Errorf("Found field shouldn't exist: %s", field)
		}
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func TestRedisCache_HLEN(t *testing.T) {
	var err error
	exp := time.Duration(5)*time.Second
	store := newRawRedisStore(t, exp)

	testData := rHash{
		Field1: 19,
		Field2: "f2 value",
		Field3: aType{Count: 1,Str: "hi"},
		field4: "unexported",
	}
	numFieldsCreated, err := store.HSet("hset4", exp, &testData)
	if err != nil {
		t.Errorf("Expected no error for Hset: %v", err)
	}
	if numFieldsCreated != 3 {
		t.Errorf("Expected 3 fields created, got %v", numFieldsCreated)
	}

	hLen, err := store.HLen("hset4")
	if err != nil {
		t.Errorf("expected hlen to succeed: %v", err)
	}
	if hLen != 3 {
		t.Errorf("Expected hlen to be 3, got: %v", hLen)
	}
}

func TestRedisCache_HINCRBY(t *testing.T) {
	var err error
	exp := time.Duration(5)*time.Second
	store := newRawRedisStore(t, exp)

	testData := rHash{
		Field1: 19,
		Field2: "f2 value",
		Field3: aType{Count: 1,Str: "hi"},
		field4: "unexported",
	}
	numFieldsCreated, err := store.HSet("hset5", exp, &testData)
	if err != nil {
		t.Errorf("Expected no error for Hset: %v", err)
	}
	if numFieldsCreated != 3 {
		t.Errorf("Expected 3 fields created, got %v", numFieldsCreated)
	}

	var intValue int
	err = store.HGet("hset5", "Field1", &intValue)
	if err != nil {
		t.Errorf("Failed to get a field hset5:Field1: %v", err)
	}
	if intValue != 19 {
		t.Errorf("Get Value: %v doesn't match what was set", intValue)
	}

	nVal, err := store.HIncrBy("hset5", "Field1", 1)
	if err != nil {
		t.Errorf("Expected no error for Hincrby: %v", err)
	}
	if nVal != 20 {
		t.Errorf("Increment by 1 should be 20 got: %v", nVal)
	}

	var nintValue int
	err = store.HGet("hset5", "Field1", &nintValue)
	if err != nil {
		t.Errorf("Failed to get a field hset5:Field1: %v", err)
	}
	if nintValue != 20 {
		t.Errorf("Get Value: %v doesn't match what was incremented", nintValue)
	}
}
