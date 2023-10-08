module sim

go 1.21.2

require (
	github.com/go-gl/gl v0.0.0-20211210172815-726fda9656d6
	github.com/go-gl/glfw v0.0.0-20221017161538-93cebf72946b
	github.com/go-gl/mathgl v1.1.0
	github.com/ojrac/opensimplex-go v1.0.2
)

require (
	github.com/gerow/go-color v0.0.0-20140219113758-125d37f527f1 // indirect
	golang.org/x/image v0.0.0-20190321063152-3fc05d484e9f // indirect
)

require common v0.0.0

replace common => ../common
