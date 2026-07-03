package app

import (
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"tmmweb/internal/media"
)

// Task is the API-facing scan task model. It is also converted to the store
// record model in state.go for persistence across restarts.
type Task struct {
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

// runScanTask streams scan results into memory and persists in small batches so
// the UI can show media incrementally without waiting for a full library pass.
func (s *Server) runningScanTaskLocked(libraryID string) *Task {
	for _, task := range s.tasks {
		if task.Type == "scan" && task.LibraryID == libraryID && task.State == "running" {
			return task
		}
	}
	return nil
}

func (s *Server) runScanTask(taskID string, library media.Library) {
	defer releaseScanMemory()

	var pending []media.Item
	var persistErr error
	lastFlush := time.Now()
	flushPending := func(force bool) error {
		if len(pending) == 0 {
			return nil
		}
		if !force && len(pending) < scanPersistBatchSize && time.Since(lastFlush) < scanPersistFlushInterval {
			return nil
		}
		batch := append([]media.Item(nil), pending...)
		if err := s.store.SaveItems(batch); err != nil {
			pending = nil
			return err
		}
		pending = pending[:0]
		lastFlush = time.Now()
		return nil
	}

	shouldCancel := func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		task := s.tasks[taskID]
		return task != nil && task.State == "canceling"
	}
	s.mu.Lock()
	existing := make(map[string]media.Item)
	for id, item := range s.items {
		if item.LibraryID == library.ID {
			existing[id] = item
		}
	}
	s.mu.Unlock()
	currentIDs, err := media.ScanLibraryIDsWithOptions(library, media.ScanOptions{
		Existing:          existing,
		ProbeMediaInfo:    scanMediaInfoEnabled(),
		SkipUnchangedDirs: scanSkipUnchangedDirsEnabled(),
	}, func(progress media.ScanProgress) {
		s.mu.Lock()
		if task := s.tasks[taskID]; task != nil {
			task.SourcePath = progress.SourcePath
			task.CurrentPath = progress.CurrentPath
			task.VisitedFiles = progress.VisitedFiles
			task.FoundItems = progress.FoundItems
		}
		if progress.Item != nil {
			item := *progress.Item
			if existing, ok := s.items[item.ID]; ok {
				item = media.MergeScannedItem(existing, item)
			}
			s.items[item.ID] = item
			if persistErr == nil {
				pending = append(pending, item)
			}
		}
		s.mu.Unlock()
		if progress.Item != nil && persistErr == nil {
			persistErr = flushPending(false)
		}
	}, shouldCancel)
	if err == nil && persistErr != nil {
		err = persistErr
	}
	if err == nil {
		err = flushPending(true)
	}
	existing = nil
	s.mu.Lock()
	defer s.mu.Unlock()
	task := s.tasks[taskID]
	if task == nil {
		return
	}
	task.FinishedAt = time.Now().Format(time.RFC3339)
	if err == media.ErrScanCanceled {
		task.State = "canceled"
		task.Error = ""
		task.ResultCount = task.FoundItems
		_ = s.store.SaveTask(task.toRecord())
		return
	}
	if err != nil {
		task.State = "failed"
		task.Error = err.Error()
		_ = s.store.SaveTask(task.toRecord())
		return
	}
	if err := s.store.PruneLibraryItemIDs(library.ID, currentIDs); err != nil {
		task.State = "failed"
		task.Error = err.Error()
		_ = s.store.SaveTask(task.toRecord())
		return
	}
	currentIDSet := make(map[string]struct{}, len(currentIDs))
	for _, id := range currentIDs {
		currentIDSet[id] = struct{}{}
	}
	for id, item := range s.items {
		if item.LibraryID == library.ID {
			if _, ok := currentIDSet[id]; !ok {
				delete(s.items, id)
			}
		}
	}
	task.State = "completed"
	task.ResultCount = len(currentIDs)
	task.FoundItems = len(currentIDs)
	_ = s.store.SaveTask(task.toRecord())
}

func releaseScanMemory() {
	runtime.GC()
	debug.FreeOSMemory()
}

func scanMediaInfoEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("TMMWEB_SCAN_MEDIAINFO"))) {
	case "", "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func scanSkipUnchangedDirsEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("TMMWEB_SCAN_SKIP_UNCHANGED_DIRS"))) {
	case "", "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}
