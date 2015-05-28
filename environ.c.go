package gojvm

//#cgo CFLAGS:-I../include/
//#cgo LDFLAGS:-ljvm	-L/usr/lib/jvm/default-java/jre/lib/amd64/server
//#include "helpers.h"
import "C"
import (
	"errors"
	"github.com/timob/gojvm/types"
	"unsafe"
	"log"
)

const (
	JAVAClass = iota
	JAVAObject
)

/* 

	An environment consists of a pointer to a JNI environment
	and a caching map of class names to (globally referenced) class objects.

	TODO: Handle references on other items (nominally) correctly.


*/
type Environment struct {
	env             *C.JNIEnv
	jvm             *JVM
	classes         map[string]*Class
	quietExceptions bool
	// various 'consts'
	_UTF8 C.jstring // "UTF8" parameter
}

// Returns the underlying JNIEnv pointer.
// (In practice you should not need this <g>)
func (self *Environment) Ptr() unsafe.Pointer {
	return unsafe.Pointer(&self.env)
}

func (self *Environment) getObjectMethod(obj *Object, static bool, mname string, rType types.Typed, params ...interface{}) (meth *Method, args argList, objList []*Object, err error) {
	meth, err = self._objMethod(obj, mname, rType, params...)
	if err != nil {
		return
	}
	args, objList, err = newArgList(self, params...)
	return
}

func (self *Environment) getMethod(t interface{}, static bool, mname string, rType types.Typed, params ...interface{}) (jval C.jvalue, meth *Method, args argList, objList []*Object, err error) {
	if debug {
		log.Printf("getmethod: %s - %v", mname, t)
	}
	switch v := t.(type) {
	case *Object:
		//print("getObjMethod\t",mname, "\t",rType.TypeString(),"\n")
		jval = C.objValue(v.object)
		meth, args, objList, err = self.getObjectMethod(v, static, mname, rType, params...)
	case *Class:
		//			print("getClassMethod\t",mname, "\t",rType.TypeString(),"\n")
		jval = C.objValue(v.class)
		meth, args, objList, err = self.getClassMethod(v, static, mname, rType, params...)
	default:
		panic("getMethod called on unknown type")
	}
	return
}

// used in testing;  a 'squelch' helper
// such that:
//	func X(){
// 		defer env.defMute()() /*note the double parens!!!*/
// 		doSomeJavaCall
//	}
//
// would not output an exception to the console during processing
// regardless othe explicit 'mutedness'.
// there is a race condition here, but you're not supposed
// to be using *Environment in multiple threads anyhow :P
func (self *Environment) defMute() func() {
	muted := self.Muted()
	self.Mute(true)
	return func() {
		self.Mute(muted)
	}
}

func (self *Environment) getClassMethod(c *Class, static bool, mname string, rType types.Typed, params ...interface{}) (meth *Method, args argList, objList []*Object, err error) {
	if !static {
		meth, err = self._classMethod(c, mname, rType, params...)
	} else {
		meth, err = self._classStaticMethod(c, mname, rType, params...)
	}
	if err != nil {
		return
	}
	args, objList, err = newArgList(self, params...)
	return
}

// (Un)Suppress the java console barf of exceptions
// (execeptions are still caught, cleared and returned)
func (self *Environment) Mute(mute bool) { self.quietExceptions = mute }

// Returns the current state of the environmental exception mute.
func (self *Environment) Muted() bool { return self.quietExceptions }

// Refcounting is probably needed here, TODO: figure that out...
func (self *Environment) utf8() C.jstring {
	if self._UTF8 == nil {
		cs := C.CString("UTF8")
		defer C.free(unsafe.Pointer(cs))
		tf8 := C.envNewStringUTF(self.env, cs)
		self._UTF8 = C.jstring(C.envNewGlobalRef(self.env, tf8))
	}
	return self._UTF8
}

func NewEnvironment(jvm *JVM) *Environment {
	return &Environment{
		env:     new(C.JNIEnv),
		classes: map[string]*Class{},
		jvm:     jvm,
		quietExceptions: true,
	}
}

func (self Class) Kind() types.Kind { return types.ClassKind }

/* represents JNI method call;  without subject, style & parameters,
it is useless.  It (appears) to be an error to ref/unref methods.
*/
type Method struct {
	method C.jmethodID
}

func (self *Environment) findCachedClass(klass types.Name) (c *Class, err error) {
	if class, ok := self.classes[klass.AsPath()]; ok {
		c = class
	} else {
		err = errors.New("cache miss")
	}
	return
}

/* 
	returns a new *Object of class 'java/lang/String', containing the (UTF16 reinterpreted)
	representation of 's'.  Mostly a helper for passing strings into Java.
*/
func (self *Environment) NewStringObject(s string) (obj *Object, err error) {
	obj, err = self.NewInstanceStr("java/lang/String", []byte(s), self.utf8())
	return
}

