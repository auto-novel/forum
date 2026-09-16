package domainfilter

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func testFilter(t testing.TB) *Filter {
	t.Helper()
	f, err := New([]Rule{{"evil.com", true}, {"only.example", false}, {"bücher.de", true}, {"例子.中国", true}, {"faß.de", true}})
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestCheck(t *testing.T) {
	f := testFilter(t)
	for _, content := range []string{
		"https://evil.com/a", "evil.com", "www.evil.com", "a.b.evil.com",
		"evil[.]com", "evil(.)com", "evil。com", "evil．com", "evil｡com", "evil . com",
		"EVIL.CoM", "ＥＶＩＬ.ＣＯＭ", "ⓔⓥⓘⓛ.ⓒⓞⓜ", "evil\t.\ncom", "evil\u00a0.\u3000com", "😀evil.com😀",
		"前文：evil.com，后文", "[链接](https://evil.com/a)", "<evil.com>", "evil.com:443/a",
		"evil.com.", "https://evil.com./", "user@evil.com", "https://user:pass@evil.com/a",
		"only.example", "bücher.de", "BU\u0308CHER.de", "xn--bcher-kva.de", "a.例子.中国",
		"xn--fsqu00a.xn--fiqs8s", "faß.de", "xn--fa-hia.de",
		"e\u200bvil.com", "evil\ufeff.com", "e\u00advil.com", "evil[。]com",
		"sub－a.evil.com", "sub﹣a.evil.com",
		"https://%65vil%2ecom/", "%zz https://evil%2Ecom/", "https://b%C3%BCcher.de/",
		strings.Repeat("safe.example ", 7000) + "evil.com",
	} {
		t.Run(content[:min(len(content), 90)], func(t *testing.T) {
			if err := f.Check(content); !errors.Is(err, ErrBlocked) {
				t.Fatalf("expected blocked, got %v", err)
			}
		})
	}
	for _, content := range []string{
		"", "普通正文，没有链接", "hello world", "notevil.com", "evil.com.example.org",
		"a.evil.com.example.org", "www.only.example", "evil . com . example.org",
		"notｅｖｉｌ.com", "evil.com。example.org", "evil_com", "evil..com", "evil.com_foo",
		"_evil.com", "-evil.com", "evil-.com", "a..evil.com", "xx--x.evil.com",
		"e vil.com", "版本 1.2.3", "127.0.0.1", "fass.de", "еvil.com", // Cyrillic е is a distinct domain.
		"evil.com-example.org", "evil.com－example.org", "evil.com﹣example.org", "evil.com＿example.org", "https://evil.com.example.org/path", "evil%2Ecom.example.org",
		strings.Repeat("a", 64) + ".evil.com", strings.Repeat("正常正文", 20000),
	} {
		t.Run("allow/"+content[:min(len(content), 90)], func(t *testing.T) {
			if err := f.Check(content); err != nil {
				t.Fatalf("expected allowed, got %v", err)
			}
		})
	}
}

func TestLimits(t *testing.T) {
	f := testFilter(t)
	for _, tc := range []struct {
		content string
		want    error
	}{
		{strings.Repeat("a", MaxTextBytes+1), ErrText},
		{string([]byte{0xff}), ErrText},
		{strings.Repeat("a", maxCandidateBytes) + ".evil.com", ErrCandidate},
		{"evil.com." + strings.Repeat("x", maxCandidateBytes), ErrCandidate},
		{strings.Repeat("a.", 5000), ErrCandidate},
		{strings.Repeat("a", 10000) + " evil.com", ErrBlocked},
	} {
		if err := f.Check(tc.content); !errors.Is(err, tc.want) {
			t.Fatalf("wanted %v, got %v", tc.want, err)
		}
	}
	var disabled *Filter
	if err := disabled.Check("evil.com"); err != nil {
		t.Fatal(err)
	}
}

func TestRules(t *testing.T) {
	f, err := Load(strings.NewReader("# comment\nEVIL.COM.\n=only.example\n=evil.com\n例子。中国\n\n"))
	if err != nil {
		t.Fatal(err)
	}
	if f.Check("a.evil.com") != ErrBlocked || f.Check("a.only.example") != nil || f.Check("only.example") != ErrBlocked {
		t.Fatal("incorrect rule modes or duplicate merge")
	}
	for _, rule := range []string{"com", "https://evil.com", "evil.com/path", "*.evil.com", "evil[.]com", "evil . com", "evil.com:80", "evil.com..", ".evil.com", "a..com", "a_b.com", "-a.com", "127.0.0.1", "xn--.com", "xn--abc.com", "a\u200d.com", "a\u202e.com", strings.Repeat("a", 64) + ".com", strings.Repeat("a.", 127) + "com"} {
		if _, err := New([]Rule{{rule, true}}); err == nil {
			t.Errorf("accepted invalid rule %q", rule)
		}
	}
	if f, err := Load(strings.NewReader("evil.com\nhttps://bad.example\n")); err == nil || f != nil || !strings.Contains(err.Error(), "line 2") {
		t.Fatal("invalid file must fail atomically with a line number")
	}
	if _, err := Load(strings.NewReader(strings.Repeat("x", 9000))); err == nil {
		t.Fatal("accepted oversized line")
	}
}

func TestUnicodeMappingPreservesDomainIdentity(t *testing.T) {
	f, err := New([]Rule{{"i.example", true}})
	if err != nil {
		t.Fatal(err)
	}
	// Unicode simple lowercase would incorrectly turn U+0130 into ASCII i;
	// UTS #46 maps it to i followed by a combining dot instead.
	if err := f.Check("İ.example"); err != nil {
		t.Fatalf("distinct IDN was conflated: %v", err)
	}
	if err := f.Check("Ｉ.example"); err != ErrBlocked {
		t.Fatalf("width mapping was missed: %v", err)
	}
}

func TestConcurrentCheck(t *testing.T) {
	f := testFilter(t)
	var wg sync.WaitGroup
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				if f.Check("https://EVIL[.]com/a") != ErrBlocked || f.Check("evil.com.example.org") != nil {
					t.Error("incorrect concurrent result")
				}
			}
		}()
	}
	wg.Wait()
}

