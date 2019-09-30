package watchcat

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

type Watcher struct {
	fp               *os.File
	lastFileInfo     os.FileInfo
	Position         int64
	WatchInterval    time.Duration
	NoChangedSeconds int
	FileSize         int64
	File             string
	Debug            bool
	Logger           *log.Logger
	Command          string
}

func NewWatcher() *Watcher {
	app := kingpin.New("watchcat", "watchcat")
	winterval := app.Flag("interval", "interval").Default("1s").Duration()
	noChangeSec := app.Flag("no-changed", "no changed seconds").Default("60").Int()
	filesize := app.Flag("filesize", "filesize").Default("10240").Int64()
	file := app.Flag("file", "file").Short('f').Required().String()
	command := app.Flag("command", "command").Short('c').String()
	debug := app.Flag("debug", "debug").Bool()

	app.Version("0.1.0")
	app.Parse(os.Args[1:])

	return &Watcher{
		WatchInterval:    *winterval,
		FileSize:         *filesize,
		NoChangedSeconds: *noChangeSec,
		File:             *file,
		Debug:            *debug,
		Command:          *command,
		Logger:           log.New(os.Stderr, "", 0),
	}
}

func getFileInfo(fp *os.File) (os.FileInfo, error) {
	fi, err := fp.Stat()
	if err != nil {
		return nil, err
	}

	return fi, nil
}

func (w *Watcher) initPosition(filename string) error {
	fp, err := os.Open(filename)
	if err != nil {
		return err
	}
	w.lastFileInfo, err = getFileInfo(fp)
	if err != nil {
		return err
	}

	pos, err := fp.Seek(w.lastFileInfo.Size(), io.SeekStart)
	if err != nil {
		return err
	}

	w.Position = pos

	w.fp = fp

	return nil
}

func (w *Watcher) Cat() error {
	err := w.initPosition(w.File)
	if err != nil {
		return err
	}

	var total int64
	var timeCounter int
	for range time.Tick(w.WatchInterval) {
		fi, err := getFileInfo(w.fp)
		if err != nil {
			return err
		}

		if int(time.Since(fi.ModTime()).Seconds()) > w.NoChangedSeconds && timeCounter >= w.NoChangedSeconds {
			sizeDiff := fi.Size() - w.lastFileInfo.Size()
			total += sizeDiff
			if sizeDiff >= w.FileSize {
				if w.Command == "" {
					w.Position, err = io.Copy(os.Stdout, w.fp)
					if err != nil {
						return err
					}

					if w.Debug {
						w.Logger.Println(fmt.Sprintf(`[%s] send STDOUT to %d bytes read total %d bytes`,
							time.Now().Format(time.RFC3339), sizeDiff, total),
						)
					}
				} else {
					cmd := exec.Command("sh", "-c", w.Command)
					var stderr bytes.Buffer
					cmd.Stderr = &stderr
					stdin, err := cmd.StdinPipe()
					if err != nil {
						return err
					}

					go func() {
						w.Position, err = io.Copy(stdin, w.fp)
						stdin.Close()
					}()

					err = cmd.Run()
					if err != nil {
						log.Println(stderr.String())
						return err
					}

					if w.Debug {
						w.Logger.Println(fmt.Sprintf(`[%s] execute %s send %d bytes read total %d bytes`,
							time.Now().Format(time.RFC3339), w.Command, sizeDiff, total),
						)
					}
				}

				w.lastFileInfo = fi
			} else {
				// update file position
				w.Position, err = w.fp.Seek(sizeDiff, io.SeekCurrent)
				if err != nil {
					return err
				}
				w.lastFileInfo = fi
				if w.Debug {
					w.Logger.Println(fmt.Sprintf(`[%s] discard %d bytes read total %d bytes`,
						time.Now().Format(time.RFC3339), sizeDiff, total),
					)
				}
			}
			timeCounter = 0
		}
		timeCounter = timeCounter + int(w.WatchInterval/time.Second)
	}

	return nil
}

func (w *Watcher) CloseFP() error {
	return w.fp.Close()
}
