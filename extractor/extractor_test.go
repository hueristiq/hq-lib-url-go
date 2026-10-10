package extractor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hueristiq/hq-lib-url-go/extractor"
)

// corpus is a shared sample text covering the URL forms the extractor recognizes:
// scheme URLs (with user info, ports, queries, fragments), Unicode and punycode
// domains, IPv4 and IPv6 hosts, emails, no-authority schemes, relative paths, and
// scheme-less hosts.
//
//nolint:gosec // G101: the credentials in the corpus are fake, exercising user info matching.
const corpus = `
	http://localhost
	http://localhost:8000/home

	https://www.example.com
	http://www.example.com:8080/page
	https://www.example.com/search?q=openai
	https://www.example.com/resources/docs/guide
	https://www.example.com/about#team
	https://user:password@example.com
	https://www.example.com:8080/search?q=openai#results

	https://www.例子.公司.cn
	http://中国.中国/中国
	http://中国.中国/foo中国
	http://उदाहरण.परीकषा

	http://192.168.1.1/setup
	http://192.168.1.1:8080/setup
	http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]
	http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080

	https://example.com/products?id=50&category=books
	https://example.com/search?name=John%20Doe&age=30
	http://✪foo.bar/pa✪th©more
	http://foo.com/🐼
	https://shmibbles.me/tmp/自殺でも？.png
	https://subdomain.example.co.uk

	foo://example.com
	file:///C:/path/to/file.txt
	mailto:user@example.com
	ftp://user:password@ftp.example.com:21
	data:text/plain;base64,aGVsbG8=
	javascript:alert('Hello World');
	sms:123
	xmpp:foo@bar
	bitcoin:Addr23?amount=1&message=foo
	cid:foo-32x32.v2_fe0f1423.png
	mid:960830.1639@XIson.com
	postgres://user:pass@host.com:5432/path?k=v#f
	zoommtg://zoom.us/join?confno=1234&pwd=xxx

	https://www.example.com/test space
	https://www.example.com, it's cool!

	/path/to/resource
	www.example.com
	www.example.com/resources
	www.example.com/search?q=openai
	192.168.1.1
	192.168.1.1:8080
	::1
	::ffff:0:0
	64:ff9b::
	64:ff9b:1::
	100::
	2001::
	2001:1::1
	2001:1::2
	2001:2::
	2001:3::
	2001:4:112::
	2001:10::
	2001:20::
	2002::
	2620:4f:8000::
	fc00::
	fe80::
	[2001:0db8:85a3:0000:0000:8a2e:0370:7334]
	[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080
`

// wantWithScheme lists the corpus matches expected when a scheme is required.
var wantWithScheme = []string{
	"http://localhost",
	"http://localhost:8000/home",
	"https://www.example.com",
	"http://www.example.com:8080/page",
	"https://www.example.com/search?q=openai",
	"https://www.example.com/resources/docs/guide",
	"https://www.example.com/about#team",
	"https://user:password@example.com",
	"https://www.example.com:8080/search?q=openai#results",
	"https://www.例子.公司.cn",
	"http://中国.中国/中国",
	"http://中国.中国/foo中国",
	"http://उदाहरण.परीकषा",
	"http://192.168.1.1/setup",
	"http://192.168.1.1:8080/setup",
	"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
	"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080",
	"https://example.com/products?id=50&category=books",
	"https://example.com/search?name=John%20Doe&age=30",
	"http://✪foo.bar/pa✪th©more",
	"http://foo.com/🐼",
	"https://shmibbles.me/tmp/自殺でも？.png",
	"https://subdomain.example.co.uk",
	"foo://example.com",
	"file:///C:/path/to/file.txt",
	"mailto:user@example.com",
	"ftp://user:password@ftp.example.com:21",
	"sms:123",
	"xmpp:foo@bar",
	"bitcoin:Addr23?amount=1&message=foo",
	"cid:foo-32x32.v2_fe0f1423.png",
	"mid:960830.1639@XIson.com",
	"postgres://user:pass@host.com:5432/path?k=v#f",
	"zoommtg://zoom.us/join?confno=1234&pwd=xxx",
	"https://www.example.com/test",
	"https://www.example.com",
}

// wantWithSchemePattern lists the corpus matches expected when the scheme is restricted
// to http and https.
var wantWithSchemePattern = []string{
	"http://localhost",
	"http://localhost:8000/home",
	"https://www.example.com",
	"http://www.example.com:8080/page",
	"https://www.example.com/search?q=openai",
	"https://www.example.com/resources/docs/guide",
	"https://www.example.com/about#team",
	"https://user:password@example.com",
	"https://www.example.com:8080/search?q=openai#results",
	"https://www.例子.公司.cn",
	"http://中国.中国/中国",
	"http://中国.中国/foo中国",
	"http://उदाहरण.परीकषा",
	"http://192.168.1.1/setup",
	"http://192.168.1.1:8080/setup",
	"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
	"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080",
	"https://example.com/products?id=50&category=books",
	"https://example.com/search?name=John%20Doe&age=30",
	"http://✪foo.bar/pa✪th©more",
	"http://foo.com/🐼",
	"https://shmibbles.me/tmp/自殺でも？.png",
	"https://subdomain.example.co.uk",
	"https://www.example.com/test",
	"https://www.example.com",
}

