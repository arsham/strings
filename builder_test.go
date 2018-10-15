package strings_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	. "github.com/arsham/strings"
)

const N = 10000       // make this bigger for a larger (and slower) test
var testString string // test data for write tests
var testBytes []byte  // test data; same as testString but as a slice.

func init() {
	testBytes = make([]byte, N)
	for i := 0; i < N; i++ {
		testBytes[i] = 'a' + byte(i%26)
	}
	testString = string(testBytes)
}

func check(t *testing.T, b *Builder, want string) {
	got := b.String()
	if got != want {
		t.Errorf("String: got %#q; want %#q", got, want)
		return
	}
	if n := b.Len(); n != len(got) {
		t.Errorf("Len: got %d; but len(String()) is %d", n, len(got))
	}
}

func TestBuilder(t *testing.T) {
	var b Builder
	check(t, &b, "")
	n, err := b.WriteString("hello")
	if err != nil || n != 5 {
		t.Errorf("WriteString: got %d,%s; want 5,nil", n, err)
	}
	check(t, &b, "hello")
	if err = b.WriteByte(' '); err != nil {
		t.Errorf("WriteByte: %s", err)
	}
	check(t, &b, "hello ")
	n, err = b.WriteString("world")
	if err != nil || n != 5 {
		t.Errorf("WriteString: got %d,%s; want 5,nil", n, err)
	}
	check(t, &b, "hello world")
}

func TestBuilderString(t *testing.T) {
	var b Builder
	b.WriteString("alpha")
	check(t, &b, "alpha")
	s1 := b.String()
	b.WriteString("beta")
	check(t, &b, "alphabeta")
	s2 := b.String()
	b.WriteString("gamma")
	check(t, &b, "alphabetagamma")
	s3 := b.String()

	// Check that subsequent operations didn't change the returned strings.
	if want := "alpha"; s1 != want {
		t.Errorf("first String result is now %q; want %q", s1, want)
	}
	if want := "alphabeta"; s2 != want {
		t.Errorf("second String result is now %q; want %q", s2, want)
	}
	if want := "alphabetagamma"; s3 != want {
		t.Errorf("third String result is now %q; want %q", s3, want)
	}
}

func TestBuilderReset(t *testing.T) {
	var b Builder
	check(t, &b, "")
	b.WriteString("aaa")
	s := b.String()
	check(t, &b, "aaa")
	b.Reset()
	check(t, &b, "")

	// Ensure that writing after Reset doesn't alter
	// previously returned strings.
	b.WriteString("bbb")
	check(t, &b, "bbb")
	if want := "aaa"; s != want {
		t.Errorf("previous String result changed after Reset: got %q; want %q", s, want)
	}
}

func TestBuilderGrow(t *testing.T) {
	for _, growLen := range []int{0, 100, 1000, 10000, 100000} {
		p := bytes.Repeat([]byte{'a'}, growLen)
		allocs := testing.AllocsPerRun(100, func() {
			var b Builder
			b.Grow(growLen) // should be only alloc, when growLen > 0
			b.Write(p)
			if b.String() != string(p) {
				t.Fatalf("growLen=%d: bad data written after Grow", growLen)
			}
		})
		wantAllocs := 1
		if growLen == 0 {
			wantAllocs = 0
		}
		if g, w := int(allocs), wantAllocs; g != w {
			t.Errorf("growLen=%d: got %d allocs during Write; want %v", growLen, g, w)
		}
	}
}

func TestBuilderWrite2(t *testing.T) {
	const s0 = "hello 世界"
	for _, tt := range []struct {
		name string
		fn   func(b *Builder) (int, error)
		n    int
		want string
	}{
		{
			"Write",
			func(b *Builder) (int, error) { return b.Write([]byte(s0)) },
			len(s0),
			s0,
		},
		{
			"WriteRune",
			func(b *Builder) (int, error) { return b.WriteRune('a') },
			1,
			"a",
		},
		{
			"WriteRuneWide",
			func(b *Builder) (int, error) { return b.WriteRune('世') },
			3,
			"世",
		},
		{
			"WriteString",
			func(b *Builder) (int, error) { return b.WriteString(s0) },
			len(s0),
			s0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b Builder
			n, err := tt.fn(&b)
			if err != nil {
				t.Fatalf("first call: got %s", err)
			}
			if n != tt.n {
				t.Errorf("first call: got n=%d; want %d", n, tt.n)
			}
			check(t, &b, tt.want)

			n, err = tt.fn(&b)
			if err != nil {
				t.Fatalf("second call: got %s", err)
			}
			if n != tt.n {
				t.Errorf("second call: got n=%d; want %d", n, tt.n)
			}
			check(t, &b, tt.want+tt.want)
		})
	}
}

