package store

import (
	"encoding/json"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"

	"tmmweb/internal/media"
)

var (
	librariesBucket = []byte("libraries")
	itemsBucket     = []byte("items")
	tasksBucket     = []byte("tasks")
)

type TaskRecord struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	LibraryID    string `json:"libraryId"`
	LibraryName  string `json:"libraryName"`
	State        string `json:"state"`
	SourcePath   string `json:"sourcePath,omitempty"`
	CurrentPath  string `json:"currentPath,omitempty"`
	VisitedFiles int    `json:"visitedFiles"`
	FoundItems   int    `json:"foundItems"`
	ResultCount  int    `json:"resultCount"`
	Error        string `json:"error,omitempty"`
	StartedAt    string `json:"startedAt"`
	FinishedAt   string `json:"finishedAt,omitempty"`
}

type Store struct {
	db *bolt.DB
}

func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	db, err := bolt.Open(filepath.Join(dataDir, "tmmweb.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) init() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range [][]byte{librariesBucket, itemsBucket, tasksBucket} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) Libraries() ([]media.Library, error) {
	var libraries []media.Library
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(librariesBucket).ForEach(func(_, value []byte) error {
			var library media.Library
			if err := json.Unmarshal(value, &library); err != nil {
				return err
			}
			libraries = append(libraries, library)
			return nil
		})
	})
	return libraries, err
}

func (s *Store) SaveLibrary(library media.Library) error {
	return putJSON(s.db, librariesBucket, library.ID, library)
}

func (s *Store) DeleteLibrary(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(librariesBucket).Delete([]byte(id))
	})
}

func (s *Store) Items() ([]media.Item, error) {
	var items []media.Item
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(itemsBucket).ForEach(func(_, value []byte) error {
			var item media.Item
			if err := json.Unmarshal(value, &item); err != nil {
				return err
			}
			items = append(items, item)
			return nil
		})
	})
	return items, err
}

func (s *Store) SaveItem(item media.Item) error {
	return putJSON(s.db, itemsBucket, item.ID, item)
}

func (s *Store) SaveItems(items []media.Item) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(itemsBucket)
		for _, item := range items {
			data, err := json.Marshal(item)
			if err != nil {
				return err
			}
			if err := bucket.Put([]byte(item.ID), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) PruneLibraryItems(libraryID string, items []media.Item) error {
	current := make(map[string]struct{}, len(items))
	for _, item := range items {
		current[item.ID] = struct{}{}
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(itemsBucket)
		var staleKeys [][]byte
		if err := bucket.ForEach(func(key, value []byte) error {
			var item media.Item
			if err := json.Unmarshal(value, &item); err != nil {
				return err
			}
			if item.LibraryID == libraryID {
				if _, ok := current[item.ID]; !ok {
					staleKeys = append(staleKeys, append([]byte(nil), key...))
				}
			}
			return nil
		}); err != nil {
			return err
		}
		for _, key := range staleKeys {
			if err := bucket.Delete(key); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) DeleteLibraryItems(libraryID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(itemsBucket)
		var staleKeys [][]byte
		if err := bucket.ForEach(func(key, value []byte) error {
			var item media.Item
			if err := json.Unmarshal(value, &item); err != nil {
				return err
			}
			if item.LibraryID == libraryID {
				staleKeys = append(staleKeys, append([]byte(nil), key...))
			}
			return nil
		}); err != nil {
			return err
		}
		for _, key := range staleKeys {
			if err := bucket.Delete(key); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) Tasks() ([]TaskRecord, error) {
	var tasks []TaskRecord
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(tasksBucket).ForEach(func(_, value []byte) error {
			var task TaskRecord
			if err := json.Unmarshal(value, &task); err != nil {
				return err
			}
			tasks = append(tasks, task)
			return nil
		})
	})
	return tasks, err
}

func (s *Store) SaveTask(task TaskRecord) error {
	return putJSON(s.db, tasksBucket, task.ID, task)
}

func putJSON(db *bolt.DB, bucketName []byte, key string, value interface{}) error {
	return db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return tx.Bucket(bucketName).Put([]byte(key), data)
	})
}
