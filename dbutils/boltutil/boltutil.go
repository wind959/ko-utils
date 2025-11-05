package boltutil

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"go.etcd.io/bbolt"
	"sync"
	"time"
)

// 全局唯一的数据库实例
var (
	dbInstance *bbolt.DB    // 全局 store 实例
	once       sync.Once    // 确保初始化只执行一次
	mu         sync.RWMutex // 用于并发控制
)

// BoltConfig 数据库配置
type BoltConfig struct {
	Path    string        // 数据库文件路径
	Timeout time.Duration // 连接超时时间
	Options *bbolt.Options
}

// GetDBInstance 获取全局唯一的数据库实例(线程安全)
func GetDBInstance(cfg BoltConfig) (*bbolt.DB, error) {
	var initErr error
	once.Do(func() {
		mu.Lock()
		defer mu.Unlock()

		// 设置默认选项
		if cfg.Options == nil {
			cfg.Options = &bbolt.Options{
				Timeout:  cfg.Timeout,
				ReadOnly: false,
			}
		}
		dbInstance, initErr = bbolt.Open(cfg.Path, 0600, cfg.Options)
	})
	return dbInstance, initErr
}

// Close 关闭数据库连接
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if dbInstance == nil {
		return errors.New("database is not open")
	}

	err := dbInstance.Close()
	dbInstance = nil
	return err
}

// CreateBucket 创建存储桶
func CreateBucket(bucketName []byte) error {
	db, err := GetDBInstance(BoltConfig{})
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})

}

// Put 存储数据(自动序列化)
func Put(bucketName, key []byte, value interface{}) error {
	db, err := GetDBInstance(BoltConfig{})
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return fmt.Errorf("encoding failed: %v", err)
	}

	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		return b.Put(key, buf.Bytes())
	})
}

// Get 获取数据(自动反序列化)
func Get(bucketName, key []byte, value interface{}) error {
	db, err := GetDBInstance(BoltConfig{})
	if err != nil {
		return err
	}

	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		data := b.Get(key)
		if data == nil {
			return errors.New("key not found")
		}

		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		return dec.Decode(value)
	})
}

// Delete 删除数据
func Delete(bucketName, key []byte) error {
	db, err := GetDBInstance(BoltConfig{})
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		return b.Delete(key)
	})
}

// ForEach 遍历存储桶中的所有键值对
func ForEach(bucketName []byte, fn func(k, v []byte) error) error {
	db, err := GetDBInstance(BoltConfig{})
	if err != nil {
		return err
	}

	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		return b.ForEach(fn)
	})
}

// Backup 备份数据库
func Backup(path string) error {
	db, err := GetDBInstance(BoltConfig{})
	if err != nil {
		return err
	}

	return db.View(func(tx *bbolt.Tx) error {
		return tx.CopyFile(path, 0600)
	})
}

// Stats 获取数据库统计信息
func Stats() bbolt.Stats {
	db, err := GetDBInstance(BoltConfig{})
	if err != nil {
		return bbolt.Stats{}
	}
	return db.Stats()
}
