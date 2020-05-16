package gadget

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

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
					got := Parse(s[2])

					if got.Browser() != s[0] || got.OS() != s[1] {
						t.Log(s[2])
					}

					if got.Browser() != s[0] {
						t.Errorf("browser\nwant: %q\ngot:  %q", s[0], got.Browser())
					}
					if got.OS() != s[1] {
						t.Errorf("OS\nwant: %q\ngot:  %q", s[1], got.OS())
					}
				})
			}

			err = scanner.Err()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMalformed(t *testing.T) {
	tests := []string{
		"/",
		"hello/",
		"/1.0",
		")(",
		")",
		"(",
		"(CPU OS)",
	}

	for i, s := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// Just ensure it won't panic.
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