func (self *Environment) setObjectArrayElement(arr *Object, pos int, item *Object) (err error) {
	C.envSetObjectArrayElement(self.env, arr.object, C.jsize(pos), item.object)
	return
}

func (self *Environment) newObjectArray(sz int, klass *Class, init C.jobject) (o *Object, err error) {
	ja := C.envNewObjectArray(self.env, C.jsize(sz), klass.class, init)
	if ja == nil {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		o = newObject(C.jobject(ja))
	}
	return
}

func (self *Environment) newByteObject(bts []byte) (o *Object, err error) {
	ja := C.envNewByteArray(self.env, C.jsize(len(bts)))
	if ja == nil {
		err = errors.New("Error allocating byte array")
	}
	if err == nil && len(bts) > 0 {
		bptr := make([]byte, len(bts))
		copy(bptr, bts)
		//log.Printf("bptr: %s %p %p", bptr,bptr, &bptr[0] )
		C.envSetByteArrayRegion(self.env, ja, 0, C.jsize(len(bptr)), unsafe.Pointer(&bptr[0]))
	}
	if err == nil {
		o = newObject(C.jobject(ja))
	}
	return
}

/* 
	returns a new *Object of the class named by 'klass' (Wrapper around NewInstance(types.NewName(...)))
*/
func (self *Environment) NewInstanceStr(klass string, params ...interface{}) (obj *Object, err error) {
	class, err := self.GetClass(types.NewName(klass))
	if err != nil {
		return
	}
	return self.NewInstance(class, params...)
}

/*
	returns a new *Object of type *Class, using the constructor identified by []params
*/
func (self *Environment) NewInstance(c *Class, params ...interface{}) (o *Object, err error) {
	meth, alp, localStack, err := self.getClassMethod(c, false, "<init>", types.Basic(types.VoidKind), params...)
	//	meth, alp, err := self.getObjectMethod(newObject(self, c, C.jobject( c.class)), "<init>", BasicType(JavaVoidKind), params...)
	if err != nil {
		return
	}
	defer blowStack(self, localStack)
	obj := C.envNewObjectA(self.env, c.class, meth.method, alp.Ptr())
	if obj != nil {
		obj = C.envNewGlobalRef(self.env, obj)
		o = newObject(obj)
	} else {
		err = self.ExceptionOccurred()
	}
	return
}

// returns a Class object;  the object will first be looked up in cache,
// and if not found there, resolved via Java and stored in the cache path.
// classes returned via /THIS/ channel, need not be unrefed, as they all
// hold a global ref.
//
// TODO: in truth, they should probably ALL be local-refs of the cached one...
func (self *Environment) GetClass(klass types.Name) (c *Class, err error) {
	c, err = self.findCachedClass(klass)
	if err == nil {
		return
	}
	s := C.CString(klass.AsPath())
	defer C.free(unsafe.Pointer(s))
	// print("envFindClass ", klass,"\n")
	kl := C.envFindClass(self.env, s)
	if kl == nil {
		//print("GetClass missed ", klass.AsPath(), "\n\n")
		err = self.ExceptionOccurred()
	} else {
		err = nil // clear the cache error
		//print("found ", klass,"\n")
		kl = C.jclass(C.envNewGlobalRef(self.env, kl))
		c = newClass(kl)
		self.classes[klass.AsPath()] = c
	}
	return
}

// Wrapper around GetClass(types.NewName(...))
func (self *Environment) GetClassStr(klass string) (c *Class, err error) {
	class := types.NewName(klass)
	return self.GetClass(class)
}

func (self *Environment) GetObjectClass(o *Object) (c *Class, err error) {
	kl := C.envGetObjectClass(self.env, o.object)
	if kl == nil {
		err = self.ExceptionOccurred()
	} else {
		c = newClass(kl)
	}
	return
}

func (self *Environment) _objMethod(obj *Object, name string, jt types.Typed, params ...interface{}) (meth *Method, err error) {
	class, err := self.GetObjectClass(obj)
	defer self.DeleteLocalClassRef(class)
	if err != nil {
		return
	}
	form, err := FormFor(self, jt, params...)
	if err != nil {
		return
	}

	cmethod := C.CString(name)
	defer C.free(unsafe.Pointer(cmethod))
	cform := C.CString(form)
	defer C.free(unsafe.Pointer(cform))

	if debug {
		log.Printf("_objMethod %V %V",  name, form)
	}
	m := C.envGetMethodID(self.env, class.class, cmethod, cform)
	if m == nil {
		err = self.ExceptionOccurred()
	} else {
		meth = &Method{m}
	}
	return

}

