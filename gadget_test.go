package gadget

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// Tests are loaded from testdata/* files; Every line is a Tab-Seperated value
// of <expected browser>\t<expected OS>\t<User-Agent>.
//
// Leave a value empty if you expect nothing to be set.
//
// Comments can only start at the beginning of a line.
//
// The test name is <filename>/<line number>.
func TestParse(t *testing.T) {
	files, err := ioutil.ReadDir("./testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		t.Run(f.Name(), func(t *testing.T) {
			fp, err := os.Open("./testdata/" + f.Name())
			if err != nil {
				t.Fatal()
			}
			defer fp.Close()

			scanner := bufio.NewScanner(fp)
			i := 0
			for scanner.Scan() {
				i++
				line := scanner.Text()
				if line == "" || line[0] == '#' {
					continue
				}

				t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
					s := strings.Split(line, "\t")
					if len(s) != 3 {
						t.Fatalf("Malformed line: %q\n%#v", line, s)
					}
					got := Parse(s[2])
					if got.Browser() == s[0] && got.OS() == s[1] {
						return
					}

					t.Errorf("\nwant: %-18q %-18q\ngot:  %-18q %-18q", s[0], s[1], got.Browser(), got.OS())
				})
			}

			err = scanner.Err()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Various junk data found with the fuzzer; just ensure it won't panic.
func TestMalformed(t *testing.T) {
	tests := []string{
		"/",
		"hello/",
		"/1.0",
		")(",
		")",
		"(",
		"(CPU OS)",
		"(CPU OS00)",
		"Edge/()Chrome/",
		"(Trident/MSIE )",
	}

	for i, s := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			Parse(s)
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		in   UserAgent
		want string
	}{
		{UserAgent{}, ""},
		{UserAgent{OSName: "x"}, "x"},
		{UserAgent{BrowserName: "x"}, "x"},
		{UserAgent{BrowserName: "x", OSName: "y"}, "x on y"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tt.in.String()
			if got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestAfter(t *testing.T) {
	tests := []struct {
		in   string
		n    int
		want string
	}{
		{"", 1, ""},
		{"Hello", 0, "Hello"},
		{"Hello", 1, "ello"},
		{"Hello", 4, "o"},
		{"Hello", 5, ""},
		{"Hello", 6, ""},
		{"Hello", -1, ""},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := after(tt.in, tt.n)
			if got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	var list []string
	fp, err := os.Open("./testdata/top500")
	if err != nil {
		b.Fatal()
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '#' {
			continue
		}
		list = append(list, strings.Split(line, "\t")[2])
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Parse(list[n%len(list)])
	}
}
