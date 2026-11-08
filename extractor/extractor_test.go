package extractor_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgourlextractor "github.com/hueristiq/hq-lib-url-go/extractor"
)

func mustNew(tb testing.TB, ofs ...hqgourlextractor.OptionFunc) (extractor *hqgourlextractor.Extractor) {
	tb.Helper()

	extractor, err := hqgourlextractor.New(ofs...)

	require.NoError(tb, err)

	return
}

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
	"data:text/plain;base64,aGVsbG8=",
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
	"data:text/plain;base64,aGVsbG8=",
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

	assert.NotNil(t, mustNew(t))
}

func Test_Extractor_Regexp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []hqgourlextractor.OptionFunc
		want    []string
	}{
		{
			"URLs with scheme",
			[]hqgourlextractor.OptionFunc{hqgourlextractor.WithScheme()},
			wantWithScheme,
		},
		{
			"URLs with scheme pattern",
			[]hqgourlextractor.OptionFunc{hqgourlextractor.WithSchemePattern(`(?:https?://)`)},
			wantWithSchemePattern,
		},
		{
			"URLs with host",
			[]hqgourlextractor.OptionFunc{hqgourlextractor.WithHost()},
			wantWithHost,
		},
		{
			"URLs with host pattern",
			[]hqgourlextractor.OptionFunc{hqgourlextractor.WithHostPattern(`(?:(?:\w+[.])*example\.com` + hqgourlextractor.PortOptionalPattern + `)`)},
			wantWithHostPattern,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			regex := mustNew(t, tt.options...).Regexp()

			require.NotNil(t, regex)

			assert.Equal(t, tt.want, regex.FindAllString(corpus, -1))
		})
	}
}

func Test_Extractor_Regexp_withoutEmail(t *testing.T) {
	t.Parallel()

	text := `
		See https://example.com/docs?q=1 or mail admin@example.com.
	`

	// withoutEmail drops the bare-email alternation; the match survives via
	// the user info + host URL form. Combine with WithoutUserInfo to exclude
	// both.
	want := []string{
		"https://example.com/docs?q=1",
		"admin@example.com",
	}

	regex := mustNew(t, hqgourlextractor.WithoutEmail()).Regexp()

	require.NotNil(t, regex)

	assert.Equal(t, want, regex.FindAllString(text, -1))
}

func Test_Extractor_Regexp_withoutUserInfo(t *testing.T) {
	t.Parallel()

	hostPattern := `(?:(?:\w+[.])*example\.com` + hqgourlextractor.PortOptionalPattern + `)`

	//nolint:gosec // G101: the credentials are fake, exercising user info matching.
	text := `
		See https://user:pass@example.com/private or https://example.com/public.
	`

	// withoutUserInfo strips user info from the authority, so the scheme URL
	// loses its prefix; the bare-email alternation then swallows the host
	// ("pass@example.com"), which is why callers wanting neither pair it
	// with WithoutEmail.
	want := []string{
		"pass@example.com",
		"https://example.com/public",
	}

	regex := mustNew(t,
		hqgourlextractor.WithHostPattern(hostPattern),
		hqgourlextractor.WithoutUserInfo(),
	).Regexp()

	require.NotNil(t, regex)

	assert.Equal(t, want, regex.FindAllString(text, -1))
}

func Test_Extractor_Regexp_withoutEmailAndUserInfo(t *testing.T) {
	t.Parallel()

	hostPattern := `(?:(?:\w+[.])*example\.com` + hqgourlextractor.PortOptionalPattern + `)`

	text := `
		See https://'jane@example.com/private or https://example.com/public,
		mail admin@example.com.
	`

	want := []string{
		"example.com/private",
		"https://example.com/public",
		"example.com",
	}

	regex := mustNew(t,
		hqgourlextractor.WithHostPattern(hostPattern),
		hqgourlextractor.WithoutEmail(),
		hqgourlextractor.WithoutUserInfo(),
	).Regexp()

	require.NotNil(t, regex)

	assert.Equal(t, want, regex.FindAllString(text, -1))
}

