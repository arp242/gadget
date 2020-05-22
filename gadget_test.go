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
					got := Parse(Unshorten(s[2]))
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

func TestShorten(t *testing.T) {
	tests := []struct {
		ua, short string
	}{
		{"", ""},
		{"~m~~~A~", "~~m~~~~~~A~~"},
		{
			`Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:73.0) Gecko/20100101 Firefox/73.0`,
			`~Z (~W NT 10.0; Win64; x64; rv:73.0) ~g20100101 ~f73.0`,
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36",
			"~Z (~W NT 10.0; Win64; x64) ~a537.36 ~G ~c80.0.3987.132 ~s537.36",
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:73.0) Gecko/20100101 Firefox/73.0",
			"~Z (~W NT 10.0; Win64; x64; rv:73.0) ~g20100101 ~f73.0",
		},
		{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.5 Safari/605.1.15",
			"~Z (~I; Intel Mac OS X 10_15_3) ~a605.1.15 ~G ~v13.0.5 ~s605.1.15",
		},
		{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36",
			"~Z (~I; Intel Mac OS X 10_15_3) ~a537.36 ~G ~c80.0.3987.132 ~s537.36",
		},
		{
			"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:73.0) Gecko/20100101 Firefox/73.0",
			"~Z (X11; Ubuntu; ~L x86_64; rv:73.0) ~g20100101 ~f73.0",
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.102 Safari/537.36 Edge/18.18362",
			"~Z (~W NT 10.0; Win64; x64) ~a537.36 ~G ~c70.0.3538.102 ~s537.36 ~e18.18362",
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36 OPR/66.0.3515.72",
			"~Z (~W NT 10.0; Win64; x64) ~a537.36 ~G ~c79.0.3945.130 ~s537.36 OPR/66.0.3515.72",
		},
		{
			"Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.0 Mobile/15E148 Safari/604.1",
			"~Z (~i; CPU ~i OS 12_0 like Mac OS X) ~a605.1.15 ~G ~v12.0 ~m15E148 ~s604.1",
		},
		{
			"Mozilla/5.0 (Linux; Android 8.0.0; SM-G960F Build/R16NW) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.84 Mobile Safari/537.36",
			"~Z (~L; ~A 8.0.0; SM-G960F Build/R16NW) ~a537.36 ~G ~c62.0.3202.84 ~M ~s537.36",
		},
		{
			"Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.92 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			"~Z (~L; ~A 6.0.1; Nexus 5X Build/MMB29P) ~a537.36 ~G ~c80.0.3987.92 ~M ~s537.36 (~C; Googlebot/2.1; +http://www.google.com/bot.html)",
		},
		{
			"Mozilla/5.0 (compatible; AhrefsBot/6.1; +http://ahrefs.com/robot/)",
			"~Z (~C; AhrefsBot/6.1; +http://ahrefs.com/robot/)",
		},
		{
			"Mozilla/5.0 (Linux; Android 5.0) AppleWebKit/537.36 (KHTML, like Gecko) Mobile Safari/537.36 (compatible; Bytespider; https://zhanzhang.toutiao.com/)",
			"~Z (~L; ~A 5.0) ~a537.36 ~G ~M ~s537.36 (~C; Bytespider; https://zhanzhang.toutiao.com/)",
		},
		{
			"Mozilla/5.0 (compatible; Nimbostratus-Bot/v1.3.2; http://cloudsystemnetworks.com)",
			"~Z (~C; Nimbostratus-Bot/v1.3.2; http://cloudsystemnetworks.com)",
		},
		{
			"Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)",
			"~Z (~C; YandexBot/3.0; +http://yandex.com/bots)",
		},
		{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/600.2.5 (KHTML, like Gecko) Version/8.0.2 Safari/600.2.5 (Applebot/0.1; +http://www.apple.com/go/applebot)",
			"~Z (~I; Intel Mac OS X 10_10_1) ~a600.2.5 ~G ~v8.0.2 ~s600.2.5 (Applebot/0.1; +http://www.apple.com/go/applebot)",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			short := Shorten(tt.ua)
			if short != tt.short {
				t.Errorf("Shorten\ngot:  %q\nwant: %q", short, tt.short)
			}

			unshort := Unshorten(short)
			if unshort != tt.ua {
				t.Errorf("Unshorten\ngot:  %q\nwant: %q", unshort, tt.ua)
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