func TestBuilderWriteByte(t *testing.T) {
	var b Builder
	if err := b.WriteByte('a'); err != nil {
		t.Error(err)
	}
	if err := b.WriteByte(0); err != nil {
		t.Error(err)
	}
	check(t, &b, "a\x00")
}

func TestBuilderWriteBytes(t *testing.T) {
	var b Builder
	input := []byte("abcd")
	n, err := b.WriteBytes(input)
	if err != nil {
		t.Error(err)
	}
	if n != len(input) {
		t.Errorf("input=%q, got %d, want %d", input, n, len(input))
	}
	if err := b.WriteByte(0); err != nil {
		t.Error(err)
	}
	check(t, &b, "abcd\x00")
}

func TestBuilderAllocs(t *testing.T) {
	var b Builder
	const msg = "hello"
	b.Grow(len(msg) * 2) // because AllocsPerRun does an extra "warm-up" iteration
	var s string
	allocs := int(testing.AllocsPerRun(1, func() {
		b.WriteString("hello")
		s = b.String()
	}))
	if want := msg + msg; s != want {
		t.Errorf("String: got %#q; want %#q", s, want)
	}
	if allocs > 0 {
		t.Fatalf("got %d alloc(s); want 0", allocs)
	}

	// Issue 23382; verify that copyCheck doesn't force the
	// Builder to escape and be heap allocated.
	n := testing.AllocsPerRun(10000, func() {
		var b Builder
		b.Grow(5)
		b.WriteString("abcde")
		_ = b.String()
	})
	if n != 1 {
		t.Errorf("Builder allocs = %v; want 1", n)
	}
}

func TestBuilderCopyPanic(t *testing.T) {
	tests := []struct {
		name      string
		fn        func()
		wantPanic bool
	}{
		{
			name:      "String",
			wantPanic: false,
			fn: func() {
				var a Builder
				a.WriteByte('x')
				b := a
				_ = b.String() // appease vet
			},
		},
		{
			name:      "Len",
			wantPanic: false,
			fn: func() {
				var a Builder
				a.WriteByte('x')
				b := a
				b.Len()
			},
		},
		{
			name:      "Reset",
			wantPanic: false,
			fn: func() {
				var a Builder
				a.WriteByte('x')
				b := a
				b.Reset()
				b.WriteByte('y')
			},
		},
		{
			name:      "Write",
			wantPanic: true,
			fn: func() {
				var a Builder
				a.Write([]byte("x"))
				b := a
				b.Write([]byte("y"))
			},
		},
		{
			name:      "WriteByte",
			wantPanic: true,
			fn: func() {
				var a Builder
				a.WriteByte('x')
				b := a
				b.WriteByte('y')
			},
		},
		{
			name:      "WriteString",
			wantPanic: true,
			fn: func() {
				var a Builder
				a.WriteString("x")
				b := a
				b.WriteString("y")
			},
		},
		{
			name:      "WriteRune",
			wantPanic: true,
			fn: func() {
				var a Builder
				a.WriteRune('x')
				b := a
				b.WriteRune('y')
			},
		},
		{
			name:      "Grow",
			wantPanic: true,
			fn: func() {
				var a Builder
				a.Grow(1)
				b := a
				b.Grow(2)
			},
		},
	}
	for _, tt := range tests {
		didPanic := make(chan bool)
		go func() {
			defer func() { didPanic <- recover() != nil }()
			tt.fn()
		}()
		if got := <-didPanic; got != tt.wantPanic {
			t.Errorf("%s: panicked = %v; want %v", tt.name, got, tt.wantPanic)
		}
	}
}

var someBytes = []byte("some bytes sdljlk jsklj3lkjlk djlkjw")

var sinkS string

func benchmarkBuilder(b *testing.B, f func(b *testing.B, numWrite int, grow bool)) {
	b.Run("1Write_NoGrow", func(b *testing.B) {
		b.ReportAllocs()
		f(b, 1, false)
	})
	b.Run("3Write_NoGrow", func(b *testing.B) {
		b.ReportAllocs()
		f(b, 3, false)
	})
	b.Run("3Write_Grow", func(b *testing.B) {
		b.ReportAllocs()
		f(b, 3, true)
	})
}

func BenchmarkBuildString_Builder(b *testing.B) {
	benchmarkBuilder(b, func(b *testing.B, numWrite int, grow bool) {
		for i := 0; i < b.N; i++ {
			var buf Builder
			if grow {
				buf.Grow(len(someBytes) * numWrite)
			}
			for i := 0; i < numWrite; i++ {
				buf.Write(someBytes)
			}
			sinkS = buf.String()
		}
	})
}

