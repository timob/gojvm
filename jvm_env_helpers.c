#include "helpers.h"
#include "_cgo_export.h"

/* Exception handlers */
jboolean envExceptionCheck(JNIEnv *env) {
  return (*env)->ExceptionCheck(env);
}

jthrowable envExceptionOccurred(JNIEnv *env) {
	return (*env)->ExceptionOccurred(env);
}

void  envExceptionDescribe(JNIEnv* env){ (*env)->ExceptionDescribe(env); }
void  envExceptionClear(JNIEnv* env) { (*env)->ExceptionClear(env); }


/* 'Local' ref handlers */
jobject envNewLocalRef(JNIEnv *env, jobject obj) { return (*env)->NewLocalRef(env, obj); }

void envDeleteLocalRef(JNIEnv *env, jobject obj) { (*env)->DeleteLocalRef(env, obj); }   

jclass envFindClass(JNIEnv *env, const char *string){ return  (*env)->FindClass(env, string); }

jmethodID envGetMethodID(JNIEnv *env, jobject jobj, const char *meth, const char *sig){ return  (*env)->GetMethodID(env, jobj, meth, sig); }

jmethodID envGetStaticMethodID(JNIEnv *env, jclass jobj, const char *meth, const char *sig){ return  (*env)->GetStaticMethodID(env, jobj, meth, sig); }

jclass	envGetObjectClass(JNIEnv *env, jobject jobj){ return (*env)->GetObjectClass(env, jobj); }


// Call[Static]XXXMethodA
jint    envCallIntMethodA(JNIEnv *env, jobject o, jmethodID m, void *val){ return (*env)->CallIntMethodA(env,o,m,val); }
jint			envCallStaticIntMethodA(JNIEnv *env, jclass o, jmethodID m, void *val){ return (*env)->CallStaticIntMethodA(env,o,m,val); }

jshort    envCallShortMethodA(JNIEnv *env, jobject o, jmethodID m, void *val){ return (*env)->CallShortMethodA(env,o,m,val); }
jshort		envCallStaticShortMethodA(JNIEnv *env, jclass o, jmethodID m, void *val){ return (*env)->CallStaticShortMethodA(env,o,m,val); }

jlong			envCallLongMethodA(JNIEnv *env, jobject o, jmethodID m, void *val){ return (*env)->CallLongMethodA(env,o,m,val); }
jlong			envCallStaticLongMethodA(JNIEnv *env, jclass o, jmethodID m, void *val){ return (*env)->CallStaticLongMethodA(env,o,m,val); }

jobject		envCallObjectMethodA(JNIEnv *env, jobject o, jmethodID m, void *val){ return (*env)->CallObjectMethodA(env,o,m,val); }
jobject		envCallStaticObjectMethodA(JNIEnv *env, jclass o, jmethodID m, void *val){ return (*env)->CallStaticObjectMethodA(env,o,m,val); }

jfloat		envCallFloatMethodA(JNIEnv *env, jobject o, jmethodID m, void *val){ return (*env)->CallFloatMethodA(env,o,m,val); }
jfloat		envCallStaticFloatMethodA(JNIEnv *env, jclass o, jmethodID m, void *val){ return (*env)->CallStaticFloatMethodA(env,o,m,val); }

jdouble		envCallDoubleMethodA(JNIEnv *env, jobject o, jmethodID m, void *val){ return (*env)->CallDoubleMethodA(env,o,m,val); }
jdouble		envCallStaticDoubleMethodA(JNIEnv *env, jclass o, jmethodID m, void *val){ return (*env)->CallStaticDoubleMethodA(env,o,m,val); }


jboolean	envCallBoolMethodA(JNIEnv *env, jclass o, jmethodID m, void *val){ return (*env)->CallBooleanMethodA(env,o,m,val); }
jboolean	envCallStaticBoolMethodA(JNIEnv *env, jclass o, jmethodID m, void *val){ return (*env)->CallStaticBooleanMethodA(env,o,m,val); }

void	envCallStaticVoidMethodA(JNIEnv *env, jclass o, jmethodID m, void *val){ (*env)->CallStaticVoidMethodA(env,o,m,val); }
void	envCallVoidMethodA(JNIEnv *env, jobject o, jmethodID m, void *val){ (*env)->CallVoidMethodA(env,o,m,val); }

