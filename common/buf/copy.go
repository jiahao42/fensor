package buf

import (
	"io"
	"time"
  //"net/url"
  //"net"
  "fmt"
  "strings"

	"v2ray.com/core/common/errors"
	"v2ray.com/core/common/signal"
)

type dataHandler func(MultiBuffer)

type copyHandler struct {
	onData []dataHandler
}

// SizeCounter is for counting bytes copied by Copy().
type SizeCounter struct {
	Size int64
}

// CopyOption is an option for copying data.
type CopyOption func(*copyHandler)

// UpdateActivity is a CopyOption to update activity on each data copy operation.
func UpdateActivity(timer signal.ActivityUpdater) CopyOption {
	return func(handler *copyHandler) {
		handler.onData = append(handler.onData, func(MultiBuffer) {
			timer.Update()
		})
	}
}

// CountSize is a CopyOption that sums the total size of data copied into the given SizeCounter.
func CountSize(sc *SizeCounter) CopyOption {
	return func(handler *copyHandler) {
		handler.onData = append(handler.onData, func(b MultiBuffer) {
			sc.Size += int64(b.Len())
		})
	}
}

type readError struct {
	error
}

func (e readError) Error() string {
	return e.error.Error()
}

func (e readError) Inner() error {
	return e.error
}

// IsReadError returns true if the error in Copy() comes from reading.
func IsReadError(err error) bool {
	_, ok := err.(readError)
	return ok
}

type writeError struct {
	error
}

func (e writeError) Error() string {
	return e.error.Error()
}

func (e writeError) Inner() error {
	return e.error
}

// IsWriteError returns true if the error in Copy() comes from writing.
func IsWriteError(err error) bool {
	_, ok := err.(writeError)
	return ok
}

func copyInternal(reader Reader, writer Writer, handler *copyHandler) error {
	for {
		buffer, err := reader.ReadMultiBuffer()
		if !buffer.IsEmpty() {
			for _, handler := range handler.onData {
				handler(buffer)
			}

			if werr := writer.WriteMultiBuffer(buffer); werr != nil {
				return writeError{werr}
			}
		}

		if err != nil {
			return readError{err}
		}
	}
}

// smartCopy
func smartCopyInternal(reader Reader, writer Writer, handler *copyHandler) (string, error) {
  //stage := 0
  ret := ""
  for {
		buffer, err := reader.ReadMultiBuffer()
    // TODO:
    // 1. read until get URL
    // 2. call DB to see if the URL is blocked
    // 3. if yes, break and return error, then try another relay server
    // 4. if no, continue using this socks
		if !buffer.IsEmpty() {
			for _, handler := range handler.onData {
				handler(buffer)
        str := buffer.String()
        ret += str
        newDebugMsg(str)
        if len(str) > 3 && str[:3] == "\x05\x01\x00" {
          newDebugMsg(str[3:])
          addrType := str[3]
          if addrType == byte('\x01') { // IPv4
            arr := make([]string, 4, 4)
            newDebugMsg("Buf: raw ip " + str[4:8])
            for i, b := range []byte(str[4:8]) {
              arr[i] = fmt.Sprintf("%d", b)
            }
            ipStr := strings.Join(arr, ".")
            newDebugMsg("Buf: parsed ip " + ipStr)
          } else if addrType == byte('\x03') { // domain
            //urlAddr := url.URL{}
            //urlAddr.UnmarshalBinary([]byte(str[4:8]))
          }
        }
        //if stage == 0 && str == "\x05\x01\x00" {
          //stage = 1
          //newDebugMsg(StructString(stage) + "STAGE 1!")
        //}
        //if stage == 1 && str == "\x05\x00" {
          //stage = 2
          //newDebugMsg("STAGE 2!")
        //} else if stage == 2 {
          //newDebugMsg("STAGE 3!")
        //}
        //if str == "\x05\x00" {
          //newDebugMsg(StructString(stage) + "SHIT!")
        //}
			}

			if werr := writer.WriteMultiBuffer(buffer); werr != nil {
        newDebugMsg("Buf: copyReturnInternal writeError")
				return ret, writeError{werr}
			}
		}

		if err != nil {
      newDebugMsg("Buf: copyReturnInternal readError")
			return ret, readError{err}
		}
  }
  //newDebugMsg("Buf: copyReturnInternal " + ret)
  return ret, nil
}

// Copy dumps all payload from reader to writer or stops when an error occurs. It returns nil when EOF.
func Copy(reader Reader, writer Writer, options ...CopyOption) error {
	var handler copyHandler
	for _, option := range options {
		option(&handler)
	}
	err := copyInternal(reader, writer, &handler)
	if err != nil && errors.Cause(err) != io.EOF {
		return err
	}
	return nil
}

func SmartCopy(reader Reader, writer Writer, options ...CopyOption) (string, error) {
	var handler copyHandler
	for _, option := range options {
		option(&handler)
	}
	buffer, err := smartCopyInternal(reader, writer, &handler)
	if err != nil && errors.Cause(err) != io.EOF {
		return buffer, err
	}
  //newDebugMsg("Buf: CopyReturn " + buffer)
	return buffer, nil
}

var ErrNotTimeoutReader = newError("not a TimeoutReader")

func CopyOnceTimeout(reader Reader, writer Writer, timeout time.Duration) error {
	timeoutReader, ok := reader.(TimeoutReader)
	if !ok {
		return ErrNotTimeoutReader
	}
	mb, err := timeoutReader.ReadMultiBufferTimeout(timeout)
	if err != nil {
		return err
	}
	return writer.WriteMultiBuffer(mb)
}
