package dbutils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/wind959/ko-utils/dbutils/boltutil"
	"go.etcd.io/bbolt"
	"log"
	"testing"
	"time"
)

type User struct {
	Name  string
	Email string
	Age   int
}

type Order struct {
	ItemName string
	Price    float64
	Unit     int
}

func TestBolt(t *testing.T) {

	cfg := boltutil.BoltConfig{
		Options: &bbolt.Options{
			NoSync: true, // 生产环境建议设为false
		},
		Path:    "bolt_test.db",
		Timeout: 5 * time.Second,
	}
	_, err := boltutil.GetDBInstance(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := boltutil.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	//=======================================================================
	// 2. 创建存储桶 User
	bucketUser := []byte("Users")
	if err := boltutil.CreateBucket(bucketUser); err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}
	// 3. 存储数据
	user := User{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}
	user1 := User{
		Name:  "bob",
		Email: "bob@example.com",
		Age:   30,
	}
	if err := boltutil.Put(bucketUser, []byte("alice"), user); err != nil {
		t.Fatalf("Failed to store user: %v", err)
	}
	if err := boltutil.Put(bucketUser, []byte("bob"), user1); err != nil {
		t.Fatalf("Failed to store user: %v", err)
	}

	// 获取用户数据
	var retrievedUser User
	if err := boltutil.Get(bucketUser, []byte("alice"), &retrievedUser); err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	t.Log(retrievedUser)
	//=======================================================================

	// 2 创建存储桶 Order
	bucketOrder := []byte("Orders")
	if err := boltutil.CreateBucket(bucketOrder); err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	order := Order{
		ItemName: "新三国",
		Price:    19.99,
		Unit:     5,
	}
	if err := boltutil.Put(bucketOrder, []byte("Book"), order); err != nil {
		t.Fatalf("Failed to store order: %v", err)
	}

	var retrievedOrder Order
	if err := boltutil.Get(bucketOrder, []byte("Book"), &retrievedOrder); err != nil {
		t.Fatalf("Failed to get order: %v", err)
	}
	t.Log(retrievedOrder)

	//=======================================================================
	// 遍历存储桶
	err = boltutil.ForEach(bucketUser, func(k, v []byte) error {
		var u User
		if err := gob.NewDecoder(bytes.NewBuffer(v)).Decode(&u); err != nil {
			return err
		}
		fmt.Printf("Key: %s, Name: %s, Email: %s\n", k, u.Name, u.Email)
		return nil
	})
	if err != nil {
		log.Printf("Error iterating bucket: %v", err)
	}
}
