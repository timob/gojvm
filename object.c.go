package gojvm

//#cgo CFLAGS:-I../include/
//#cgo LDFLAGS:-ljvm	-L/usr/lib/jvm/java-6-sun/jre/lib/amd64/server/
//#include "helpers.h"
import "C"
import (
	"github.com/timob/gojvm/types"
//	"log"
    "unsafe"
)

type Object struct {
	object C.jobject
}

func NewObjectStruct(o unsafe.Pointer) *Object {
    return &Object{C.jobject(o)}
}

// returns a new object value with specified parameters
// NB: refs are NOT adjusted directly by this call! Use it as a casting/construction-helper,
// not a Clone()
func newObject(obj C.jobject) *Object {
	return &Object{obj}
}

/* 
	Returns the Class() associated with the object
*/
func (self *Object) ObjectClass(env *Environment) (c *Class, err error) {
	return env.GetObjectClass(self)
}

var ClassClass = types.Name{"java", "lang", "Class"}
/*
	Returns the (potentially cached) name of the ObjectClass of the
	named object.
*/
func (self *Object) Name(env *Environment) (name types.Name, err error) {
	clsObj, err := self.CallObj(env, false, "getClass", types.Class{ClassClass})
	if err != nil {
		return
	}
	
	nameStr, _, err := clsObj.CallString(env, false, "getName")
	if err != nil {
		return
	}
	return types.NewName(nameStr), nil
/*		
	var c *Class
	c, err = clsObj.ObjectClass(env)
	if err == nil {
		defer env.DeleteLocalClassRef(c)
		name, err = c.GetName(env)
	} else {
		log.Printf("Couldn't get object class!")
	}
	return
*/	
}

func (self *Object) CallVoid(env *Environment, static bool, mname string, params ...interface{}) (err error) {
	return env.CallObjectVoid(self, static, mname, params...)
}

func (self *Object) CallInt(env *Environment, static bool, mname string, params ...interface{}) (i int, err error) {
	return env.CallObjectInt(self, static, mname, params...)
}

func (self *Object) CallLong(env *Environment, static bool, mname string, params ...interface{}) (i int64, err error) {
	return env.CallObjectLong(self, static, mname, params...)
}

func (self *Object) CallBool(env *Environment, static bool, mname string, params ...interface{}) (i bool, err error) {
	return env.CallObjectBool(self, static, mname, params...)
}

func (self *Object) CallFloat(env *Environment, static bool, mname string, params ...interface{}) (i float32, err error) {
	return env.CallObjectFloat(self, static, mname, params...)
}

func (self *Object) CallShort(env *Environment, static bool, mname string, params ...interface{}) (i int16, err error) {
	return env.CallObjectShort(self, static, mname, params...)
}

func (self *Object) CallDouble(env *Environment, static bool, mname string, params ...interface{}) (i float64, err error) {
	return env.CallObjectDouble(self, static, mname, params...)
}

func (self *Object) CallLongArray(env *Environment, static bool, mname string, params ...interface{}) (i []int64, err error) {
	return env.CallObjectLongArray(self, static, mname, params...)
}

func (self *Object) CallIntArray(env *Environment, static bool, mname string, params ...interface{}) (i []int, err error) {
	return env.CallObjectIntArray(self, static, mname, params...)
}

// Calls the named Object-method on the object instance
func (self *Object) CallObj(env *Environment, static bool, mname string, rval types.Typed, params ...interface{}) (vObj *Object, err error) {
	return env.CallObjectObj(self, static, mname, rval, params...)
}

/* 
	A wrapper around ObjCallObj specific to java/lang/String, that will return the result as a GoString 

	A null string returned with no Exception can be differentiated via the wasNull return value.
*/
func (self *Object) CallString(env *Environment, static bool, mname string, params ...interface{}) (str string, wasNull bool, err error) {
	return env.CallObjectString(self, static, mname, params...)
}

type CastObject struct {
	*Object
	types.Name
}

type ObjectArray struct {
	Objects []*Object
	types.Name
}

/*
func (self *AsJavaLangObject) Name(env *Environment) (name types.Name, err error) {
	return types.Name{"java", "lang", "Object"}, nil
}
*/


//fields

func (self *Object) GetObjField(env *Environment, static bool, name string, rval types.Typed) (*Object, error) {
    return env.GetObjectObjField(self, static, name, rval)
}

func (self *Object) GetBooleanField(env *Environment, static bool, name string) (bool, error) {
    return env.GetObjectBooleanField(self, static, name)
}

func (self *Object) GetShortField(env *Environment, static bool, name string) (int16, error) {
    return env.GetObjectShortField(self, static, name)
}

func (self *Object) GetIntField(env *Environment, static bool, name string) (int, error) {
    return env.GetObjectIntField(self, static, name)
}

func (self *Object) GetLongField(env *Environment, static bool, name string) (int64, error) {
    return env.GetObjectLongField(self, static, name)
}

func (self *Object) GetFloatField(env *Environment, static bool, name string) (float32, error) {
    return env.GetObjectFloatField(self, static, name)
}

func (self *Object) GetDoubleField(env *Environment, static bool, name string) (float64, error) {
    return env.GetObjectDoubleField(self, static, name)
}

func (self *Object) GetIntArrayField(env *Environment, static bool, name string) ([]int, error) {
    return env.GetObjecIntArrayField(self, static, name)
}


// ----


func (self *Object) SetObjField(env *Environment, static bool, name string, rval types.Typed, val *Object) (err error) {
    return env.setObjField(self, static, name, rval, val)
}

func (self *Object) SetBooleanField(env *Environment, static bool, name string, val bool) (err error) {
    return env.setBoolField(self, static, name, val)
}

func (self *Object) SetShortField(env *Environment, static bool, name string, val int16) (err error) {
    return env.setShortField(self, static, name, val)
}

func (self *Object) SetIntField(env *Environment, static bool, name string, val int) (err error) {
    return env.setIntField(self, static, name, val)
}

func (self *Object) SetLongField(env *Environment, static bool, name string, val int64) (err error) {
    return env.setLongField(self, static, name, val)
}

func (self *Object) SetFloatField(env *Environment, static bool, name string, val float32) (err error) {
    return env.setFloatField(self, static, name, val)
}

func (self *Object) SetDoubleField(env *Environment, static bool, name string, val float64) (err error) {
    return env.setDoubleField(self, static, name, val)
}
