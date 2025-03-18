# Lua Bridge

A cross platform lua wrapper for go with some nice abstractions

# Build

Assuming [Brent Farris' README](https://github.com/BrentFarris/Cgo-Lua/blob/master/README.md) to be correct.
I modified the `./lua/wrapper.go` to include the `LDFLAGS` and it built fine on my mac system running:
`Darwin Kernel Version 24.3.0 -- arm64`. If the build is working for Brent, I am going to consider this covering
all platforms ∈ {MacOS, Windows, Linux} (and maybe any POSIX system) as true until proven otherwise.

You will need to ensure you have gcc on your system since this package uses [CGo](https://pkg.go.dev/cmd/cgo).

If you want to contribute, I also am assuming that you have wget, make, zip installed and in your path.

# Managing Lua

If there is a change in lua dependency for the project, the make file will change. Run `make` to install the
changes and patch the install. If you are using this project through `go get` you will not need to worry
about this as each new dependency or change will be a tag in the github repo for you to update to with `go get -u`

# Credit where credit it due

I am using the Lua official codebase. I do not own the code, I am simply using,
patching, and writting a wrapper around it to ensure that you can interop with lua without all the
hassle I went through :)

Lua us under the MIT licence. See it's licence [here](https://www.lua.org/license.html).

My jumping off point for this project was the [BrentFarris/Cgo-Lua](https://github.com/BrentFarris/Cgo-Lua)
project. It is also under the MIT licence and can be viewed [here](https://github.com/BrentFarris/Cgo-Lua/blob/master/LICENSE).

Inspiration and some points of research from [martin/go-sqlite3](https://github.com/mattn/go-sqlite3).

## Disclaimer

A lot of my code will be similar to Brent Farris' code as we have similar projects.
My project will likely implement a few more features over Brent's code and a few less.