func (self *Environment) _classMethod(class *Class, name string, jt types.Typed, params ...interface{}) (meth *Method, err error) {
	form, err := FormFor(self, jt, params...)
	if err != nil {
		return
	}
	cmethod := C.CString(name)
	defer C.free(unsafe.Pointer(cmethod))
	cform := C.CString(form)
	defer C.free(unsafe.Pointer(cform))
	//cname, err := class.Name()
	//if err != nil { return }
	if debug {
		log.Printf("_classMethod %V %V", name, form)
	}
	m := C.envGetMethodID(self.env, class.class, cmethod, cform)
	if m == nil {
		err = self.ExceptionOccurred()
	} else {
		meth = &Method{m}
	}
	return
}

func (self *Environment) _classStaticMethod(class *Class, name string, jt types.Typed, params ...interface{}) (meth *Method, err error) {
	form, err := FormFor(self, jt, params...)
	if err != nil {
		return
	}
	cmethod := C.CString(name)
	defer C.free(unsafe.Pointer(cmethod))
	cform := C.CString(form)
	defer C.free(unsafe.Pointer(cform))
	//cname, err := class.Name()
	//if err != nil { return }
	//print("Looking for (static)", name, "\t", form, "\t in ", cname.AsPath(), "\n")
	m := C.envGetStaticMethodID(self.env, class.class, cmethod, cform)
	if m == nil {
		err = self.ExceptionOccurred()
	} else {
		meth = &Method{m}
	}
	return
}

type Exception struct {
	env *Environment
	ex C.jthrowable
}

func (self *Exception) Error() string {
	obj := newObject(C.jobject(self.ex))
	str, _, err := obj.CallString(self.env, false, "toString")
	if err != nil {
		panic(err)
	}
	return str
}

/*
	JNI documentation is unclear on the semantics of calling this
	when an exception has NOT occurred (e.g., is not indicated by
	a NULL value), but logic dictates that it _should_ be safe
	to call;  In that event, nil (should) be returned. 
*/
func (self *Environment) ExceptionOccurred() (ex *Exception) {
	throwable := C.envExceptionOccurred(self.env)
	if throwable != nil {
		// TODO: We'll need to do a global reference to this
		// if it outlasts a callback...
		ex = &Exception{self, throwable}
		if !self.quietExceptions {
			C.envExceptionDescribe(self.env)
		}
		C.envExceptionClear(self.env)
	}
	return
}

// Returns true if an ExceptionOccurred in this thread
// should produce a non-nil *Exception
func (self *Environment) ExceptionCheck() bool {
	return (C.envExceptionCheck(self.env) != C.JNI_FALSE)
}

// Syntactic sugar around &Class{C.jclass(LocalRef(&Object{C.jobject(class.class)}))}
func (self *Environment) NewLocalClassRef(c *Class) *Class {
	return newClass(C.jclass(C.envNewLocalRef(self.env, c.class)))
}

// Syntactic sugar around LocalUnref(&Object{C.jobject(class.class)})
func (self *Environment) DeleteLocalClassRef(c *Class) {
	C.envDeleteLocalRef(self.env, c.class)
}

// Adds a 'local' ref to the JVM for Object, and returns an object that is contains reference
func (self *Environment) NewLocalRef(o *Object) *Object {
	return newObject(C.envNewLocalRef(self.env, o.object))
}

// Release a local reference (returned from LocalRef) back to the JVM
func (self *Environment) DeleteLocalRef(o *Object) {
	C.envDeleteLocalRef(self.env, o.object)
}

// As gojvm is typically the /hosting/ context,
// a global reference in gojvm is more of a 'dont bother GC'ing this,
// I'm going to lose it somewhere in my stack',
// and as such should be use sparingly
func (self *Environment) NewGlobalRef(o *Object) *Object {
	return newObject(C.envNewGlobalRef(self.env, o.object))
}

func (self *Environment) RegisterNative(className string, method string, sig types.MethodSignature, fptr interface{}) error {
    class, err := self.GetClass(types.NewName(className))
    if err != nil {
        return err
    }
    cname := C.CString(method)
    defer C.free(unsafe.Pointer(cname))

    csig := C.CString(sig.String())
    defer C.free(unsafe.Pointer(csig))
    C.envRegisterNative(self.env, class.class, cname, csig, fptr.(unsafe.Pointer))

    return nil
}

/*
func (self *Environment) UnregisterNatives(c *Class) (err error) {
	if 0 != C.envUnregisterNatives(self.env, c.class) {
		err = self.ExceptionOccurred()
	}
	return
}

func (self *Environment) RegisterNative(c *Class, name string, fptr interface{}) (err error) {
	//	env.RegisterNative(klass, "NativePing", func(E *environment.Environment, O *environment.Object)(Error){
	//		nativePings += 1
	//	})

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	id, sig, err := self.jvm.addNative(self, fptr)
	csig := C.CString(sig.String())
	defer C.free(unsafe.Pointer(csig))
	if err != nil {
		return
	}

	if 0 != C.envRegisterNative(self.env, c.class, cname, csig, C.int(id)) {
		err = self.ExceptionOccurred()
	}
	return
}
*/


