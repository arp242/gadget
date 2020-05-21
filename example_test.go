package gadget_test

import (
	"fmt"

	"zgo.at/gadget"
)

func ExampleParse() {
	ua := gadget.Parse(`Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:73.0) Gecko/20100101 Firefox/73.0`)

	fmt.Println(ua.String())  // "Firefox 73 on Windows 10"
	fmt.Println(ua.Browser()) // "Firefox 73"
	fmt.Println(ua.OS())      // "Windows 10"

	// Or for more detailed information:
	fmt.Println(ua.BrowserName)    // "Firefox"
	fmt.Println(ua.BrowserVersion) // "73"
	fmt.Println(ua.OSName)         // "Windows"
	fmt.Println(ua.OSVersion)      // "10"

	// Helper to shorten the UA string while remaining readable:
	uaHeader := `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4029.0 Safari/537.36`
	short := gadget.Shorten(uaHeader)
	fmt.Println(short)                               // ~Z (~W NT 10.0; Win64; x64) ~a537.36 ~G ~c81.0.4029.0 ~s537.36
	fmt.Println(gadget.Unshorten(short) == uaHeader) // true

	// Output:
	// Firefox 73 on Windows 10
	// Firefox 73
	// Windows 10
	// Firefox
	// 73
	// Windows
	// 10
	// ~Z (~W NT 10.0; Win64; x64) ~a537.36 ~G ~c81.0.4029.0 ~s537.36
	// true
}