func Test_Extractor_Regexp_withoutUserInfo_host(t *testing.T) {
	t.Parallel()

	//nolint:gosec // G101: the credentials are fake, exercising user info matching.
	text := `
		See https://user:pass@example.com/private or mailto:user@example.com.
	`

	// With the host required and user info excluded, the authority becomes
	// mandatory, so the scheme URL drops its userinfo prefix; no-authority
	// schemes are unaffected.
	want := []string{
		"pass@example.com",
		"mailto:user@example.com",
	}

	regex := mustNew(t,
		hqgourlextractor.WithHost(),
		hqgourlextractor.WithoutUserInfo(),
	).Regexp()

	require.NotNil(t, regex)

	assert.Equal(t, want, regex.FindAllString(text, -1))
}

func Test_Extractor_Regexp_default(t *testing.T) {
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

	regex := mustNew(t).Regexp()

	require.NotNil(t, regex)

	assert.Equal(t, want, regex.FindAllString(text, -1))
}

func Test_Extractor_Regexp_portBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []hqgourlextractor.OptionFunc
		text    string
		want    []string
	}{
		{
			"highest valid port matches with host",
			[]hqgourlextractor.OptionFunc{hqgourlextractor.WithHost()},
			"example.com:65535/x",
			[]string{"example.com:65535/x"},
		},
		{
			// 65536 is not a valid port: the host matches alone instead of absorbing a
			// truncated port.
			"out-of-range port does not match with host",
			[]hqgourlextractor.OptionFunc{hqgourlextractor.WithHost()},
			"example.com:65536/x",
			[]string{"example.com"},
		},
		{
			// With a scheme the authority is followed by a path continuation, which
			// absorbs the out-of-range port either way.
			"out-of-range port is absorbed by the path with scheme",
			[]hqgourlextractor.OptionFunc{hqgourlextractor.WithScheme()},
			"http://example.com:65536/y",
			[]string{"http://example.com:65536/y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			regex := mustNew(t, tt.options...).Regexp()

			require.NotNil(t, regex)

			assert.Equal(t, tt.want, regex.FindAllString(tt.text, -1))
		})
	}
}

func Test_Extractor_Regexp_punycode(t *testing.T) {
	t.Parallel()

	text := "Visit https://xn--nxasmq6b.example.com and https://example.xn--p1ai now."

	want := []string{
		"https://xn--nxasmq6b.example.com",
		"https://example.xn--p1ai",
	}

	regex := mustNew(t, hqgourlextractor.WithScheme()).Regexp()

	require.NotNil(t, regex)

	assert.Equal(t, want, regex.FindAllString(text, -1))
}

func TestNew_InvalidConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []hqgourlextractor.OptionFunc
	}{
		{
			"invalid scheme pattern",
			[]hqgourlextractor.OptionFunc{hqgourlextractor.WithSchemePattern(`(`)},
		},
		{
			"invalid host pattern",
			[]hqgourlextractor.OptionFunc{hqgourlextractor.WithHostPattern(`[`)},
		},
		{
			"custom pattern with ignored options",
			[]hqgourlextractor.OptionFunc{hqgourlextractor.WithRegexp(regexp.MustCompile(`https://\S+`)), hqgourlextractor.WithHost()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			extractor, err := hqgourlextractor.New(tt.options...)

			require.Error(t, err)
			assert.Nil(t, extractor)
			assert.ErrorContains(t, err, "extractor.New")
		})
	}
}

func Test_Extractor_Regexp_schemeWithDigits(t *testing.T) {
	t.Parallel()

	regex := mustNew(t, hqgourlextractor.WithScheme()).Regexp()

	require.NotNil(t, regex)

	assert.Equal(t, []string{"h323://example.com/call"}, regex.FindAllString("h323://example.com/call", -1))
}