jint    envGetArrayLength(JNIEnv *env, jobject o){ return (*env)->GetArrayLength(env,o); }

jbyte     *envGetByteArrayElements(JNIEnv *env, jobject o, jboolean *b){
	return (*env)->GetByteArrayElements(env,o, b);
}

void      envReleaseByteArrayElements(JNIEnv *env, jobject o, jbyte *bts, jint mode){
	(*env)->ReleaseByteArrayElements(env,o, bts, mode);
}

jlong     *envGetLongArrayElements(JNIEnv *env, jobject o, jboolean *b){
	return (*env)->GetLongArrayElements(env, o, b);
}

void      envReleaseLongArrayElements(JNIEnv *env, jobject o, jlong *native, jint mode){
	(*env)->ReleaseLongArrayElements(env,o, native, mode);
}

jint     *envGetIntArrayElements(JNIEnv *env, jobject o, jboolean *b){
	return (*env)->GetIntArrayElements(env, o, b);
}

void      envReleaseIntArrayElements(JNIEnv *env, jobject o, jint *native, jint mode){
	(*env)->ReleaseIntArrayElements(env,o, native, mode);
}


jboolean  envIsSameObject(JNIEnv *env, jobject o, jobject o2){
	return (*env)->IsSameObject(env,o, o2);
}

jobject	envNewGlobalRef(JNIEnv *env, jobject o){
	return (*env)->NewGlobalRef(env,o);
}

jobject	envNewObjectA(JNIEnv *env, jobject o, jmethodID meth, void *jv){
	return (*env)->NewObjectA(env,o, meth, jv);
}

jobject	envNewObjectALP(JNIEnv *env, jobject o, jmethodID meth, ArgListPtr args){
	return (*env)->NewObjectA(env,o, meth, NULL);
}

jbyteArray  envNewByteArray(JNIEnv *env, jsize len){
	return (*env)->NewByteArray(env,len);
}

jobjectArray  envNewObjectArray(JNIEnv *env, jsize len, jclass klass, jobject init){
	return (*env)->NewObjectArray(env,len, klass, init);
}

void envSetObjectArrayElement(JNIEnv *env, jobjectArray array, jsize index, jobject val){
	(*env)->SetObjectArrayElement(env, array, index, val);
}

jobject envGetObjectArrayElement(JNIEnv *env, jobjectArray array, jsize index) {
	return (*env)->GetObjectArrayElement(env, array, index);
}

jstring   envNewStringUTF(JNIEnv *env, const char *s){
	return (*env)->NewStringUTF(env, s);
}

jsize     envGetStringUTFLength(JNIEnv *env, jstring s){
	return (*env)->GetStringUTFLength(env, s);
}

const char *envGetStringUTFChars(JNIEnv *env, jstring s, jboolean *jb){
	return (*env)->GetStringUTFChars(env, s, jb);
}

void      envReleaseStringUTFChars(JNIEnv *env, jstring s, const char *jb){
	(*env)->ReleaseStringUTFChars(env, s, jb);
}

void        envSetByteArrayRegion(JNIEnv *env, jbyteArray array, jsize start, jsize len, const void *buf){
	(*env)->SetByteArrayRegion(env, array, start, len, buf);
}

//fields
jfieldID envGetStaticFieldID(JNIEnv *env, jclass clazz, const char *name, const char *sig) {
	return (*env)->GetStaticFieldID(env, clazz, name, sig);
}

jobject envGetStaticObjectField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetStaticObjectField(env, clazz, fieldID);
}

jboolean envGetStaticBooleanField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetStaticBooleanField(env, clazz, fieldID);
}

jbyte envGetStaticByteField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetStaticByteField(env, clazz, fieldID);
}

jshort envGetStaticShortField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetStaticShortField(env, clazz, fieldID);
}

jint envGetStaticIntField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetStaticIntField(env, clazz, fieldID);
}

jlong envGetStaticLongField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetStaticLongField(env, clazz, fieldID);
}

jfloat envGetStaticFloatField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetStaticFloatField(env, clazz, fieldID);
}

