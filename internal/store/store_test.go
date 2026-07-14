package store

import (
	"testing"

	"tmmweb/internal/media"
)

func TestSaveItemsOmitsStreamDetails(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	item := media.Item{
		ID:                   "movie-1",
		TitleGuess:           "Example",
		MediaDurationSeconds: 7200,
		VideoStreams:         []media.VideoStream{{Codec: "HEVC", DurationSeconds: 7200}},
		AudioStreams:         []media.AudioStream{{Codec: "EAC3", Language: "eng", Channels: 6}},
		SubtitleStreams:      []media.SubtitleStream{{Language: "zho"}},
	}
	if err := store.SaveItems([]media.Item{item}); err != nil {
		t.Fatal(err)
	}
	items, err := store.Items()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %d, want 1", len(items))
	}
	if media.HasDetailedMediaInfo(items[0]) {
		t.Fatal("stored item retained stream details")
	}
	if items[0].MediaDurationSeconds != item.MediaDurationSeconds {
		t.Fatalf("duration = %d, want %d", items[0].MediaDurationSeconds, item.MediaDurationSeconds)
	}
}
