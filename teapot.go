package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/sheenobu/go-obj/obj"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func main() {
	// Load our teapot
	f, err := os.Open("assets/teapot.obj")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	o, err := obj.NewReader(f).Read()
	if err != nil {
		panic(err)
	}
	fmt.Printf("This is the object %v\n", o)

	// Start gl!
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	for !window.ShouldClose() {
		// Do OpenGL stuff.
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
