package main

import (
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	SENT    = "SENT"
	PENDING = "PENDING"
)

func main() {
	fmt.Print("Learning...")

	a, b := swap("a", "b")
	fmt.Println(a, b)

	// variables
	var x, y int = 3, 4
	var f float64 = math.Sqrt(float64(x*x + y*y))
	fmt.Println(f)

	// constants
	const c = "ciao"
	fmt.Println(c)
	fmt.Println(SENT)

	// iterations
	sum := 0
	for i := 0; i < 10; i++ {
		sum += 1
	}
	fmt.Println("Sum", sum)

	// conditions
	if sum <= 10 {
		fmt.Println("Sum is less or equal than 10")
	} else {
		fmt.Println("Sum is greater than than 10")
	}

	fmt.Println(pow(4, 2, 23))

	printRuntime(runtime.GOOS)
	isSaturday(time.Now())

	// switch
	t := time.Now()
	switch {
	case t.Hour() < 12:
		fmt.Println("Good morning!")
	case t.Hour() < 17:
		fmt.Println("Good afternoon.")
	default:
		fmt.Println("Good evening.")
	}

	defer fmt.Println("I'm last print")

	panicAndRecover()

	// pointers
	var p *int
	i := 42
	p = &i
	fmt.Println(*p)

	// arrays
	var arr [2]string
	arr[0] = "elem1"
	arr[1] = "elem2"
	fmt.Println(arr)
	fmt.Println([3]int{1, 2, 3})

	// slices are only reference to arrayy
	var arr2 = [4]int{1, 2, 3, 4}
	fmt.Println(arr2[:2])
	fmt.Println(arr2[2:4])
	arr2[:2][0] = 10
	fmt.Println(arr2)
	fmt.Println("Capacity of slice", cap(arr2[:2]))
	fmt.Println("Length of slice", len(arr2[:2]))

	// for range
	for i, v := range []int{1, 4, 3} {
		fmt.Printf("2**%d = %d\n", i, v)
	}

	// maps
	var m = make(map[string]string)
	m["a"] = "b"
	fmt.Println(m)
	var m2 = map[string]string{
		"a": "b",
	}
	fmt.Println(m2)

	// high order
	fmt.Println(compute(2, 3, math.Pow))
	fmt.Println(compute(2, 3, func(x float64, y float64) float64 {
		return x / y
	}))

	// methods
	v := Vertex{
		X: 1,
		Y: 2,
	}
	fmt.Println("Vertex", v)
	fmt.Println("Vertex abs", v.Abs())
	v.Scale(10)
	fmt.Println("Vertex scale", v)

	// interfaces
	var abser Abser
	vertex := Vertex{10, 23}
	abser = &vertex
	fmt.Println(abser.Abs())

	// switch by types
	doByType("ciao")
	doByType(123)
	doByType(4.0)

	// to string
	fmt.Println(v)

	// useRot13
	useRot13()

	tryImage()

	// type parameters
	si := []int{10, 20, 15, -10}
	ss := []string{"foo", "bar", "baz"}
	fmt.Println(Index(si, 15))
	fmt.Println(Index(ss, "hello"))

	var example = sms{
		ID:      "12343",
		Content: "Example",
		From:    "123456789",
		To:      "123456789",
	}

	example.From = "2"
	fmt.Println(example)

	goroutines()
}

type sms struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	From    string `json:"from"`
	To      string `json:"to"`
}

func swap(x, y string) (string, string) {
	return y, x
}

func pow(x, n, lim float64) float64 {
	if v := math.Pow(x, n); v < lim {
		return v
	}
	return lim
}

func printRuntime(runtime string) {
	switch os := runtime; os {
	case "darwin":
		fmt.Println("OS X")
	case "linux":
		fmt.Println("Linux")
	default:
		fmt.Printf("%s \n", os)
	}
}

func isSaturday(now time.Time) {
	today := now.Weekday()
	switch time.Saturday {
	case today + 0:
		fmt.Println("Today.")
	case today + 1:
		fmt.Println("Tomorrow.")
	case today + 2:
		fmt.Println("In two days.")
	default:
		fmt.Println("Too far away.")
	}
}

func panicAndRecover() {
	untilPanic := 3
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered function", r)
		}
	}()
	willPanic(untilPanic)
}

func willPanic(panicCount int) {
	if panicCount <= 0 {
		fmt.Println("Going to panic")
		panic(fmt.Sprintf(""))
	}
	defer fmt.Println("Deferring panic", panicCount)
	willPanic(panicCount - 1)
}

// high order
func compute(a, b float64, fn func(float64, float64) float64) float64 {
	return fn(a, b)
}

// methods

type Vertex struct {
	X, Y float64
}

func (v Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v *Vertex) Scale(f float64) {
	v.X = v.X * f
	v.Y = v.Y * f
}

// interfaces

type Abser interface {
	Abs() float64
}

func doByType(i interface{}) {
	switch i.(type) {
	case string:
		fmt.Println("Is string")
	case int:
		fmt.Println("Is integer")
	default:
		fmt.Println("Is unknown type")
	}
}

// to string of fmt package
func (v Vertex) String() string {
	return fmt.Sprintf("X: %f, Y: %f", v.X, v.Y)
}

// rot13 reader
type Rot13Reader struct {
	r io.Reader
}

func useRot13() {
	s := strings.NewReader("Lbh penpxrq gur pbqr!")
	r := Rot13Reader{s}
	io.Copy(os.Stdout, &r)
}

func (rot Rot13Reader) Read(b []byte) (int, error) {
	var n, err = rot.r.Read(b)
	if n > 0 {
		for i := 0; i < n; i++ {
			switch v := b[i]; {
			case v >= 'a' && v <= 'z':
				b[i] = (b[i]-'a'+13)%26 + 'a'
			case v >= 'A' && v <= 'Z':
				b[i] = ((b[i] - 'A' + 13) % 26) + 'A'
			}
		}
	}
	return n, err
}

func tryImage() {
	m := image.NewRGBA(image.Rect(0, 0, 100, 100))
	fmt.Println(m.Bounds())
	fmt.Println(m.At(0, 0).RGBA())
}

// type parameters and generics
func Index[T comparable](s []T, x T) int {
	for i, v := range s {
		if v == x {
			return i
		}
	}
	return -1
}

type List[T any] struct {
	next *List[T]
	val  T
}
