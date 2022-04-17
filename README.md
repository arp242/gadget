gadget gets the browser and OS from the `User-Agent` header.

```go
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
short := gadget.ShortenUA(uaHeader)
fmt.Println(short)                               // ~Z (~W NT 10.0; Win64; x64) ~a537.36 ~G ~c81.0.4029.0 ~s537.36
fmt.Println(gadget.UnshortenUA(short) == uaHeader) // true
```

Some design principles:

- Just get a "common sense" name and version. Stuff like "Chrome 80.0.3987.87"
  or "Linux x86_64" is rarely useful; just get "Chrome 80" and "Linux".

- This mostly identifies the browser *engine* rather than the actual browser. It
  doesn't really matter if someone is using Opera 80, Edge 80, Samsung browser,
  or Chrome 80: they all exhibit the same behaviour, so just report it as
  "Chrome 80".

- Don't try to guess if we're dealing with a bot. Use [zgo.at/isbot][isbot] if
  you want to do that. This also doesn't go out of its way to parse the bot
  name; it's mostly intended for actual browsers people use.

- It also doesn't try to determine if this is a "mobile" browser; what does
  "mobile" even mean? Why should a 12" tablet be mobile and my 12" laptop not?
  It's usually better (and more reliable!) to rely on the screen width and/or
  use JS to determine if a client supports touch events and is "mobile".

- Don't go out of our way to get rather old User-Agents correct; some libraries
  have test cases for things like "Netscape on Windows ME", which doesn't strike
  me as very useful in 2020.

- If we don't know, then we don't know. Don't return useless values like
  "AppleWebKit 605.1.15" if there is no other information.

While this won't cover 100% of the use cases, it makes it fast and easy to use
for other use cases. Specifically, it was designed to show browser and OS stats
in [GoatCounter][gc], where you typically don't really care if someone is using
Opera or Chrome, but just want to know which browser engines your customers are
using and you need to support.

[isbot]: https://github.com/arp242/isbot
[gc]: https://github.com/arp242/goatcounter

---

Most other libraries give far too detailed information to be useful, and some
are lacking in accuracy too.

There's a small tail-end of browsers that aren't recognized correctly; only 17
out of those 371,021 are parsed to "junk data" (or 46 requests in total, out of
9.4 million). This is mostly due to people sending junk data such as misspelling
Mozilla as "Mozzila", and not too much can be done about that.

Getting it right 99.999995% of the time is good enough for me :-) It's not like
the User-Agent is reliable anyway (Ever heard of "Chrome 66.6" or "Chrome
999999"?), so this is fine.

Simple comparison benchmark:

    Library     Total (577×10)  Per op
    gadget      0.0164s         2.846µs
    uasurfer    0.0361s         6.255µs
    useragent   0.0310s         5.379µs
    user_agent  0.0275s         4.764µs
    uaparser    11.7110s        2.029633ms
