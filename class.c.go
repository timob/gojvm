package gojvm

//#cgo LDFLAGS:-ljvm	-L/usr/lib/jvm/default-java/jre/lib/amd64/server
//#include "helpers.h"
import "C"
import (
	"github.com/timob/gojvm/types"
	//	"log"
)

/* represents a class (object) */
type Class struct {
	class C.jclass
}

func newClass(class C.jclass) *Class {
	return &Class{class}
}

/*
	returns the (potentially cached) types.Name of the class.
*/
func (self *Class) GetName(env *Environment) (name types.Name, err error) {
	//log.Printf("Name(miss)")
	var cstr string
	cstr, _, err = self.CallString(env, false, "getName")
	if err == nil {
		name = types.NewName(cstr)
	}
	return
}

// Calls the named void-method on the class
func (self *Class) CallVoid(env *Environment, static bool, mname string, params ...interface{}) (err error) {
	return env.CallClassVoid(self, static, mname, params...)
}

func (self *Class) CallBoolean(env *Environment, static bool, mname string, params ...interface{}) (i bool, err error) {
	return env.CallClassBool(self, static, mname, params...)
}

func (self *Class) CallShort(env *Environment, static bool, mname string, params ...interface{}) (i int16, err error) {
	return env.CallClassShort(self, static, mname, params...)
}

func (self *Class) CallInt(env *Environment, static bool, mname string, params ...interface{}) (i int, err error) {
	return env.CallClassInt(self, static, mname, params...)
}

func (self *Class) CallLong(env *Environment, static bool, mname string, params ...interface{}) (i int64, err error) {
	return env.CallClassLong(self, static, mname, params...)
}

func (self *Class) CallDouble(env *Environment, static bool, mname string, params ...interface{}) (i float64, err error) {
	return env.CallClassDouble(self, static, mname, params...)
}

func (self *Class) CallFloat(env *Environment, static bool, mname string, params ...interface{}) (i float32, err error) {
	return env.CallClassFloat(self, static, mname, params...)
}

func (self *Class) CallObj(env *Environment, static bool, mname string, rval types.Typed, params ...interface{}) (o *Object, err error) {
	return env.CallClassObj(self, static, mname, rval, params...)
}

func (self *Class) CallString(env *Environment, static bool, mname string, params ...interface{}) (str string, isnull bool, err error) {
	return env.CallClassString(self, static, mname, params...)
}

func (self *Class) CallLongArray(env *Environment, static bool, mname string, params ...interface{}) (i []int64, err error) {
	return env.CallClassLongArray(self, static, mname, params...)
}

func (self *Class) CallIntArray(env *Environment, static bool, mname string, params ...interface{}) (i []int, err error) {
	return env.CallClassIntArray(self, static, mname, params...)
}

//fields

func (self *Class) GetObjField(env *Environment, static bool, name string, rval types.Typed) (*Object, error) {
	return env.GetClassObjField(self, static, name, rval)
}

func (self *Class) GetBooleanField(env *Environment, static bool, name string) (bool, error) {
	return env.GetClassBooleanField(self, static, name)
}

func (self *Class) GetShortField(env *Environment, static bool, name string) (int16, error) {
	return env.GetClassShortField(self, static, name)
}

func (self *Class) GetIntField(env *Environment, static bool, name string) (int, error) {
	return env.GetClassIntField(self, static, name)
}

func (self *Class) GetLongField(env *Environment, static bool, name string) (int64, error) {
	return env.GetClassLongField(self, static, name)
}

func (self *Class) GetFloatField(env *Environment, static bool, name string) (float32, error) {
	return env.GetClassFloatField(self, static, name)
}

func (self *Class) GetDoubleField(env *Environment, static bool, name string) (float64, error) {
	return env.GetClassDoubleField(self, static, name)
}


