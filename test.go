package main_deprecated 

import (
	"fmt"
	"slices"

	// "runtime"
	"strings"

	"github.com/Breadleaf/LuaBridge/lua"
)

func main_deprecated() {
	// Lock the OS thread to avoid signal issues when mixing Go and C
	// runtime.LockOSThread()

	L := lua.NewState()
	defer L.Close()

// 	relaxedRequirePatch := strings.TrimSpace(`
// -- bootstrap procedure start --
// _G.__defaultRequire = require
// _G.require = function( moduleName )
// 	local module = import( moduleName )
// 	if type( module ) == "string" and module:find( "not found" ) then
// 		return _G.__defaultRequire( moduleName )
// 	else
// 		return module
// 	end
// end
// --- bootstrap procedure end ---
// 	`)

	strictRequirePatch := strings.TrimSpace(`
-- bootstrap procedure start --
_G.__defaultRequire = require
_G.require = function(moduleName)
    local module, err = import(moduleName)
    if module == nil then
        if string.find(err or "", "blacklisted") then
            error(err, 2)
        end
        if not string.find(err or "", "blacklisted") then
            local success, result = pcall(_G.__defaultRequire, moduleName)
            if success then
                return result
            else
                error(err .. "\n" .. tostring(result), 2)
            end
        else
            error(err, 2)
        end
    else
        return module
    end
end
--- bootstrap procedure end ---
	`)

	program := strings.TrimSpace(`
function test()
	-- causes a stack trace since os was removed
	print( os.date() )
end

-- test that black list works and is compliant with either relaxed or strict require patch
-- NOTE: if using strict require patch, you will need to comment both out as in vanilla lua it
--       will cause a stack trace and hault execution (this is the behaviour of strict patch)
-- print( "NOT_REAL_MODULE_OR_IMPORT", require( "NOT_REAL_MODULE_OR_IMPORT" ) )
-- print( "os", require( "os" ) )

-- test _G.require on default module
print( require( "math" ) )
-- local math = require( "math" )
print( math.sin( 190 ) )

-- test _G.require on custom import
local mmRequire = require( "mymod" )
mmRequire.sayHello()

-- test _G.import on custom import
local mmImport = import( "mymod" )
mmImport.sayHello()

-- test stack trace w/ program crash
-- test remove from global scope
test()
	`)

	patchProgram := func(source string, patches []string) string {
		var patchedProgram string
		for _, patch := range patches {
			patchedProgram += patch + "\n"
		}
		patchedProgram += source
		return patchedProgram
	}

	// patchedProgram := patchProgram(program, []string{
	// 	relaxedRequirePatch,
	// })

	patchedProgram := patchProgram(program, []string{
		strictRequirePatch,
	})

	fmt.Printf(
		"%s\n%s\n%s\n",
		"[ PATCHED PROGRAM START ]",
		patchedProgram,
		"[  PATCHED PROGRAM END  ]",
	)

	var imports = map[string]string{
		"mymod": strings.TrimSpace(`
local M = {}
function M.sayHello()
	print("Hello from mymod!")
end
return M
		`),
	}

	var blacklist = []string{
		"os",
	}

	L.RemoveGlobal("os")

	// relaxed import
	// L.RegisterGlobalFunction("import", func(L *lua.LuaState) int {
	// 	importName := L.ToString(1)

	// 	if found := slices.Contains(blacklist, importName); found {
	// 		L.PushNil()
	// 		L.PushString(fmt.Sprintf("module '%s' is blacklisted.", importName))
	// 		return 2
	// 	}

	// 	source, ok := imports[importName]
	// 	if !ok {
	// 		L.PushNil()
	// 		errorMsg := fmt.Sprintf("module '%s' not found:\n\tno custom import found", importName)
	// 		L.PushString(errorMsg)
	// 		return 2
	// 	}

	// 	if err := L.LoadString(source); err != nil {
	// 		L.PushNil()
	// 		L.PushString(err.Error())
	// 		return 2
	// 	}

	// 	if err := L.PCall(0, 1, 0); err  != nil {
	// 		L.PushNil()
	// 		L.PushString(err.Error())
	// 		return 2
	// 	}

	// 	// if import didnt return, push true (default lua module behavior)
	// 	if L.IsNil(-1) {
	// 		L.Pop(1)
	// 		L.PushBoolean(true)
	// 	}

	// 	return 1
	// })

	// strict import
	L.RegisterGlobalFunction("import", func(L *lua.LuaState) int {
		importName := L.ToString(1)

		if found := slices.Contains(blacklist, importName); found {
			L.PushNil()
			errorMsg := fmt.Sprintf("module '%s' not found:\n\tmodule is blacklisted", importName)
			L.PushString(errorMsg)
			return 2
		}

		source, ok := imports[importName]
		if !ok {
			L.PushNil()
			errorMsg := fmt.Sprintf("module '%s' not found:\n\tno field package.preload['%s']\n\tno custom import found",
				importName, importName)
			L.PushString(errorMsg)
			return 2
		}

		if err := L.LoadString(source); err != nil {
			L.PushNil()
			L.PushString(err.Error())
			return 2
		}

		if err := L.PCall(0, 1, 0); err != nil {
			L.PushNil()
			L.PushString(err.Error())
			return 2
		}

		if L.IsNil(-1) {
			L.Pop(1)
			L.PushBoolean(true)
		}

		return 1
	})

	fmt.Println("RunStringError()")
	if err := L.RunStringError(patchedProgram); err != nil {
		fmt.Println(err.Error())
	}

	/*
	fmt.Println("RunStringTraceback()")
	if err := L.RunStringTraceback(patchedProgram); err != nil {
		fmt.Println(err.Error())
	}
	*/
}
