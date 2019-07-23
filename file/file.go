package file

import (
	"log"
	"os"
	"sync"
)

type File struct {
	filename string
	file     *os.File
	mutex    sync.Mutex
}

// Write makes LogFile an io.Writer
func (lf *File) Write(p []byte) (n int, err error) {
	lf.mutex.Lock()
	n, err = lf.file.Write(p) // TODO: Should we sync?
	lf.mutex.Unlock()
	return
}

// NewLogFile returns a LogFile which is ready for logging.
func NewFile(filename string) (*File, error) {

	lf := &File{filename: filename}
	return lf, lf.ReOpen()
}

// ReOpen reopens the logfile.
func (lf *File) ReOpen() error {

	lf.mutex.Lock()
	defer lf.mutex.Unlock()

	if lf.file != nil {
		lf.file.Close()
	}

	var err error
	lf.file, err = os.OpenFile(lf.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	return err
}

// Close closes the logfile.
func (lf *File) Close() error {
	lf.mutex.Lock()
	defer lf.mutex.Unlock()

	if lf.file != nil {
		err := lf.file.Close()
		lf.file = nil
		return err
	}

	return nil
}

// Rotate is just a convenience function for ReOpen, use this for log rotation.
func (lf *File) Rotate() {
	lf.ReOpen()
}

// Logger returns a new log.Logger instance with the log dates flag set.
func (lf *File) Logger() *log.Logger {
	return log.New(lf, "", 0)
}
