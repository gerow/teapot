package main

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

	_ "image/png"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sheenobu/go-obj/obj"
)

var cubeVertices = []float32{
	//  X, Y, Z, U, V
	// Bottom
	-1.0, -1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, -1.0, 1.0, 1.0, 1.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,

	// Top
	-1.0, 1.0, -1.0, 0.0, 0.0,
	-1.0, 1.0, 1.0, 0.0, 1.0,
	1.0, 1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, -1.0, 1.0, 0.0,
	-1.0, 1.0, 1.0, 0.0, 1.0,
	1.0, 1.0, 1.0, 1.0, 1.0,

	// Front
	-1.0, -1.0, 1.0, 1.0, 0.0,
	1.0, -1.0, 1.0, 0.0, 0.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,
	1.0, -1.0, 1.0, 0.0, 0.0,
	1.0, 1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,

	// Back
	-1.0, -1.0, -1.0, 0.0, 0.0,
	-1.0, 1.0, -1.0, 0.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	-1.0, 1.0, -1.0, 0.0, 1.0,
	1.0, 1.0, -1.0, 1.0, 1.0,

	// Left
	-1.0, -1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, -1.0, 1.0, 0.0,
	-1.0, -1.0, -1.0, 0.0, 0.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0,
	-1.0, 1.0, -1.0, 1.0, 0.0,

	// Right
	1.0, -1.0, 1.0, 1.0, 1.0,
	1.0, -1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, 1.0, 1.0, 1.0,
	1.0, 1.0, -1.0, 0.0, 0.0,
	1.0, 1.0, 1.0, 0.0, 1.0,
}

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func ReadObj(name string) (*obj.Object, error) {
	f, err := os.Open(path.Join("assets", name))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return obj.NewReader(f).Read()
}

func LoadObj(o *obj.Object) uint32 {
	stride := 8
	// X, Y, Z, nX, nY, nZ, U, V
	vertices := make([]float32, len(o.Faces)*stride*3)

	for i, f := range o.Faces {
		if len(f.Points) != 3 {
			panic("We only do triangles!")
		}

		for j, p := range f.Points {
			vertices[(i*3+j)*stride] = float32(p.Vertex.X)
			vertices[(i*3+j)*stride+1] = float32(p.Vertex.Y)
			vertices[(i*3+j)*stride+2] = float32(p.Vertex.Z)
			vertices[(i*3+j)*stride+3] = float32(p.Normal.X)
			vertices[(i*3+j)*stride+4] = float32(p.Normal.Y)
			vertices[(i*3+j)*stride+5] = float32(p.Normal.Z)
			vertices[(i*3+j)*stride+6] = float32(p.Texture.U)
			vertices[(i*3+j)*stride+7] = float32(p.Texture.V)
		}
	}

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	return vao
}

func LoadCube() uint32 {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(cubeVertices)*4, gl.Ptr(cubeVertices), gl.STATIC_DRAW)

	return vao
}

func CompileAndLinkProgram(vert, frag string) (uint32, error) {
	v, err := CompileShader(vert, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	defer gl.DeleteShader(v)

	f, err := CompileShader(frag, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}
	defer gl.DeleteShader(f)

	p := gl.CreateProgram()
	gl.AttachShader(p, v)
	gl.AttachShader(p, f)
	gl.LinkProgram(p)

	var status int32
	gl.GetProgramiv(p, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var len int32
		gl.GetProgramiv(p, gl.INFO_LOG_LENGTH, &len)
		log := strings.Repeat("\x00", int(len+1))
		gl.GetProgramInfoLog(p, len, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link %s and %s: %v", frag, vert, log)
	}

	return p, nil
}

func CompileShader(name string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	b, err := ioutil.ReadFile(path.Join("assets", name))
	if err != nil {
		return 0, err
	}
	// Ensure it's null terminated.
	b = append(b, 0)

	bCStr, free := gl.Strs(string(b))
	defer free()
	gl.ShaderSource(shader, 1, bCStr, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", name, log)
	}

	return shader, nil
}

func LoadTexture(name string) (uint32, error) {
	f, err := os.Open(path.Join("assets", name))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, errors.New("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}

func main() {
	// Initialize glfw.
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	windowWidth := 640
	windowHeight := 480
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize gl.
	if err := gl.Init(); err != nil {
		panic(err)
	}
	fmt.Println("OpenGL version", gl.GoStr(gl.GetString(gl.VERSION)))

	// Create our program
	program, err := CompileAndLinkProgram("basic.vert", "basic.frag")
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)

	// Bind our shader uniforms
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/float32(windowHeight), 0.1, 10.0)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	camera := mgl32.LookAtV(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	textureUniform := gl.GetUniformLocation(program, gl.Str("tex\x00"))
	gl.Uniform1i(textureUniform, 0)

	ambientUniform := gl.GetUniformLocation(program, gl.Str("ambient\x00"))
	gl.Uniform3f(ambientUniform, 0.2, 0.2, 0.2)

	sunDirUniform := gl.GetUniformLocation(program, gl.Str("sunDir\x00"))
	sunDir := mgl32.Vec3{1.0, -1.0, 1.0}.Normalize()
	gl.Uniform3fv(sunDirUniform, 1, &sunDir[0])

	sunDiffuseUniform := gl.GetUniformLocation(program, gl.Str("sunDiffuse\x00"))
	gl.Uniform3f(sunDiffuseUniform, 0.8, 0.8, 0.8)

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// Load the texture
	texture, err := LoadTexture("square.png")
	if err != nil {
		panic(err)
	}

	// Load our teapot
	monkey, err := ReadObj("suzanne.obj")
	if err != nil {
		panic(err)
	}
	monkeyVAO := LoadObj(monkey)
	//monkeyVAO := LoadCube()

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))

	normalAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertNormal\x00")))
	gl.EnableVertexAttribArray(normalAttrib)
	gl.VertexAttribPointer(normalAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))

	texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib)
	gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(6*4))

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)

	angle := 0.0
	previousTime := glfw.GetTime()

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Update
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		angle += elapsed
		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})

		//ambient := mgl32.Vec3{
		//	float32(0.2 * math.Sin(time)),
		//	float32(0.2 * math.Sin(time+2.0*math.Pi/3.0)),
		//	float32(0.2 * math.Sin(time+4.0*math.Pi/3.0)),
		//}
		//gl.Uniform3fv(ambientUniform, 1, &ambient[0])

		// Render
		gl.UseProgram(program)
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		gl.BindVertexArray(monkeyVAO)

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		gl.DrawArrays(gl.TRIANGLES, 0, 968*3)

		// Do OpenGL stuff.
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
