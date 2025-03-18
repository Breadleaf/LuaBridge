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

extern int panic_callback(lua_State* L);
extern int pcallk_callback(lua_State* L, int status, lua_KContext ctx);
extern int cclosure_callback(lua_State* L);
extern int print_stack(lua_State* lua);
*/
import "C"

func Hello() {
	println("hello")
}