/* CallObject methods */
func asBool(jb C.jboolean) bool {
	if jb == C.JNI_FALSE {
		return false
	}
	return true
}

func (self *Environment) CallObjectVoid(obj *Object, static bool, name string, params ...interface{}) (err error) {
	return self.callVoid(obj, static, name, params...)
}

func (self *Environment) CallClassVoid(obj *Class, static bool, name string, params ...interface{}) (err error) {
	return self.callVoid(obj, static, name, params...)
}

func (self *Environment) CallObjectInt(obj *Object, static bool, name string, params ...interface{}) (v int, err error) {
	return self.callInt(obj, static, name, params...)
}

func (self *Environment) CallClassInt(obj *Class, static bool, name string, params ...interface{}) (v int, err error) {
	return self.callInt(obj, static, name, params...)
}

func (self *Environment) CallObjectLong(obj *Object, static bool, name string, params ...interface{}) (v int64, err error) {
	return self.callLong(obj, static, name, params...)
}

func (self *Environment) CallClassLong(obj *Class, static bool, name string, params ...interface{}) (v int64, err error) {
	return self.callLong(obj, static, name, params...)
}

func (self *Environment) CallObjectShort(obj *Object, static bool, name string, params ...interface{}) (v int16, err error) {
	return self.callShort(obj, static, name, params...)
}

func (self *Environment) CallClassShort(obj *Class, static bool, name string, params ...interface{}) (v int16, err error) {
	return self.callShort(obj, static, name, params...)
}

func (self *Environment) CallObjectBool(obj *Object, static bool, name string, params ...interface{}) (v bool, err error) {
	return self.callBool(obj, static, name, params...)
}

func (self *Environment) CallClassBool(obj *Class, static bool, name string, params ...interface{}) (v bool, err error) {
	return self.callBool(obj, static, name, params...)
}

func (self *Environment) CallObjectFloat(obj *Object, static bool, name string, params ...interface{}) (v float32, err error) {
	return self.callFloat(obj, static, name, params...)
}

func (self *Environment) CallClassFloat(obj *Class, static bool, name string, params ...interface{}) (v float32, err error) {
	return self.callFloat(obj, static, name, params...)
}

func (self *Environment) CallObjectDouble(obj *Object, static bool, name string, params ...interface{}) (v float64, err error) {
	return self.callDouble(obj, static, name, params...)
}

func (self *Environment) CallClassDouble(obj *Class, static bool, name string, params ...interface{}) (v float64, err error) {
	return self.callDouble(obj, static, name, params...)
}

func (self *Environment) CallObjectLongArray(obj *Object, static bool, name string, params ...interface{}) (v []int64, err error) {
	return self.callLongArray(obj, static, name, params...)
}

func (self *Environment) CallClassLongArray(obj *Class, static bool, name string, params ...interface{}) (v []int64, err error) {
	return self.callLongArray(obj, static, name, params...)
}

func (self *Environment) CallObjectIntArray(obj *Object, static bool, name string, params ...interface{}) (v []int, err error) {
	return self.callIntArray(obj, static, name, params...)
}

func (self *Environment) CallClassIntArray(obj *Class, static bool, name string, params ...interface{}) (v []int, err error) {
	return self.callIntArray(obj, static, name, params...)
}

func (self *Environment) CallObjectObj(obj *Object, static bool, name string, rtype types.Typed, params ...interface{}) (v *Object, err error) {
	return self.callObj(obj, static, name, rtype, params...)
}

func (self *Environment) CallClassObj(obj *Class, static bool, name string, rtype types.Typed, params ...interface{}) (v *Object, err error) {
	return self.callObj(obj, static, name, rtype, params...)
}

func (self *Environment) CallObjectString(obj *Object, static bool, name string, params ...interface{}) (s string, isNull bool, err error) {
	strObj, err := self.callObj(obj, static, name, types.Class{types.JavaLangString}, params...)
	if err == nil {
		defer self.DeleteLocalRef(strObj)
		return self.ToString(strObj)
	}
	return
}

func (self *Environment) CallClassString(obj *Class, static bool, name string, params ...interface{}) (s string, isNull bool, err error) {
	strObj, err := self.callObj(obj, static, name, types.Class{types.JavaLangString}, params...)
	if err == nil {
		defer self.DeleteLocalRef(strObj)
		return self.ToString(strObj)
	}
	return
}

func (self *Environment) callBool(z interface{}, static bool, name string, params ...interface{}) (b bool, err error) {
	jval, meth, args, localStack, err := self.getMethod(z, static, name, types.Basic(types.BoolKind), params...)
	if err != nil {
		return
	}
	defer blowStack(self, localStack)
	var ji C.jboolean
	if static {
		ji = C.envCallStaticBoolMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	} else {
		ji = C.envCallBoolMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		b = asBool(ji)
	}
	return
}

