package retry

import "v2ray.com/core/common/errors"
import "os"
import "time"

type errPathObjHolder struct {}
func newError(values ...interface{}) *errors.Error { return errors.New(values...).WithPathObj(errPathObjHolder{}) }

func newDebugMsg(msg string) {
	f, err := os.OpenFile("/tmp/v2ray_debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
			panic(err)
	}
	t := time.Now()
	ts := t.Format("2006-01-02 15:04:05")
	defer f.Close()
	if _, err = f.WriteString(ts + ": " + msg + "\n"); err != nil {
		panic(err)
	}
}