func TestLargeRuleSet(t *testing.T) {
	rules := make([]Rule, 100000)
	for i := range rules {
		rules[i] = Rule{fmt.Sprintf("blocked%d.example", i), true}
	}
	f, err := New(rules)
	if err != nil {
		t.Fatal(err)
	}
	for _, domain := range []string{"blocked0.example", "a.b.blocked99999.example", "blocked54321.example"} {
		if f.Check(domain) != ErrBlocked {
			t.Fatalf("missed %s", domain)
		}
	}
	if f.Check("notblocked99999.example blocked99999.example.org") != nil {
		t.Fatal("false positive in large rule set")
	}
}

func TestLoadLimits(t *testing.T) {
	for name, input := range map[string]string{
		"rules":      strings.Repeat("evil.com\n", maxRules+1),
		"bytes_CRLF": strings.Repeat("#"+strings.Repeat("x", 4000)+"\r\n", 8400),
	} {
		t.Run(name, func(t *testing.T) {
			if f, err := Load(strings.NewReader(input)); err == nil || f != nil {
				t.Fatal("oversized rule file did not fail atomically")
			}
		})
	}
}

func FuzzCheck(f *testing.F) {
	filter := testFilter(f)
	for _, seed := range []string{"evil.com", "evil[.]com", "evil.com.example.org", "例子。中国", "\xff", "a\u200d.com", "%E3%80%82", strings.Repeat("a.", 3000)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		err := filter.Check(input)
		if err != nil && !errors.Is(err, ErrBlocked) && !errors.Is(err, ErrText) && !errors.Is(err, ErrCandidate) {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(input) < 100000 {
			// Long or malformed preceding candidates must never hide a later hit.
			if filter.Check(input+" / evil.com") == nil {
				t.Fatal("missed independent blocked domain")
			}
		}
	})
}

func BenchmarkCheck(b *testing.B) {
	for _, count := range []int{1, 100000} {
		rules := make([]Rule, count)
		for i := range rules {
			rules[i] = Rule{fmt.Sprintf("blocked%d.example", i), true}
		}
		filter, err := New(rules)
		if err != nil {
			b.Fatal(err)
		}
		for name, input := range map[string]string{
			"short":      "一条正常评论 https://safe.example/hello",
			"many":       strings.Repeat("safe.example ", 7000),
			"unicode":    strings.Repeat("正常.中国 ", 16000),
			"late_hit":   strings.Repeat("safe.example ", 7000) + "a.blocked0.example",
			"long_token": strings.Repeat("a", 99900) + ".example",
			"dots":       strings.Repeat("a.", 49000),
		} {
			b.Run(fmt.Sprintf("%d/%s", count, name), func(b *testing.B) {
				b.SetBytes(int64(len(input)))
				b.ReportAllocs()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						_ = filter.Check(input)
					}
				})
			})
		}
	}
}