func (self *Environment) callVoid(z interface{}, static bool, name string, params ...interface{}) (err error) {
	jval, meth, args, localStack, err := self.getMethod(z, static, name, types.Basic(types.VoidKind), params...)
	if err != nil {
		return
	}
	defer blowStack(self, localStack)
	if static {
		C.envCallStaticVoidMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	} else {
		C.envCallVoidMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	return
}

func (self *Environment) callInt(z interface{}, static bool, name string, params ...interface{}) (v int, err error) {
	jval, meth, args, localStack, err := self.getMethod(z, static, name, types.Basic(types.IntKind), params...)
	if err != nil {
		return
	}
	defer blowStack(self, localStack)
	var ji C.jint
	if static {
		ji = C.envCallStaticIntMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	} else {
		ji = C.envCallIntMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	v = int(ji)
	return
}

func (self *Environment) callDouble(z interface{}, static bool, name string, params ...interface{}) (v float64, err error) {
	jval, meth, args, localStack, err := self.getMethod(z, static, name, types.Basic(types.DoubleKind), params...)
	if err != nil {
		return
	}
	defer blowStack(self, localStack)
	var ji C.jdouble
	if static {
		ji = C.envCallStaticDoubleMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	} else {
		ji = C.envCallDoubleMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	v = float64(ji)
	return
}

func (self *Environment) callFloat(z interface{}, static bool, name string, params ...interface{}) (v float32, err error) {
	jval, meth, args, localStack, err := self.getMethod(z, static, name, types.Basic(types.FloatKind), params...)
	if err != nil {
		return
	}
	defer blowStack(self, localStack)
	var ji C.jfloat
	if static {
		ji = C.envCallStaticFloatMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	} else {
		ji = C.envCallFloatMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	v = float32(ji)
	return
}

func (self *Environment) callObj(z interface{}, static bool, name string, rval types.Typed, params ...interface{}) (vObj *Object, err error) {
	jval, meth, args, localStack, err := self.getMethod(z, static, name, rval, params...)
	if err != nil {
		return
	}
	defer blowStack(self, localStack)
	var oval C.jobject
	if static {
		oval = C.envCallStaticObjectMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	} else {
		oval = C.envCallObjectMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		vObj = newObject(oval)
	}
	return
}

func (self *Environment) callLong(z interface{}, static bool, name string, params ...interface{}) (v int64, err error) {
	jval, meth, args, localStack, err := self.getMethod(z, static, name, types.Basic(types.LongKind), params...)
	if err != nil {
		return
	}
	defer blowStack(self, localStack)
	var oval C.jlong
	if static {
		oval = C.envCallStaticLongMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	} else {
		oval = C.envCallLongMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		v = int64(oval)
	}
	return
}

func (self *Environment) callShort(z interface{}, static bool, name string, params ...interface{}) (v int16, err error) {
	jval, meth, args, localStack, err := self.getMethod(z, static, name, types.Basic(types.ShortKind), params...)
	if err != nil {
		return
	}
	defer blowStack(self, localStack)
	var oval C.jshort
	if static {
		oval = C.envCallStaticShortMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	} else {
		oval = C.envCallShortMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		v = int16(oval)
	}
	return
}

func (self *Environment) callLongArray(z interface{}, static bool, name string, params ...interface{}) (v []int64, err error) {
	jval, meth, args, localStack, err := self.getMethod(z, static, name, types.Array{types.Basic(types.LongKind)}, params...)
	if err != nil {
		return
	}
	defer blowStack(self, localStack)
	var oval C.jobject
	if static {
		oval = C.envCallStaticObjectMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	} else {
		oval = C.envCallObjectMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	}
	if oval == nil {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		v = self.ToInt64Array(newObject(oval))
	}
	return
}

func (self *Environment) callIntArray(z interface{}, static bool, name string, params ...interface{}) (v []int, err error) {
	jval, meth, args, localStack, err := self.getMethod(z, static, name, types.Array{types.Basic(types.IntKind)}, params...)
	if err != nil {
		return
	}
	defer blowStack(self, localStack)
	var oval C.jobject
	if static {
		oval = C.envCallStaticObjectMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	} else {
		oval = C.envCallObjectMethodA(self.env, C.valObject(jval), meth.method, args.Ptr())
	}
	if oval == nil {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		v = self.ToIntArray(newObject(oval))
	}
	return
}

func (self *Environment) ToString(strobj *Object) (str string, isNull bool, err error) {
	if strobj.object == nil {
		isNull = true
		return
	}
	var bytesObj *Object
	bytesObj, err = self.CallObjectObj(strobj, false, "getBytes", types.Array{types.Basic(types.ByteKind)}, self.utf8())
	if err == nil && bytesObj == nil {
		isNull = true
		return
	}
	if err == nil {
		defer self.DeleteLocalRef(bytesObj)
		alen := C.envGetArrayLength(self.env, bytesObj.object)
		_false := C.jboolean(C.JNI_FALSE)
		ptr := C.envGetByteArrayElements(self.env, bytesObj.object, &_false)
		defer C.envReleaseByteArrayElements(self.env, bytesObj.object, ptr, 0)
		str = string(C.GoBytes(unsafe.Pointer(ptr), C.int(alen)))
	}
	return
}

func (self *Environment) ToInt64Array(arrayObj *Object) (array []int64) {
	alen := C.envGetArrayLength(self.env, arrayObj.object)
	_false := C.jboolean(C.JNI_FALSE)
	ptr := C.envGetLongArrayElements(self.env, arrayObj.object, &_false)
	defer C.envReleaseLongArrayElements(self.env, arrayObj.object, ptr, 0)
	bytes := C.GoBytes(unsafe.Pointer(ptr), C.int(alen) * 8)
	array = (*(*[1024*1024]int64)(unsafe.Pointer(&bytes)))[0:int(alen)]
	return
}

func (self *Environment) ToIntArray(arrayObj *Object) (array []int) {
	alen := C.envGetArrayLength(self.env, arrayObj.object)
	_false := C.jboolean(C.JNI_FALSE)
	ptr := C.envGetIntArrayElements(self.env, arrayObj.object, &_false)
	defer C.envReleaseIntArrayElements(self.env, arrayObj.object, ptr, 0)
	bytes := C.GoBytes(unsafe.Pointer(ptr), C.int(alen) * 4)
	array = (*(*[1024*1024]int)(unsafe.Pointer(&bytes)))[0:int(alen)]
	return
}

func (self *Environment) ToObjectArray(arrayObj *Object) []*Object {
    var glen int
    if arrayObj.object == nil {
        glen = 0
    } else {
        jlen := C.envGetArrayLength(self.env, arrayObj.object)
        glen = int(jlen)
    }
	objs := make([]*Object, glen)
	for i := 0; i < glen; i++ {
		objs[i] = &Object{C.envGetObjectArrayElement(self.env, arrayObj.object, C.jsize(i))}
	}	
	return objs
}

//fields
func (self *Environment) GetClassObjField(obj *Class, static bool, name string, rval types.Typed) (*Object, error) {
	return self.getObjField(obj, static, name, rval)
}

func (self *Environment) GetClassBooleanField(obj *Class, static bool, name string) (v bool, err error) {
	return self.getBoolField(obj, static, name)
}

func (self *Environment) GetClassShortField(obj *Class, static bool, name string) (v int16, err error) {
	return self.getShortField(obj, static, name)
}

func (self *Environment) GetClassIntField(obj *Class, static bool, name string) (v int, err error) {
	return self.getIntField(obj, static, name)
}

func (self *Environment) GetClassLongField(obj *Class, static bool, name string) (v int64, err error) {
	return self.getLongField(obj, static, name)
}

func (self *Environment) GetClassFloatField(obj *Class, static bool, name string) (v float32, err error) {
	return self.getFloatField(obj, static, name)
}

func (self *Environment) GetClassDoubleField(obj *Class, static bool, name string) (v float64, err error) {
	return self.getDobuleField(obj, static, name)
}

//object fields
func (self *Environment) GetObjectObjField(obj *Object, static bool, name string, rval types.Typed) (*Object, error) {
    return self.getObjField(obj, static, name, rval)
}

func (self *Environment) GetObjectBooleanField(obj *Object, static bool, name string) (v bool, err error) {
    return self.getBoolField(obj, static, name)
}

func (self *Environment) GetObjectShortField(obj *Object, static bool, name string) (v int16, err error) {
    return self.getShortField(obj, static, name)
}

func (self *Environment) GetObjectIntField(obj *Object, static bool, name string) (v int, err error) {
    return self.getIntField(obj, static, name)
}

func (self *Environment) GetObjectLongField(obj *Object, static bool, name string) (v int64, err error) {
    return self.getLongField(obj, static, name)
}

func (self *Environment) GetObjectFloatField(obj *Object, static bool, name string) (v float32, err error) {
    return self.getFloatField(obj, static, name)
}

func (self *Environment) GetObjectDoubleField(obj *Object, static bool, name string) (v float64, err error) {
    return self.getDobuleField(obj, static, name)
}

func (self *Environment) GetObjecIntArrayField(obj *Object, static bool, name string) (v []int, err error) {
    return self.getIntArrayField(obj, static, name)
}

type Field struct {
	field C.jfieldID
}

func (self *Environment) getClassField(c *Class, static bool, mname string, rType types.Typed) (meth *Field, err error) {
	cmethod := C.CString(mname)
	defer C.free(unsafe.Pointer(cmethod))
	cform := C.CString(rType.TypeString())
	defer C.free(unsafe.Pointer(cform))

	var m C.jfieldID 
	if !static {
		//todo
	} else {
		m = C.envGetStaticFieldID(self.env, c.class, cmethod, cform)
	}
	if m == nil {
		err = self.ExceptionOccurred()
		return nil, err
	} 

	return &Field{m}, nil
}

func (self *Environment) getObjectField(o *Object, static bool, mname string, rType types.Typed) (meth *Field, err error) {
    class, err := self.GetObjectClass(o)
    defer self.DeleteLocalClassRef(class)
    if err != nil {
        return
    }

    cmethod := C.CString(mname)
    defer C.free(unsafe.Pointer(cmethod))
    cform := C.CString(rType.TypeString())
    defer C.free(unsafe.Pointer(cform))

    var m C.jfieldID
    if static {
        //todo
    } else {
        m = C.envGetFieldID(self.env, class.class, cmethod, cform)
    }
    if m == nil {
        err = self.ExceptionOccurred()
        return nil, err
    }

    return &Field{m}, nil
}


func (self *Environment) getField(t interface{}, static bool, mname string, rType types.Typed) (jval C.jvalue, field *Field, err error) {
	if debug {
		log.Printf("getfield: %s - %v", mname, t)
	}
	switch v := t.(type) {
    case *Object:
        jval = C.objValue(v.object)
        field, err = self.getObjectField(v, static, mname, rType)
/*
	case *Object:
		//print("getObjMethod\t",mname, "\t",rType.TypeString(),"\n")
		jval = C.objValue(v.object)
		meth, args, objList, err = self.getObjectMethod(v, static, mname, rType, params...)
*/
	case *Class:
		jval = C.objValue(v.class)
		field, err = self.getClassField(v, static, mname, rType)
	default:
		panic("getField called on unknown type")
	}
	
	return
}

func (self *Environment) getObjField(z interface{}, static bool, name string, rval types.Typed) (v *Object, err error) {
	jval, field, err := self.getField(z, static, name, rval)
	if err != nil {
		return
	}
	var oval C.jobject
	if static {
		oval = C.envGetStaticObjectField(self.env, C.valObject(jval), field.field);
	} else {
        oval = C.envGetObjectField(self.env, C.valObject(jval), field.field);
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		v = &Object{oval}
	}
	return
}

func (self *Environment) getBoolField(z interface{}, static bool, name string) (v bool, err error) {
	jval, field, err := self.getField(z, static, name, types.Basic(types.BoolKind))
	if err != nil {
		return
	}
	var oval C.jboolean
	if static {
		oval = C.envGetStaticBooleanField(self.env, C.valObject(jval), field.field);
	} else {
        oval = C.envGetBooleanField(self.env, C.valObject(jval), field.field);
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		v = asBool(oval)
	}
	return
}

func (self *Environment) getShortField(z interface{}, static bool, name string) (v int16, err error) {
	jval, field, err := self.getField(z, static, name, types.Basic(types.ShortKind))
	if err != nil {
		return
	}
	var oval C.jshort
	if static {
		oval = C.envGetStaticShortField(self.env, C.valObject(jval), field.field);
	} else {
        oval = C.envGetShortField(self.env, C.valObject(jval), field.field);
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		v = int16(oval)
	}
	return
}

func (self *Environment) getIntField(z interface{}, static bool, name string) (v int, err error) {
	jval, field, err := self.getField(z, static, name, types.Basic(types.IntKind))
	if err != nil {
		return
	}
	var oval C.jint
	if static {
		oval = C.envGetStaticIntField(self.env, C.valObject(jval), field.field);
	} else {
        oval = C.envGetIntField(self.env, C.valObject(jval), field.field);
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		v = int(oval)
	}
	return
}

func (self *Environment) getLongField(z interface{}, static bool, name string) (v int64, err error) {
	jval, field, err := self.getField(z, static, name, types.Basic(types.IntKind))
	if err != nil {
		return
	}
	var oval C.jlong
	if static {
		oval = C.envGetStaticLongField(self.env, C.valObject(jval), field.field);
	} else {
        oval = C.envGetLongField(self.env, C.valObject(jval), field.field);
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		v = int64(oval)
	}
	return
}

func (self *Environment) getFloatField(z interface{}, static bool, name string) (v float32, err error) {
	jval, field, err := self.getField(z, static, name, types.Basic(types.IntKind))
	if err != nil {
		return
	}
	var oval C.jfloat
	if static {
		oval = C.envGetStaticFloatField(self.env, C.valObject(jval), field.field);
	} else {
        oval = C.envGetFloatField(self.env, C.valObject(jval), field.field);
	}
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		v = float32(oval)
	}
	return
}

func (self *Environment) getDobuleField(z interface{}, static bool, name string) (v float64, err error) {
	jval, field, err := self.getField(z, static, name, types.Basic(types.IntKind))
	if err != nil {
		return
	}
	var oval C.jdouble
	if static {
		oval = C.envGetStaticDoubleField(self.env, C.valObject(jval), field.field);
	} else {
		oval = C.envGetDoubleField(self.env, C.valObject(jval), field.field);
    }
	if self.ExceptionCheck() {
		err = self.ExceptionOccurred()
	}
	if err == nil {
		v = float64(oval)
	}
	return
}

func (self *Environment) getIntArrayField(z interface{}, static bool, name string) (v []int, err error) {
    jval, field, err := self.getField(z, static, name, types.Array{types.Basic(types.IntKind)})
    if err != nil {
        return
    }
    var oval C.jobject
    if static {
        oval = C.envGetStaticObjectField(self.env, C.valObject(jval), field.field)
    } else {
        oval = C.envGetObjectField(self.env, C.valObject(jval), field.field)
    }
    if oval == nil {
        err = self.ExceptionOccurred()
    }
    if err == nil {
        v = self.ToIntArray(newObject(oval))
    }

    return
}

/* ==== */

func (self *Environment) setObjField(z interface{}, static bool, name string, rval types.Typed, val *Object) (err error) {
    jval, field, err := self.getField(z, static, name, rval)
    if err != nil {
        return
    }
    if static {
        C.envSetStaticObjectField(self.env, C.valObject(jval), field.field, val.object);
    } else {
        C.envSetObjectField(self.env, C.valObject(jval), field.field, val.object);
    }
    if self.ExceptionCheck() {
        err = self.ExceptionOccurred()
    }
    return
}

func (self *Environment) setBoolField(z interface{}, static bool, name string, gval bool) (err error) {
    jval, field, err := self.getField(z, static, name, types.Basic(types.BoolKind))
    if err != nil {
        return
    }
    val := C.jboolean(C.JNI_FALSE)
    if gval {
        val = C.JNI_TRUE
    }
    if static {
        C.envSetStaticBooleanField(self.env, C.valObject(jval), field.field, val);
    } else {
        C.envSetBooleanField(self.env, C.valObject(jval), field.field, val);
    }
    if self.ExceptionCheck() {
        err = self.ExceptionOccurred()
    }
    return
}

func (self *Environment) setShortField(z interface{}, static bool, name string, val int16) (err error) {
    jval, field, err := self.getField(z, static, name, types.Basic(types.ShortKind))
    if err != nil {
        return
    }
    if static {
        C.envSetStaticShortField(self.env, C.valObject(jval), field.field, C.jshort(val));
    } else {
        C.envSetShortField(self.env, C.valObject(jval), field.field, C.jshort(val));
    }
    if self.ExceptionCheck() {
        err = self.ExceptionOccurred()
    }
    return
}

func (self *Environment) setIntField(z interface{}, static bool, name string, val int) (err error) {
    jval, field, err := self.getField(z, static, name, types.Basic(types.IntKind))
    if err != nil {
        return
    }
    if static {
        C.envSetStaticIntField(self.env, C.valObject(jval), field.field, C.jint(val));
    } else {
        C.envSetIntField(self.env, C.valObject(jval), field.field, C.jint(val));
    }
    if self.ExceptionCheck() {
        err = self.ExceptionOccurred()
    }
    return
}

func (self *Environment) setLongField(z interface{}, static bool, name string, val int64) (err error) {
    jval, field, err := self.getField(z, static, name, types.Basic(types.IntKind))
    if err != nil {
        return
    }
    if static {
        C.envSetStaticLongField(self.env, C.valObject(jval), field.field, C.jlong(val));
    } else {
        C.envSetLongField(self.env, C.valObject(jval), field.field, C.jlong(val));
    }
    if self.ExceptionCheck() {
        err = self.ExceptionOccurred()
    }
    return
}

func (self *Environment) setFloatField(z interface{}, static bool, name string, val float32) (err error) {
    jval, field, err := self.getField(z, static, name, types.Basic(types.IntKind))
    if err != nil {
        return
    }
    if static {
        C.envSetStaticFloatField(self.env, C.valObject(jval), field.field, C.jfloat(val));
    } else {
        C.envSetFloatField(self.env, C.valObject(jval), field.field, C.jfloat(val));
    }
    if self.ExceptionCheck() {
        err = self.ExceptionOccurred()
    }
    return
}

func (self *Environment) setDoubleField(z interface{}, static bool, name string, val float64) (err error) {
    jval, field, err := self.getField(z, static, name, types.Basic(types.IntKind))
    if err != nil {
        return
    }
    if static {
        C.envSetStaticDoubleField(self.env, C.valObject(jval), field.field, C.jdouble(val));
    } else {
        C.envSetDoubleField(self.env, C.valObject(jval), field.field, C.jdouble(val));
    }
    if self.ExceptionCheck() {
        err = self.ExceptionOccurred()
    }
    return
}
