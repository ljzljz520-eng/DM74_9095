package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"

	"theatre39/internal/domain"
)

var ErrNotFound = errors.New("record not found")

var buckets = [][]byte{
	[]byte("batches"),
	[]byte("items"),
	[]byte("holds"),
	[]byte("reviews"),
	[]byte("archives"),
	[]byte("events"),
}

type DB struct {
	db *bbolt.DB
}

func Open(path string) (*DB, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	conn, err := bbolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := conn.Update(func(tx *bbolt.Tx) error {
		for _, name := range buckets {
			if _, createErr := tx.CreateBucketIfNotExists(name); createErr != nil {
				return createErr
			}
		}
		return nil
	}); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	return &DB{db: conn}, nil
}

func (d *DB) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.Close()
}

func encode(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode record: %w", err)
	}
	return data, nil
}

func decode(data []byte, target any) error {
	if len(data) == 0 {
		return ErrNotFound
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode record: %w", err)
	}
	return nil
}

func put(tx *bbolt.Tx, bucket, key string, value any) error {
	data, err := encode(value)
	if err != nil {
		return err
	}
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("bucket %s is missing", bucket)
	}
	return b.Put([]byte(key), data)
}

func get(tx *bbolt.Tx, bucket, key string, target any) error {
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("bucket %s is missing", bucket)
	}
	return decode(b.Get([]byte(key)), target)
}

func cloneBytes(data []byte) []byte {
	return append([]byte(nil), data...)
}

func list[T any](tx *bbolt.Tx, bucket string, decodeTarget func([]byte) (T, error)) ([]T, error) {
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return nil, fmt.Errorf("bucket %s is missing", bucket)
	}
	result := make([]T, 0)
	err := b.ForEach(func(_, value []byte) error {
		if value == nil {
			return nil
		}
		item, err := decodeTarget(cloneBytes(value))
		if err != nil {
			return err
		}
		result = append(result, item)
		return nil
	})
	return result, err
}

func (d *DB) SnapshotCounts() (map[string]int, error) {
	counts := make(map[string]int)
	err := d.db.View(func(tx *bbolt.Tx) error {
		for _, name := range buckets {
			bucket := tx.Bucket(name)
			if bucket != nil {
				counts[string(name)] = bucket.Stats().KeyN
			}
		}
		return nil
	})
	return counts, err
}

func (d *DB) RecordEvent(event domain.BatchEvent) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		return put(tx, "events", event.ID, event)
	})
}