func Test_Extractor_Regexp_schemeCaseInsensitive(t *testing.T) {
	t.Parallel()

	regex := mustNew(t, hqgourlextractor.WithScheme()).Regexp()

	require.NotNil(t, regex)

	for _, text := range []string{
		"MAILTO:user@example.com",
		"TEL:+123",
		"SMS:123",
		"HTTP://example.com/path",
	} {
		assert.Equal(t, []string{text}, regex.FindAllString(text, -1))
	}
}

func Test_Extractor_Regexp_hostQueryFragment(t *testing.T) {
	t.Parallel()

	regex := mustNew(t, hqgourlextractor.WithHost()).Regexp()

	require.NotNil(t, regex)

	tests := []struct {
		name string
		text string
		want []string
	}{
		{"query without path", "www.example.com?q=1", []string{"www.example.com?q=1"}},
		{"fragment without path", "www.example.com#frag", []string{"www.example.com#frag"}},
		{"port and query", "192.168.1.1:8080?x=1", []string{"192.168.1.1:8080?x=1"}},
		{"trailing dot", "www.example.com.", []string{"www.example.com"}},
		{"path and query", "www.example.com/search?q=openai", []string{"www.example.com/search?q=openai"}},
		{"empty query", "www.example.com?", []string{"www.example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, regex.FindAllString(tt.text, -1))
		})
	}
}

func Test_Extractor_Regexp_customPatternGrouping(t *testing.T) {
	t.Parallel()

	t.Run("scheme alternation", func(t *testing.T) {
		t.Parallel()

		regex := mustNew(t, hqgourlextractor.WithSchemePattern(`http://|https://`)).Regexp()

		require.NotNil(t, regex)

		text := "http://example.com https://example.com/x"

		assert.Equal(t, []string{"http://example.com", "https://example.com/x"}, regex.FindAllString(text, -1))
	})

	t.Run("host alternation in email", func(t *testing.T) {
		t.Parallel()

		regex := mustNew(t,
			hqgourlextractor.WithHostPattern(`example\.com|example\.org`),
			hqgourlextractor.WithoutUserInfo(),
		).Regexp()

		require.NotNil(t, regex)

		assert.Equal(t, []string{"admin@example.org"}, regex.FindAllString("admin@example.org", -1))
	})

	t.Run("host alternation in URL", func(t *testing.T) {
		t.Parallel()

		regex := mustNew(t, hqgourlextractor.WithHostPattern(`example\.com|example\.org`)).Regexp()

		require.NotNil(t, regex)

		assert.Equal(t, []string{"example.org/x"}, regex.FindAllString("example.org/x", -1))
	})
}

func Test_Extractor_Regexp_dataScheme(t *testing.T) {
	t.Parallel()

	const dataURL = "data:text/plain;base64,aGVsbG8="

	tests := []struct {
		name    string
		options []hqgourlextractor.OptionFunc
	}{
		{"with scheme", []hqgourlextractor.OptionFunc{hqgourlextractor.WithScheme()}},
		{"with host", []hqgourlextractor.OptionFunc{hqgourlextractor.WithHost()}},
		{"default", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			regex := mustNew(t, tt.options...).Regexp()

			require.NotNil(t, regex)

			assert.Equal(t, []string{dataURL}, regex.FindAllString(dataURL, -1))
		})
	}
}

func Test_Extractor_New_IgnoresNilOptions(t *testing.T) {
	t.Parallel()

	e := mustNew(t, nil)
	require.NotNil(t, e)
}

func ExampleExtractor_Extract() {
	e, err := hqgourlextractor.New()
	if err != nil {
		fmt.Println(err)

		return
	}

	for _, match := range e.Extract("Visit https://example.com and https://scanme.sh!") {
		fmt.Println(match)
	}
	// Output:
	// https://example.com
	// https://scanme.sh
}
