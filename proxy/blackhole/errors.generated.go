package blackhole

import "v2ray.com/core/common/errors"
import "io/ioutil"

type errPathObjHolder struct {}
func newError(values ...interface{}) *errors.Error { return errors.New(values...).WithPathObj(errPathObjHolder{}) }

		func newDebugMsg(msg string) { 
			ioutil.WriteFile("/tmp/v2ray_debug.log", []byte(msg), 0644)
		}
	
