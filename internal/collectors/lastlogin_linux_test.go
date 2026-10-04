package collectors

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

func wtmpRecord(kind uint16, line, user, host string, at time.Time) []byte {
	return wtmpRecordFor(utmpCompat32, kind, line, user, host, at)
}

func wtmpRecordFor(layout utmpLayout, kind uint16, line, user, host string, at time.Time) []byte {
	rec := make([]byte, layout.size)
	binary.NativeEndian.PutUint16(rec[0:2], kind)
	copy(rec[8:40], line)
	copy(rec[44:76], user)
	copy(rec[76:332], host)
	if layout.sec64 {
		binary.NativeEndian.PutUint64(rec[layout.tvSec:], uint64(at.Unix())) //nolint:gosec // test fixture, timestamps fit
	} else {
		binary.NativeEndian.PutUint32(rec[layout.tvSec:], uint32(at.Unix())) //nolint:gosec // test fixture, timestamps fit
	}
	return rec
}

func TestScanWtmp_SkipsCurrentSession(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	yesterday := now.Add(-26 * time.Hour)
	var log bytes.Buffer
	log.Write(wtmpRecord(utmpUserProc, "pts/1", "alice", "198.51.100.4", yesterday))
	log.Write(wtmpRecord(8, "pts/1", "", "", yesterday.Add(time.Hour))) // logout
	log.Write(wtmpRecord(utmpUserProc, "pts/2", "bob", "203.0.113.7", now.Add(-time.Hour)))
	log.Write(wtmpRecord(utmpUserProc, "pts/3", "alice", "203.0.113.9", now)) // this session

	got, err := scanWtmp(bytes.NewReader(log.Bytes()), utmpCompat32, "alice", "pts/3")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Source != "198.51.100.4" || !got.Timestamp.Equal(yesterday) {
		t.Fatalf("want yesterday's login from 198.51.100.4, got %+v", got)
	}
}

func TestScanWtmp_ConsoleLoginUsesLine(t *testing.T) {
	at := time.Now().Add(-time.Hour).Truncate(time.Second)
	var log bytes.Buffer
	log.Write(wtmpRecord(utmpUserProc, "tty1", "alice", "", at))
	log.Write(wtmpRecord(utmpUserProc, "pts/0", "alice", "203.0.113.9", time.Now()))
	got, _ := scanWtmp(bytes.NewReader(log.Bytes()), utmpCompat32, "alice", "pts/0")
	if got == nil || got.Source != "tty1" {
		t.Fatalf("a console login has no host; want tty1, got %+v", got)
	}
}

func TestScanWtmp_FirstLoginEver(t *testing.T) {
	rec := wtmpRecord(utmpUserProc, "pts/0", "alice", "203.0.113.9", time.Now())
	for _, tty := range []string{"pts/0", "pts/7", ""} {
		got, _ := scanWtmp(bytes.NewReader(rec), utmpCompat32, "alice", tty)
		if got != nil {
			t.Fatalf("tty %q: only the current session exists; want nil, got %+v", tty, got)
		}
	}
}

func TestScanWtmp_BothLayouts(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	yesterday := now.Add(-26 * time.Hour)
	for name, layout := range map[string]utmpLayout{"compat32": utmpCompat32, "native64": utmpNative64} {
		t.Run(name, func(t *testing.T) {
			var log bytes.Buffer
			log.Write(wtmpRecordFor(layout, utmpUserProc, "pts/1", "alice", "198.51.100.4", yesterday))
			log.Write(wtmpRecordFor(layout, utmpUserProc, "pts/3", "alice", "203.0.113.9", now))
			got, err := scanWtmp(bytes.NewReader(log.Bytes()), layout, "alice", "pts/3")
			if err != nil {
				t.Fatal(err)
			}
			if got == nil || !got.Timestamp.Equal(yesterday) || got.Source != "198.51.100.4" {
				t.Fatalf("want yesterday's login, got %+v", got)
			}
		})
	}
}

// A torn final write must not shift every record before it.
func TestScanWtmp_IgnoresPartialTrailingRecord(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	yesterday := now.Add(-26 * time.Hour)
	var log bytes.Buffer
	log.Write(wtmpRecord(utmpUserProc, "pts/1", "alice", "198.51.100.4", yesterday))
	log.Write(wtmpRecord(utmpUserProc, "pts/3", "alice", "203.0.113.9", now))
	log.Write(make([]byte, 100))
	got, _ := scanWtmp(bytes.NewReader(log.Bytes()), utmpCompat32, "alice", "pts/3")
	if got == nil || !got.Timestamp.Equal(yesterday) {
		t.Fatalf("want yesterday's login, got %+v", got)
	}
}

// tmux panes, new tabs and nested shells have no wtmp record of their own;
// the newest login is the current one and must not be reported.
func TestScanWtmp_ShellWithoutOwnRecord(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	yesterday := now.Add(-26 * time.Hour)
	var log bytes.Buffer
	log.Write(wtmpRecord(utmpUserProc, "pts/1", "alice", "198.51.100.4", yesterday))
	log.Write(wtmpRecord(utmpUserProc, "pts/0", "alice", "203.0.113.9", now.Add(-10*time.Second)))
	for _, tty := range []string{"pts/5", ""} {
		got, _ := scanWtmp(bytes.NewReader(log.Bytes()), utmpCompat32, "alice", tty)
		if got == nil || !got.Timestamp.Equal(yesterday) {
			t.Fatalf("tty %q: want yesterday's login, got %+v", tty, got)
		}
	}
}

// The current session is not always the newest record: a later login on
// another tty must not be reported as the previous one.
func TestScanWtmp_ReportsLoginBeforeCurrentSession(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	yesterday := now.Add(-26 * time.Hour)
	var log bytes.Buffer
	log.Write(wtmpRecord(utmpUserProc, "pts/1", "alice", "198.51.100.4", yesterday))
	log.Write(wtmpRecord(utmpUserProc, "pts/3", "alice", "203.0.113.9", now.Add(-time.Hour)))
	log.Write(wtmpRecord(utmpUserProc, "pts/4", "alice", "203.0.113.9", now))
	got, _ := scanWtmp(bytes.NewReader(log.Bytes()), utmpCompat32, "alice", "pts/3")
	if got == nil || !got.Timestamp.Equal(yesterday) {
		t.Fatalf("want yesterday's login, got %+v", got)
	}
}
