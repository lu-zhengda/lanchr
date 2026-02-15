package logs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// Tailer reads log files specified in plist StandardOutPath/StandardErrorPath.
type Tailer struct{}

// NewTailer creates a new log file tailer.
func NewTailer() *Tailer {
	return &Tailer{}
}

// Tail returns the last N lines from the given file path.
func (t *Tailer) Tail(path string, lines int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", path, err)
	}
	defer f.Close()

	// Read all lines and return the last N.
	var allLines []string
	scanner := bufio.NewScanner(f)
	// Increase buffer size for long lines.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read log file %s: %w", path, err)
	}

	if lines <= 0 || lines >= len(allLines) {
		return allLines, nil
	}

	return allLines[len(allLines)-lines:], nil
}

// Follow streams new lines from the given file path to the returned channel.
// The caller should cancel the context to stop streaming.
func (t *Tailer) Follow(ctx context.Context, path string) (<-chan string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", path, err)
	}

	// Seek to the end of the file.
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to seek to end of %s: %w", path, err)
	}

	ch := make(chan string, 100)

	go func() {
		defer f.Close()
		defer close(ch)

		reader := bufio.NewReader(f)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := reader.ReadString('\n')
				if err != nil {
					// No new data; wait briefly and try again.
					time.Sleep(250 * time.Millisecond)
					continue
				}
				// Remove trailing newline.
				if len(line) > 0 && line[len(line)-1] == '\n' {
					line = line[:len(line)-1]
				}
				select {
				case ch <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}