func BenchmarkBuildString_ByteBuffer(b *testing.B) {
	benchmarkBuilder(b, func(b *testing.B, numWrite int, grow bool) {
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			if grow {
				buf.Grow(len(someBytes) * numWrite)
			}
			for i := 0; i < numWrite; i++ {
				buf.Write(someBytes)
			}
			sinkS = buf.String()
		}
	})
}

//
// Tests for added functionalities. Mostly copied from bytes.Buffer, because the
// functionalities mostly match.
//

// Verify that contents of b match the string s.
func checkRead(t *testing.T, testname string, b *Builder, s string) {
	bytes := b.Bytes()
	str := b.String()
	if b.Len() != len(str) {
		t.Errorf("%s: b.Len() == %d, len(b.String()) == %d", testname, b.Len(), len(str))
	}
	if string(bytes) != s {
		t.Errorf("%s: string(b.Bytes()) == %q, s == %q", testname, string(bytes), s)
	}
}

// Fill b through n writes of string fus.
// The initial contents of b corresponds to the string s;
// the result is the final contents of b returned as a string.
func fillString(t *testing.T, testname string, b *Builder, s string, n int, fus string) string {
	checkRead(t, testname+" (fill 1)", b, s)
	for ; n > 0; n-- {
		m, err := b.WriteString(fus)
		if m != len(fus) {
			t.Errorf(testname+" (fill 2): m == %d, expected %d", m, len(fus))
		}
		if err != nil {
			t.Errorf(testname+" (fill 3): err should always be nil, found err == %s", err)
		}
		s += fus
		checkRead(t, testname+" (fill 4)", b, s)
	}
	return s
}

// Fill b through n writes of byte slice fub.
// The initial contents of b corresponds to the string s;
// the result is the final contents of b returned as a string.
func fillBytes(t *testing.T, testname string, b *Builder, s string, n int, fub []byte) string {
	checkRead(t, testname+" (fill 1)", b, s)
	for ; n > 0; n-- {
		m, err := b.Write(fub)
		if m != len(fub) {
			t.Errorf(testname+" (fill 2): m == %d, expected %d", m, len(fub))
		}
		if err != nil {
			t.Errorf(testname+" (fill 3): err should always be nil, found err == %s", err)
		}
		s += string(fub)
		checkRead(t, testname+" (fill 4)", b, s)
	}
	return s
}

// Empty b through repeated reads into fub.
// The initial contents of b corresponds to the string s.
func empty(t *testing.T, testname string, b *Builder, s string, fub []byte) {
	checkRead(t, testname+" (empty 1)", b, s)

	for {
		n, err := b.Read(fub)
		if n == 0 {
			break
		}
		if err != nil {
			t.Errorf(testname+" (empty 2): err should always be nil, found err == %s", err)
		}
		s = s[n:]
		checkRead(t, testname+" (empty 3)", b, s)
	}

	checkRead(t, testname+" (empty 4)", b, "")
}

func TestMixedReadsAndWrites(t *testing.T) {
	var b Builder
	s := ""
	for i := 0; i < 50; i++ {
		wlen := rand.Intn(len(testString))
		testName := fmt.Sprintf("i=%d: TestMixedReadsAndWrites (1)", i)
		if i%2 == 0 {
			s = fillString(t, testName, &b, s, 1, testString[0:wlen])
		} else {
			s = fillBytes(t, testName, &b, s, 1, testBytes[0:wlen])
		}

		rlen := rand.Intn(len(testString))
		fub := make([]byte, rlen)
		n, _ := b.Read(fub)
		s = s[n:]
	}
	empty(t, "TestMixedReadsAndWrites (2)", &b, s, make([]byte, b.Len()))
}

// Was a bug: used to give EOF reading empty slice at EOF.
func TestReadEmptyAtEOF(t *testing.T) {
	b := new(Builder)
	slice := make([]byte, 0)
	n, err := b.Read(slice)
	if err != nil {
		t.Errorf("read error: %v", err)
	}
	if n != 0 {
		t.Errorf("wrong count; got %d want 0", n)
	}
}

func TestLargeByteReads(t *testing.T) {
	var b Builder
	for i := 3; i < 30; i += 3 {
		s := fillBytes(t, "TestLargeReads (1)", &b, "", 5, testBytes[0:len(testBytes)/i])
		empty(t, "TestLargeReads (2)", &b, s, make([]byte, len(testString)))
	}
	checkRead(t, "TestLargeByteReads (3)", &b, "")
}

// From Issue 5154.
func BenchmarkBuilderNotEmptyWriteRead(b *testing.B) {
	buf := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		var b Builder
		b.Write(buf[0:1])
		for i := 0; i < 5<<10; i++ {
			b.Write(buf)
			b.Read(buf)
		}
	}
}
