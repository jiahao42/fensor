package buf

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"v2ray.com/core/common/db"
	"v2ray.com/core/common/db/model"
	"v2ray.com/core/common/errors"
	"v2ray.com/core/common/signal"
	//"v2ray.com/core/common/buf"
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
func smartCopyInternal(reader Reader, writer Writer, pool *db.Pool, handler *copyHandler) (string, error) {
	//stage := 0
	ret := ""
	newDebugMsg("Buf: smartCopyInternal started")
	for {
		buffer, err := reader.ReadMultiBuffer()
		// 1. read until get URL
		// 2. call DB to see if the URL is blocked
		// 3. if yes, break and return error, then try another relay server
		// 4. if no, continue using this socks
		if !buffer.IsEmpty() {
			for _, handler := range handler.onData {
				handler(buffer)
				str := buffer.String()
				newDebugMsg("Buf: smartCopyInternal buffer " + str)
				ret += str
				if len(str) > 3 && str[:3] == "\x05\x01\x00" {
					addrType := str[3]
					if addrType == byte('\x01') { // IPv4
						arr := make([]string, 4, 4)
						//newDebugMsg("Buf: raw ip " + str[4:8])
						for i, b := range []byte(str[4:8]) {
							arr[i] = fmt.Sprintf("%d", b)
						}
						ipStr := strings.Join(arr, ".")
						newDebugMsg("Buf: parsed ip " + ipStr)
					} else if addrType == byte('\x03') { // domain
						addrLen := int(str[4])
						rawDomain := str[5 : 5+addrLen]
						//newDebugMsg("Buf: raw domain " + rawDomain)
						u := &url.URL{}
						u.UnmarshalBinary([]byte(rawDomain))
						port := binary.BigEndian.Uint16([]byte(str[5+addrLen : 5+addrLen+2]))
						domain := u.String()
						newDebugMsg("Buf: parsed domain " + domain + ":" + StructString(port))
						status, err := pool.LookupRecord(domain)
						if err != nil {
							// status not found
							// Do nothing, leave it to the freedom protocol
							newDebugMsg("Buf: domain not found " + domain)
						} else {
							newDebugMsg("Buf: domain found " + StructString(status))
							if status.Status == model.GOOD || status.Status == model.DNS_BLOCKED {
								// do nothing, leave it to the freedom protocol
							} else if status.Status == model.TCP_BLOCKED {
								//newDebugMsg("Buf: USE_RELAY for " + domain)
								return fmt.Sprintf("%s:%d", domain, port), errors.New("USE_RELAY")
							}
						}
					}
				}
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
	return ret, nil
}

// smartCopy
func relayCopyInternal(reader Reader, writer Writer, step int, addr string, handler *copyHandler) (string, error) {
	ret := ""
	newDebugMsg("Buf: relayCopyInternal started with " + addr)
	if step == 1 {
		// Set up SOCKS connection manually
		// 1. Request: Write to proxy server
		cmd_buf := &Buffer{}
		cmd_buf.WriteString("\x05\x01\x00") // \x0a
		cmd := MultiBuffer{}
		cmd = append(cmd, cmd_buf)
		if werr := writer.WriteMultiBuffer(cmd); werr != nil {
			newDebugMsg("Buf: relayCopyInternal writeError")
			return ret, writeError{werr}
		}
		newDebugMsg("Buf: relayCopyInternal sent handshake request")
	} else if step == 2 {
		// 2. Response: Got reply, should read \x05\x00
		//var reply MultiBuffer
		for {
			reply, err := reader.ReadMultiBuffer()
			if !reply.IsEmpty() {
				for _, handler := range handler.onData {
					handler(reply)
					ret += reply.String()
				}
				return ret, nil
			}
			if err != nil {
				return "", readError{err}
			}
		}
	} else if step == 3 {
		// 3. Request: Send target address
		arr := strings.Split(addr, ":")
		domain := arr[0]
		port, _ := strconv.Atoi(arr[1])
		port_str := string(byte(port))
		if port < (1 << 8) {
			port_str = string('\x00') + port_str
		}
		req_buf := &Buffer{}
		req_str := "\x05\x01\x00\x03" + string(byte(len(domain))) + domain + port_str
		newDebugMsg("Buf: relayCopyInternal send request " + req_str)
		req_buf.WriteString(req_str)
		req := MultiBuffer{}
		req = append(req, req_buf)
		if werr := writer.WriteMultiBuffer(req); werr != nil {
			newDebugMsg("Buf relayCopyInternal writeError when sending request")
			return ret, writeError{werr}
		}
	} else if step == 4 {
		//func copyInternal(reader Reader, writer Writer, handler *copyHandler) error {
		// 4. work as normal
		err := copyInternal(reader, writer, handler)
		return "USING CopyInternal", err
		//for {
		//buffer, err := reader.ReadMultiBuffer()
		//if !buffer.IsEmpty() {
		//for _, handler := range handler.onData {
		//handler(buffer)
		//str := buffer.String()
		//ret += str
		//}

		//if werr := writer.WriteMultiBuffer(buffer); werr != nil {
		//newDebugMsg("Buf: copyReturnInternal writeError" + err.Error())
		//return ret, writeError{werr}
		//}
		//}

		//if err != nil {
		//newDebugMsg("Buf: copyReturnInternal readError " + err.Error())
		//return ret, readError{err}
		//}
		//}
	}
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

func SmartCopy(reader Reader, writer Writer, pool *db.Pool, options ...CopyOption) (string, error) {
	var handler copyHandler
	for _, option := range options {
		option(&handler)
	}
	buffer, err := smartCopyInternal(reader, writer, pool, &handler)
	if err != nil && errors.Cause(err) != io.EOF {
		return buffer, err
	}
	return buffer, nil
}

func RelayCopy(reader Reader, writer Writer, step int, addr string, options ...CopyOption) (string, error) {
	newDebugMsg("Buf: RelayCopy step " + StructString(step))
	var handler copyHandler
	for _, option := range options {
		option(&handler)
	}
	buffer, err := relayCopyInternal(reader, writer, step, addr, &handler)
	if err != nil && errors.Cause(err) != io.EOF {
		return buffer, err
	}
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