jdouble envGetStaticDoubleField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetStaticDoubleField(env, clazz, fieldID);
}


//field

jfieldID envGetFieldID(JNIEnv *env, jclass clazz, const char *name, const char *sig) {
	return (*env)->GetFieldID(env, clazz, name, sig);
}

jobject envGetObjectField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetObjectField(env, clazz, fieldID);
}

jboolean envGetBooleanField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetBooleanField(env, clazz, fieldID);
}

jbyte envGetByteField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetByteField(env, clazz, fieldID);
}

jshort envGetShortField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetShortField(env, clazz, fieldID);
}

jint envGetIntField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetIntField(env, clazz, fieldID);
}

jlong envGetLongField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetLongField(env, clazz, fieldID);
}

jfloat envGetFloatField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetFloatField(env, clazz, fieldID);
}

jdouble envGetDoubleField(JNIEnv *env, jclass clazz, jfieldID fieldID) {
	return (*env)->GetDoubleField(env, clazz, fieldID);
}


/*

typedef struct {
char *name;
char *signature;
void *fnPtr;
} JNINativeMethod;

*/
// NB the actual call is registerNatives... TODO


jint	envGetJavaVM(JNIEnv  *env,JavaVM **jvm){
	return (*env)->GetJavaVM(env, jvm);
}


void *ret(GoInterface _if){ return _if.v; }

/*void  *genericCallback0(JNIEnv *env, jobject obj, void *argp){ return ret(goCallback((uintptr)env, (uintptr) obj, 0, (uintptr) argp)); }
void  *genericCallback1(JNIEnv *env, jobject obj, void *argp){ return ret(goCallback((uintptr)env, (uintptr) obj, 1, (uintptr) argp)); }
void  *genericCallback2(JNIEnv *env, jobject obj, void *argp){ return ret(goCallback((uintptr)env, (uintptr) obj, 2, (uintptr) argp)); }
void  *genericCallback3(JNIEnv *env, jobject obj, void *argp){ return ret(goCallback((uintptr)env, (uintptr) obj, 3, (uintptr) argp)); }
void  *genericCallback4(JNIEnv *env, jobject obj, void *argp){ return ret(goCallback((uintptr)env, (uintptr) obj, 4, (uintptr) argp)); }
void  *genericCallback5(JNIEnv *env, jobject obj, void *argp){ return ret(goCallback((uintptr)env, (uintptr) obj, 5, (uintptr) argp)); }
void  *genericCallback6(JNIEnv *env, jobject obj, void *argp){ return ret(goCallback((uintptr)env, (uintptr) obj, 6, (uintptr) argp)); }
void  *genericCallback7(JNIEnv *env, jobject obj, void *argp){ return ret(goCallback((uintptr)env, (uintptr) obj, 7, (uintptr) argp)); }
*/

/*

typedef struct {
  int length;
  jvalue *values;
} ArgList,  *ArgListPtr;
*/

ArgListPtr newArgList(int sz){
	ArgListPtr ptr = malloc(sizeof(ArgList));
	if (ptr == NULL) { return NULL; }
	ptr->length = sz;
	ptr->values = (jvalue *)malloc(sizeof(jvalue) * sz);
	if (ptr->values == NULL ){
		free(ptr);
		return NULL;
	}
	return ptr;
}

jvalue getArg(ArgListPtr l, int p){
	jvalue zeroval;
	if (p <= l->length){
		return l->values[p];
	} 
	// else panic
	return zeroval;
}

void delArgList(ArgListPtr alp){
	free(alp->values);
	free(alp);
}

void *doReturn(JNIEnv *env, struct goCallback_return ret){
	if (ret.r0) {
		return ret.r1.v;
	}
	printf("Failed, need to throw\n");
	return;
}

jint envRegisterNative(JNIEnv *env, jclass	klass, char *funcName, char *signature, void* fnPtr ){
	JNINativeMethod native;

	native.name = funcName;
	native.signature = signature;
    native.fnPtr = fnPtr;
	return (*env)->RegisterNatives(env, klass, &native, 1);
}

jint	envUnregisterNatives(JNIEnv  *env, jclass klass){
	return (*env)->UnregisterNatives(env, klass);
}


