package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tmmweb/internal/media"
	"tmmweb/internal/store"
)

// State loading and migration live here so server construction can stay small
// and the scan/scrape handlers can share one cached item map.
func (s *Server) loadLibraries() error {
	libraries, err := s.store.Libraries()
	if err != nil {
		return err
	}
	s.libraries = libraries
	for i := range s.libraries {
		s.libraries[i].Paths = media.NormalizePaths(s.libraries[i])
		if len(s.libraries[i].Paths) > 0 {
			s.libraries[i].Path = s.libraries[i].Paths[0]
		}
	}
	return nil
}

func (s *Server) loadItems() error {
	items, err := s.store.Items()
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.DateAdded == "" {
			item.DateAdded = time.Now().UTC().Format(time.RFC3339)
		}
		if media.HasDetailedMediaInfo(item) {
			s.needsStoreCompaction = true
			item = media.CompactCachedItem(item)
		}
		s.items[item.ID] = item
	}
	return nil
}

func (s *Server) refreshCachedItems() {
	s.mu.Lock()
	items := make([]media.ScanExistingItem, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, media.NewScanExistingItem(item))
	}
	s.mu.Unlock()

	var batch []media.Item
	flush := func() {
		if len(batch) == 0 {
			return
		}
		_ = s.store.SaveItems(batch)
		batch = batch[:0]
	}

	for _, cached := range items {
		item := cached.ToItem()
		info, err := os.Stat(item.Path)
		if err != nil {
			continue
		}
		library := media.Library{
			ID:    item.LibraryID,
			Path:  item.SourcePath,
			Paths: []string{item.SourcePath},
			Type:  item.Kind,
		}
		local := media.NewItemFromFileInfo(library, item.SourcePath, item.Path, info)
		s.mu.Lock()
		existing, ok := s.items[item.ID]
		s.mu.Unlock()
		if !ok {
			continue
		}
		merged := media.CompactCachedItem(media.MergeScannedItem(existing, local))
		if !itemChanged(existing, merged) {
			continue
		}
		s.mu.Lock()
		s.items[merged.ID] = merged
		s.mu.Unlock()
		batch = append(batch, merged)
		if len(batch) >= scanPersistBatchSize {
			flush()
		}
	}
	flush()
}

func itemChanged(a media.Item, b media.Item) bool {
	return a.DateAdded != b.DateAdded ||
		a.TitleGuess != b.TitleGuess ||
		a.YearGuess != b.YearGuess ||
		a.Original != b.Original ||
		a.Overview != b.Overview ||
		a.Runtime != b.Runtime ||
		a.Rating != b.Rating ||
		a.ShowRating != b.ShowRating ||
		a.ModTimeUnix != b.ModTimeUnix ||
		a.DirModTimeUnix != b.DirModTimeUnix ||
		a.NFOModTimeUnix != b.NFOModTimeUnix ||
		a.FileSize != b.FileSize ||
		a.FileSizeBytes != b.FileSizeBytes ||
		a.VideoFormat != b.VideoFormat ||
		a.AudioCodec != b.AudioCodec ||
		a.MediaDurationSeconds != b.MediaDurationSeconds ||
		a.MediaInfoScanned != b.MediaInfoScanned ||
		strings.Join(a.Genres, "\x00") != strings.Join(b.Genres, "\x00") ||
		strings.Join(a.Actors, "\x00") != strings.Join(b.Actors, "\x00") ||
		a.Premiered != b.Premiered ||
		a.IMDBID != b.IMDBID ||
		a.MatchedID != b.MatchedID ||
		a.MatchedName != b.MatchedName ||
		a.HasNFO != b.HasNFO ||
		a.HasPoster != b.HasPoster ||
		a.HasFanart != b.HasFanart ||
		a.HasSubtitle != b.HasSubtitle
}

// compactStoredItems migrates older full stream payloads in bounded batches.
// The runtime cache was already compacted during loadItems, so this only
// reduces bbolt's persisted values and its memory-mapped file footprint.
func (s *Server) compactStoredItems() {
	s.mu.Lock()
	ids := make([]string, 0, len(s.items))
	for id := range s.items {
		ids = append(ids, id)
	}
	s.mu.Unlock()

	for start := 0; start < len(ids); start += scanPersistBatchSize {
		end := start + scanPersistBatchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := make([]media.Item, 0, end-start)
		s.mu.Lock()
		for _, id := range ids[start:end] {
			if item, ok := s.items[id]; ok {
				batch = append(batch, item)
			}
		}
		s.mu.Unlock()
		if err := s.store.SaveItems(batch); err != nil {
			return
		}
	}
	_, _ = s.store.CompactIfNeeded()
	releaseScanMemory()
}

func (s *Server) loadTasks() error {
	tasks, err := s.store.Tasks()
	if err != nil {
		return err
	}
	for _, record := range tasks {
		task := taskFromRecord(record)
		if task.State == "running" {
			task.State = "failed"
			task.Error = "task interrupted by server restart"
			task.FinishedAt = time.Now().Format(time.RFC3339)
			_ = s.store.SaveTask(task.toRecord())
		}
		s.tasks[task.ID] = task
	}
	return nil
}

func (s *Server) migrateJSON() error {
	if len(s.libraries) == 0 {
		if libraries, err := readJSONFile[[]media.Library](filepath.Join(s.config.DataDir, "libraries.json")); err == nil {
			for _, library := range libraries {
				library.Paths = media.NormalizePaths(library)
				if len(library.Paths) > 0 {
					library.Path = library.Paths[0]
				}
				if err := s.store.SaveLibrary(library); err != nil {
					return err
				}
				s.libraries = append(s.libraries, library)
			}
		}
	}
	if len(s.items) == 0 {
		if items, err := readJSONFile[[]media.Item](filepath.Join(s.config.DataDir, "items.json")); err == nil {
			if err := s.store.SaveItems(items); err != nil {
				return err
			}
			for _, item := range items {
				s.items[item.ID] = media.CompactCachedItem(item)
			}
		}
	}
	return nil
}

func (t *Task) toRecord() store.TaskRecord {
	return store.TaskRecord{
		ID:           t.ID,
		Type:         t.Type,
		LibraryID:    t.LibraryID,
		LibraryName:  t.LibraryName,
		State:        t.State,
		SourcePath:   t.SourcePath,
		CurrentPath:  t.CurrentPath,
		VisitedFiles: t.VisitedFiles,
		FoundItems:   t.FoundItems,
		ResultCount:  t.ResultCount,
		Error:        t.Error,
		StartedAt:    t.StartedAt,
		FinishedAt:   t.FinishedAt,
	}
}

func taskFromRecord(record store.TaskRecord) *Task {
	return &Task{
		ID:           record.ID,
		Type:         record.Type,
		LibraryID:    record.LibraryID,
		LibraryName:  record.LibraryName,
		State:        record.State,
		SourcePath:   record.SourcePath,
		CurrentPath:  record.CurrentPath,
		VisitedFiles: record.VisitedFiles,
		FoundItems:   record.FoundItems,
		ResultCount:  record.ResultCount,
		Error:        record.Error,
		StartedAt:    record.StartedAt,
		FinishedAt:   record.FinishedAt,
	}
}

func readJSONFile[T any](path string) (T, error) {
	var value T
	data, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	err = json.Unmarshal(data, &value)
	return value, err
}

func randomID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return strconv.FormatInt(int64(os.Getpid()), 16)
	}
	return hex.EncodeToString(buf[:])
}
