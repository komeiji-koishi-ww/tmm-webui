package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

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
	mu   sync.Mutex
	db   *bolt.DB
	path string
}

func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	db, err := bolt.Open(filepath.Join(dataDir, "tmmweb.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db, path: filepath.Join(dataDir, "tmmweb.db")}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	return putJSON(s.db, librariesBucket, library.ID, library)
}

func (s *Store) DeleteLibrary(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(librariesBucket).Delete([]byte(id))
	})
}

func (s *Store) Items() ([]media.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	item = media.CompactCachedItem(item)
	return putJSON(s.db, itemsBucket, item.ID, item)
}

func (s *Store) SaveItems(items []media.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(itemsBucket)
		for _, item := range items {
			item = media.CompactCachedItem(item)
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

func (s *Store) DeleteItem(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(itemsBucket).Delete([]byte(id))
	})
}

func (s *Store) PruneLibraryItems(libraryID string, items []media.Item) error {
	currentIDs := make([]string, 0, len(items))
	for _, item := range items {
		currentIDs = append(currentIDs, item.ID)
	}
	return s.PruneLibraryItemIDs(libraryID, currentIDs)
}

func (s *Store) PruneLibraryItemIDs(libraryID string, currentIDs []string) error {
	current := make(map[string]struct{}, len(currentIDs))
	for _, id := range currentIDs {
		current[id] = struct{}{}
	}
	return s.PruneLibraryItemIDSet(libraryID, current)
}

func (s *Store) PruneLibraryItemIDSet(libraryID string, current map[string]struct{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	return putJSON(s.db, tasksBucket, task.ID, task)
}

// CompactIfNeeded rewrites bbolt only after updates have left a meaningful
// amount of free space. bbolt maps the complete database file, so reclaiming
// free pages reduces both disk use and the resident mapping on NAS systems.
func (s *Store) CompactIfNeeded() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stat, err := os.Stat(s.path)
	if err != nil {
		return false, err
	}
	stats := s.db.Stats()
	const minFreeBytes = 8 << 20
	if stat.Size() < minFreeBytes || int64(stats.FreeAlloc) < minFreeBytes || int64(stats.FreeAlloc)*4 < stat.Size() {
		return false, nil
	}

	temporary, err := os.CreateTemp(filepath.Dir(s.path), ".tmmweb-compact-*.db")
	if err != nil {
		return false, err
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return false, err
	}
	defer os.Remove(temporaryPath)

	compacted, err := bolt.Open(temporaryPath, 0600, nil)
	if err != nil {
		return false, err
	}
	if err := bolt.Compact(compacted, s.db, 8<<20); err != nil {
		_ = compacted.Close()
		return false, err
	}
	if err := compacted.Sync(); err != nil {
		_ = compacted.Close()
		return false, err
	}
	if err := compacted.Close(); err != nil {
		return false, err
	}
	if err := s.db.Close(); err != nil {
		return false, err
	}
	backupPath := s.path + ".precompact"
	_ = os.Remove(backupPath)
	if err := os.Rename(s.path, backupPath); err != nil {
		if reopenErr := s.reopenDatabase(); reopenErr != nil {
			return false, fmt.Errorf("preserve original database: %w (reopen: %v)", err, reopenErr)
		}
		return false, err
	}
	if err := os.Rename(temporaryPath, s.path); err != nil {
		_ = os.Rename(backupPath, s.path)
		if reopenErr := s.reopenDatabase(); reopenErr != nil {
			return false, fmt.Errorf("install compacted database: %w (reopen: %v)", err, reopenErr)
		}
		return false, err
	}
	db, err := bolt.Open(s.path, 0600, nil)
	if err != nil {
		_ = os.Remove(s.path)
		_ = os.Rename(backupPath, s.path)
		if reopenErr := s.reopenDatabase(); reopenErr != nil {
			return false, fmt.Errorf("open compacted database: %w (restore: %v)", err, reopenErr)
		}
		return false, err
	}
	s.db = db
	_ = os.Remove(backupPath)
	return true, nil
}

func (s *Store) reopenDatabase() error {
	db, err := bolt.Open(s.path, 0600, nil)
	if err != nil {
		return err
	}
	s.db = db
	return nil
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