// wantWithHost lists the corpus matches expected when a host is required: everything
// from wantWithScheme, plus scheme-less host URLs and emails.
var wantWithHost = []string{
	"http://localhost",
	"http://localhost:8000/home",
	"https://www.example.com",
	"http://www.example.com:8080/page",
	"https://www.example.com/search?q=openai",
	"https://www.example.com/resources/docs/guide",
	"https://www.example.com/about#team",
	"https://user:password@example.com",
	"https://www.example.com:8080/search?q=openai#results",
	"https://www.例子.公司.cn",
	"http://中国.中国/中国",
	"http://中国.中国/foo中国",
	"http://उदाहरण.परीकषा",
	"http://192.168.1.1/setup",
	"http://192.168.1.1:8080/setup",
	"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
	"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080",
	"https://example.com/products?id=50&category=books",
	"https://example.com/search?name=John%20Doe&age=30",
	"http://✪foo.bar/pa✪th©more",
	"http://foo.com/🐼",
	"https://shmibbles.me/tmp/自殺でも？.png",
	"https://subdomain.example.co.uk",
	"foo://example.com",
	"file:///C:/path/to/file.txt",
	"mailto:user@example.com",
	"ftp://user:password@ftp.example.com:21",
	"sms:123",
	"xmpp:foo@bar",
	"bitcoin:Addr23?amount=1&message=foo",
	"cid:foo-32x32.v2_fe0f1423.png",
	"mid:960830.1639@XIson.com",
	"postgres://user:pass@host.com:5432/path?k=v#f",
	"zoommtg://zoom.us/join?confno=1234&pwd=xxx",
	"https://www.example.com/test",
	"https://www.example.com",
	"www.example.com",
	"www.example.com/resources",
	"www.example.com/search?q=openai",
	"192.168.1.1",
	"192.168.1.1:8080",
	"[2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
	"[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080",
}

// wantWithHostPattern lists the corpus matches expected when the host is restricted to
// example.com and its subdomains.
var wantWithHostPattern = []string{
	"https://www.example.com",
	"http://www.example.com:8080/page",
	"https://www.example.com/search?q=openai",
	"https://www.example.com/resources/docs/guide",
	"https://www.example.com/about#team",
	"https://user:password@example.com",
	"https://www.example.com:8080/search?q=openai#results",
	"https://example.com/products?id=50&category=books",
	"https://example.com/search?name=John%20Doe&age=30",
	"foo://example.com",
	"mailto:user@example.com",
	"ftp://user:password@ftp.example.com:21",
	"https://www.example.com/test",
	"https://www.example.com",
	"www.example.com",
	"www.example.com/resources",
	"www.example.com/search?q=openai",
}

func Test_Extractor_New(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, extractor.New())
}

func Test_Extractor_CompileRegex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []extractor.OptionFunc
		want    []string
	}{
		{
			"URLs with scheme",
			[]extractor.OptionFunc{extractor.WithScheme()},
			wantWithScheme,
		},
		{
			"URLs with scheme pattern",
			[]extractor.OptionFunc{extractor.WithSchemePattern(`(?:https?://)`)},
			wantWithSchemePattern,
		},
		{
			"URLs with host",
			[]extractor.OptionFunc{extractor.WithHost()},
			wantWithHost,
		},
		{
			"URLs with host pattern",
			[]extractor.OptionFunc{extractor.WithHostPattern(`(?:(?:\w+[.])*example\.com` + extractor.ExtractorPortOptionalPattern + `)`)},
			wantWithHostPattern,
		},
		{
			// The Strict preset must behave exactly like WithScheme.
			"Strict preset",
			[]extractor.OptionFunc{extractor.Strict()},
			wantWithScheme,
		},
		{
			// The Relaxed preset must behave exactly like WithHost.
			"Relaxed preset",
			[]extractor.OptionFunc{extractor.Relaxed()},
			wantWithHost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			regex := extractor.New(tt.options...).CompileRegex()

			require.NotNil(t, regex)

			assert.Equal(t, tt.want, regex.FindAllString(corpus, -1))
		})
	}
}

func Test_Extractor_CompileRegex_default(t *testing.T) {
	t.Parallel()

	text := `
		See https://example.com/docs?q=1 or www.example.com, mail admin@example.com,
		ping 192.168.0.1, or fetch /static/logo.png and assets/css/site.css directly.
	`

	want := []string{
		"https://example.com/docs?q=1",
		"www.example.com",
		"admin@example.com",
		"192.168.0.1",
		"/static/logo.png",
		"assets/css/site.css",
	}

	regex := extractor.New().CompileRegex()

	require.NotNil(t, regex)

	assert.Equal(t, want, regex.FindAllString(text, -1))
}

func Test_Extractor_CompileRegex_portBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []extractor.OptionFunc
		text    string
		want    []string
	}{
		{
			"highest valid port matches with host",
			[]extractor.OptionFunc{extractor.WithHost()},
			"example.com:65535/x",
			[]string{"example.com:65535/x"},
		},
		{
			// 65536 is not a valid port: the host matches alone instead of absorbing a
			// truncated port.
			"out-of-range port does not match with host",
			[]extractor.OptionFunc{extractor.WithHost()},
			"example.com:65536/x",
			[]string{"example.com"},
		},
		{
			// With a scheme the authority is followed by a path continuation, which
			// absorbs the out-of-range port either way.
			"out-of-range port is absorbed by the path with scheme",
			[]extractor.OptionFunc{extractor.WithScheme()},
			"http://example.com:65536/y",
			[]string{"http://example.com:65536/y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			regex := extractor.New(tt.options...).CompileRegex()

			require.NotNil(t, regex)

			assert.Equal(t, tt.want, regex.FindAllString(tt.text, -1))
		})
	}
}

func Test_Extractor_CompileRegex_punycode(t *testing.T) {
	t.Parallel()

	text := "Visit https://xn--nxasmq6b.example.com and https://example.xn--p1ai now."

	want := []string{
		"https://xn--nxasmq6b.example.com",
		"https://example.xn--p1ai",
	}

	regex := extractor.New(extractor.WithScheme()).CompileRegex()

	require.NotNil(t, regex)

	assert.Equal(t, want, regex.FindAllString(text, -1))
}
