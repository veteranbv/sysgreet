package collectors

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	wtmpPath      = "/var/log/wtmp"
	utmpUserProc  = 7 // USER_PROCESS
	wtmpTailBytes = 1 << 20
)

// utmpLayout describes glibc's struct utmp. The leading fields are the same
// everywhere; the session and timestamp after ut_exit are not.
type utmpLayout struct {
	size  int
	tvSec int  // offset of ut_tv.tv_sec
	sec64 bool // tv_sec is 64 bits wide
}

var (
	// Most targets, 32- and 64-bit alike: glibc keeps int32 session and
	// timestamp fields so 32-bit and 64-bit programs share the file.
	utmpCompat32 = utmpLayout{size: 384, tvSec: 340}
	// aarch64 never had a 32-bit glibc ABI to share with, so it uses a
	// long ut_session and a native 64-bit struct timeval.
	utmpNative64 = utmpLayout{size: 400, tvSec: 344, sec64: true}
)

func hostUtmpLayout() utmpLayout {
	if runtime.GOARCH == "arm64" {
		return utmpNative64
	}
	return utmpCompat32
}

func previousLogin(username, tty string) (*LastLoginInfo, error) {
	f, err := os.Open(wtmpPath)
	if err != nil {
		// Missing history (containers, distributions that moved to
		// wtmpdb) is normal; omit the line.
		return nil, nil
	}
	defer f.Close()
	return scanWtmp(f, hostUtmpLayout(), username, tty)
}

// scanWtmp walks the newest records backwards and returns the login by
// username that preceded the current session. The current session is the
// newest record on tty; a shell with no record of its own (a tmux pane, a
// new terminal tab) belongs to the user's newest login.
func scanWtmp(r io.ReadSeeker, layout utmpLayout, username, tty string) (*LastLoginInfo, error) {
	size, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	recSize := int64(layout.size)
	start := max(0, size-wtmpTailBytes)
	start -= start % recSize
	buf := make([]byte, size-start)
	if _, err := r.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	// A torn final write leaves a partial record; walking back from it
	// would misread every record before it.
	buf = buf[:len(buf)-len(buf)%layout.size]

	// Without a record on tty, the user's newest login is taken to be
	// the current session and the one before it is the answer.
	var fallback *LastLoginInfo
	seen := 0
	for off := len(buf) - layout.size; off >= 0; off -= layout.size {
		rec := buf[off : off+layout.size]
		if binary.NativeEndian.Uint16(rec[0:2]) != utmpUserProc || cString(rec[44:76]) != username {
			continue
		}
		seen++
		line := cString(rec[8:40])
		if tty != "" && line == tty {
			// The current session; the next older login is the answer.
			return olderLogin(buf[:off], layout, username), nil
		}
		if seen == 2 {
			fallback = loginFrom(rec, layout)
		}
	}
	return fallback, nil
}

// olderLogin returns the newest login by username in buf.
func olderLogin(buf []byte, layout utmpLayout, username string) *LastLoginInfo {
	for off := len(buf) - layout.size; off >= 0; off -= layout.size {
		rec := buf[off : off+layout.size]
		if binary.NativeEndian.Uint16(rec[0:2]) == utmpUserProc && cString(rec[44:76]) == username {
			return loginFrom(rec, layout)
		}
	}
	return nil
}

func loginFrom(rec []byte, layout utmpLayout) *LastLoginInfo {
	var sec int64
	if layout.sec64 {
		sec = int64(binary.NativeEndian.Uint64(rec[layout.tvSec:])) //nolint:gosec // wtmp timestamps are far below 2^63
	} else {
		// Reading the 32-bit field unsigned keeps timestamps valid
		// past 2038.
		sec = int64(binary.NativeEndian.Uint32(rec[layout.tvSec:]))
	}
	source := cString(rec[76:332])
	if source == "" {
		source = cString(rec[8:40])
	}
	return &LastLoginInfo{Timestamp: time.Unix(sec, 0), Source: source}
}

func cString(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return strings.TrimSpace(string(b))
}

// currentTTY names the session's terminal the way wtmp records it
// ("pts/3"): sshd's SSH_TTY when set, otherwise the terminal on stdin.
func currentTTY() string {
	target := os.Getenv("SSH_TTY")
	if target == "" {
		var err error
		if target, err = os.Readlink("/proc/self/fd/0"); err != nil {
			return ""
		}
	}
	tty, ok := strings.CutPrefix(target, "/dev/")
	if !ok {
		return ""
	}
	return tty
}
