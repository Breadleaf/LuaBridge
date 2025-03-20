package lua

/*
#cgo LDFLAGS: -lm -ldl

#ifdef __linux__
#define LUA_USE_LINUX
#endif

#include <stdlib.h>
#include <stdint.h>
#include "lua.h"
#include "lualib.h"
#include "lauxlib.h"

// Macro wrappers:

// The macro
//   #define lua_tonumber(L,i)	lua_tonumberx(L,(i),NULL)
// is wrapped as:
static lua_Number LuaMacro_tonumber(lua_State *L, int index) {
    return lua_tonumber(L, index);
}

// The macro
//   #define lua_tostring(L,i)	lua_tolstring(L, (i), NULL)
// is wrapped as:
static const char* LuaMacro_tostring(lua_State *L, int index) {
    return lua_tostring(L, index);
}

// The macro
//   #define lua_pop(L,n)		lua_settop(L, -(n)-1)
// is wrapped as:
static void LuaMacro_pop(lua_State *L, int n) {
    lua_pop(L, n);
}

// The macro
//   #define lua_insert(L,idx)	lua_rotate(L, (idx), 1)
// is wrapped as:
static void LuaMacro_insert(lua_State *L, int idx) {
    lua_insert(L, idx);
}

// The macro
//   #define lua_remove(L,idx)	(lua_rotate(L, (idx), -1), lua_pop(L, 1))
// is wrapped as:
static void LuaMacro_remove(lua_State *L, int idx) {
    lua_remove(L, idx);
}

// The macro
//   #define lua_pushliteral(L, s)	lua_pushstring(L, "" s)
// is wrapped as:
static void LuaMacro_pushliteral(lua_State *L, const char* s) {
    lua_pushstring(L, s);
}

// The macro
//   #define lua_pcall(L,n,r,f)	lua_pcallk(L, (n), (r), (f), 0, NULL)
// is wrapped as:
static int LuaMacro_pcall(lua_State *L, int n, int r, int f) {
    return lua_pcallk(L, n, r, f, 0, NULL);
}

// The macro
//   #define lua_isnil(L,n)		(lua_type(L, (n)) == LUA_TNIL)
// is wrapped as:
static int LuaMacro_isnil(lua_State *L, int n) {
    return lua_type(L, n) == LUA_TNIL;
}

// Extern functions:

extern int goLuaCallback(lua_State *L);
extern int traceback(lua_State *L);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type LuaFunction func(L *LuaState) int

type LuaState struct {
    L *C.lua_State
    closures map[int64]LuaFunction
    closureID int64
}

var luaStates = make(map[*C.lua_State]*LuaState)

func NewState() *LuaState {
    L := C.luaL_newstate()
    if L == nil {
        panic("failed to create Lua state")
    }

    // open all standard libraries
    C.luaL_openlibs(L)

    s := &LuaState{
        L: L,
        closures: make(map[int64]LuaFunction),
        closureID: 0,
    }
    luaStates[L] = s

    return s
}

//export goLuaCallback
func  goLuaCallback(L *C.lua_State) C.int {
    state, ok := luaStates[L]
    if !ok {
        panic("failed to find Lua state for callback")
    }

    // get the closure ID from the upvalue
    id := int64(C.LuaMacro_tonumber(L, 1))
    if goFunction, found := state.closures[id]; found {
        return C.int(goFunction(state))
    }

    return 0
}

//export traceback
func traceback(L *C.lua_State) C.int {
    errorMsg := C.LuaMacro_tostring(L, 1)
    if errorMsg == nil {
        C.LuaMacro_pushliteral(L, C.CString("(no error message)"))
        return 1
    }
    
    // Get the current Lua state from the registry
    C.lua_pushthread(L)
    C.lua_rawget(L, C.LUA_REGISTRYINDEX)
    
    // Generate traceback
    C.luaL_traceback(L, L, errorMsg, 1)
    return 1
}

func (s *LuaState) RegisterGlobalFunction(name string, fn LuaFunction) {
    // save the function in the registry
    id := s.closureID
    s.closures[id] = fn
    s.closureID++

    // push the closure ID as an upvalue
    C.lua_pushnumber(s.L, C.lua_Number(id))
    C.lua_pushcclosure(s.L, (*[0]byte)(C.goLuaCallback), 1)

    // set it as a global variable
    cname := C.CString(name)
    defer C.free(unsafe.Pointer(cname))
    C.lua_setglobal(s.L, cname)
}

// Can be used to remove libraries or other global variables
func (s *LuaState) RemoveGlobal(name string) {
    // set the global variable to nil
    cname := C.CString(name)
    defer C.free(unsafe.Pointer(cname))
    C.lua_pushnil(s.L)
    C.lua_setglobal(s.L, cname)
}

func (s *LuaState) RunStringError(script string) error {
    cscript := C.CString(script)
    defer C.free(unsafe.Pointer(cscript))
    if C.luaL_loadstring(s.L, cscript) != 0 {
        // err := C.GoString(C.lua_tolstring(s.L, -1, nil))
        err := C.GoString(C.LuaMacro_tostring(s.L, -1))
        return fmt.Errorf("lua error (load): %s", err)
    }
    // if C.lua_pcallk(s.L, 0, C.LUA_MULTRET, 0, 0, nil) != 0 {
    if C.LuaMacro_pcall(s.L, 0, C.LUA_MULTRET, 0) != 0 {
        err := C.GoString(C.lua_tolstring(s.L, -1, nil))
        return fmt.Errorf("lua error (call): %s", err)
    }
    return nil
}

func (s *LuaState) RunStringTraceback(script string) error {
    cscript := C.CString(script)
    defer C.free(unsafe.Pointer(cscript))

    // Create a message handler for our traceback
    if err := s.LoadString("local dbg = debug; return function(err) local tb = dbg.traceback(err, 2); return tb; end"); err != nil {
        return fmt.Errorf("error creating traceback handler: %s", err)
    }
    
    // Handler is now at the top of the stack (-1)
    errorHandlerIndex := C.lua_gettop(s.L)
    
    // Load the script
    if C.luaL_loadstring(s.L, cscript) != 0 {
        // err := C.GoString(C.lua_tolstring(s.L, -1, nil))
        err := C.GoString(C.LuaMacro_tostring(s.L, -1))
        C.lua_settop(s.L, errorHandlerIndex-1) // Remove error and handler
        return fmt.Errorf("lua error (load): %s", err)
    }
    
    // Script is now at the top (-1), handler at -2
    // Call with the error handler in place
    if C.lua_pcallk(s.L, 0, C.LUA_MULTRET, errorHandlerIndex, 0, nil) != 0 {
        // Get the error message with traceback from handler
        // err := C.GoString(C.lua_tolstring(s.L, -1, nil))
        err := C.GoString(C.LuaMacro_tostring(s.L, -1))
        C.lua_settop(s.L, errorHandlerIndex-1) // Remove error and handler
        return fmt.Errorf("lua error (call):\n%s", err)
    }
    
    // Remove the handler
    C.LuaMacro_remove(s.L, errorHandlerIndex)
    return nil
}

func (s *LuaState) Close() {
    C.lua_close(s.L)
    delete(luaStates, s.L)
}

// Direct wrappers

func (s *LuaState) ToString(index int) string {
    return C.GoString(C.LuaMacro_tostring(s.L, C.int(index)))
}

func (s *LuaState) ToNumber(index int) float64 {
    return float64(C.LuaMacro_tonumber(s.L, C.int(index)))
}

func (s *LuaState) PushString(val string) {
    cval := C.CString(val)
    defer C.free(unsafe.Pointer(cval))
    C.lua_pushstring(s.L, cval)
}

func (s *LuaState) PushNumber(val float64) {
	C.lua_pushnumber(s.L, C.lua_Number(val))
}

func (s *LuaState) Pop(n int) {
	// Set the top of the stack to -n-1.
	// C.lua_settop(s.L, C.int(-n-1))
    C.LuaMacro_pop(s.L, C.int(n))
}

func (s *LuaState) Insert(index int) {
	C.LuaMacro_insert(s.L, C.int(index))
}

func (s *LuaState) Remove(index int) {
	C.LuaMacro_remove(s.L, C.int(index))
}

func (s *LuaState) LoadString(script string) error {
	cscript := C.CString(script)
	defer C.free(unsafe.Pointer(cscript))
	if C.luaL_loadstring(s.L, cscript) != 0 {
		errMsg := s.ToString(-1)
		s.Pop(1) // remove the error message from the stack.
		return fmt.Errorf("lua load error: %s", errMsg)
	}
	return nil
}

func (s *LuaState) PCall(nArgs, nResults, errorFuncIndex int) error {
	// if C.lua_pcallk(s.L, C.int(nArgs), C.int(nResults), C.int(errorFuncIndex), 0, nil) != 0 {
    if C.LuaMacro_pcall(s.L, C.int(nArgs), C.int(nResults), C.int(errorFuncIndex)) != 0 {
		errMsg := s.ToString(-1)
		s.Pop(1) // remove the error message from the stack.
		return fmt.Errorf("lua error: %s", errMsg)
	}
	return nil
}

func (s *LuaState) IsNil(index int) bool {
    return C.LuaMacro_isnil(s.L, C.int(index)) != 0
}

func (s *LuaState) PushBoolean(b bool) {
    var cVal C.int
    if b {
        cVal = 1
    }
    C.lua_pushboolean(s.L, cVal)
}

func (s *LuaState) PushNil() {
    C.lua_pushnil(s.L)
}