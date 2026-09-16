// Package domainfilter checks plain comment text against immutable domain rules.
package domainfilter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/idna"
)

const (
	MaxTextBytes      = 400000 // Existing comment limit: 100000 Unicode code points.
	maxCandidateBytes = 4096
	maxRules          = 200000
)

var (
	ErrBlocked    = errors.New("comment contains a blocked domain")
	ErrText       = errors.New("comment text exceeds limits or is invalid UTF-8")
	ErrCandidate  = errors.New("domain-like token exceeds the inspection limit")
	lookup        = idna.New(idna.MapForLookup(), idna.Transitional(false), idna.StrictDomainName(true), idna.ValidateLabels(true), idna.BidiRule(), idna.VerifyDNSLength(true))
	dotForms      = strings.NewReplacer("[.]", ".", "(.)", ".")
	symbolLetters = mappedSymbolLetters()
)

type Rule struct {
	Domain            string
	IncludeSubdomains bool
}

type node struct {
	children       map[string]*node
	exact, subtree bool
}

// Filter owns a reversed-label trie. After construction it is read-only and safe
// for concurrent use without locks, DNS requests, or a per-comment domain cache.
type Filter struct{ root node }

func New(rules []Rule) (*Filter, error) {
	if len(rules) > maxRules {
		return nil, errors.New("too many domain rules")
	}
	f := &Filter{}
	for i, rule := range rules {
		domain, err := canonical(rule.Domain)
		if err != nil {
			return nil, fmt.Errorf("domain rule %d: %w", i+1, err)
		}
		n := &f.root
		for domain != "" {
			pos := strings.LastIndexByte(domain, '.')
			label := domain[pos+1:]
			if n.children == nil {
				n.children = make(map[string]*node)
			}
			if n.children[label] == nil {
				n.children[label] = &node{}
			}
			n = n.children[label]
			if pos < 0 {
				break
			}
			domain = domain[:pos]
		}
		n.exact = true
		n.subtree = n.subtree || rule.IncludeSubdomains
	}
	return f, nil
}

// Load accepts one domain per line (apex and descendants), or =domain for
// exact-only matching. Blank lines and lines starting with # are ignored.
// Invalid configuration fails atomically; no partial rule set is returned.
func Load(r io.Reader) (*Filter, error) {
	const maxFileBytes = 32 << 20
	limited := &io.LimitedReader{R: r, N: maxFileBytes + 1}
	scanner := bufio.NewScanner(limited)
	scanner.Buffer(make([]byte, 4096), 8192)
	var rules []Rule
	total := 0
	for line := 1; scanner.Scan(); line++ {
		total += len(scanner.Bytes()) + 1
		if total > maxFileBytes {
			return nil, errors.New("domain rule file exceeds 32 MiB")
		}
		value := strings.TrimSpace(scanner.Text())
		if value == "" || strings.HasPrefix(value, "#") {
			continue
		}
		rule := Rule{Domain: value, IncludeSubdomains: true}
		if strings.HasPrefix(value, "=") {
			rule.Domain = value[1:]
			rule.IncludeSubdomains = false
		}
		if _, err := canonical(rule.Domain); err != nil {
			return nil, fmt.Errorf("domain rule line %d: %w", line, err)
		}
		rules = append(rules, rule)
		if len(rules) > maxRules {
			return nil, errors.New("too many domain rules")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read domain rules: %w", err)
	}
	if limited.N == 0 {
		return nil, errors.New("domain rule file exceeds 32 MiB")
	}
	return New(rules)
}

func canonical(s string) (string, error) {
	if len(s) > maxCandidateBytes || !utf8.ValidString(s) {
		return "", ErrCandidate
	}
	// Most candidates are ASCII. Avoid invoking IDNA normalization on that path,
	// except A-labels, whose punycode must also be validated.
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			ascii = false
			break
		}
	}
	if ascii {
		s = strings.ToLower(s)
	}
	if !ascii || strings.HasPrefix(s, "xn--") || strings.Contains(s, ".xn--") {
		var err error
		s, err = lookup.ToASCII(s)
		if err != nil {
			return "", err
		}
	}
	s = strings.TrimSuffix(s, ".")
	if len(s) > 253 || !strings.Contains(s, ".") {
		return "", errors.New("expected a multi-label DNS domain")
	}
	for rest := s; ; {
		label, next, more := strings.Cut(rest, ".")
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", errors.New("invalid DNS label")
		}
		if len(label) >= 4 && label[2:4] == "--" && !strings.HasPrefix(label, "xn--") {
			return "", errors.New("reserved DNS label")
		}
		for i := range len(label) {
			c := label[i]
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return "", errors.New("invalid DNS character")
			}
		}
		if !more {
			// Numeric dotted strings are not domain rules/candidates.
			letters := false
			for i := range len(label) {
				letters = letters || label[i] >= 'a' && label[i] <= 'z'
			}
			if !letters {
				return "", errors.New("numeric top-level label")
			}
			break
		}
		rest = next
	}
	return s, nil
}

func (f *Filter) matches(domain string) bool {
	n := &f.root
	for {
		pos := strings.LastIndexByte(domain, '.')
		n = n.children[domain[pos+1:]]
		if n == nil {
			return false
		}
		if n.subtree {
			return true
		}
		if pos < 0 {
			return n.exact
		}
		domain = domain[:pos]
	}
}

// Check scans maximal domain-like tokens, never arbitrary substrings within a
// hostname. Runtime is O(text bytes) with bounded IDNA inputs and trie walks;
// working memory is bounded by MaxTextBytes, independent of blacklist size.
func (f *Filter) Check(text string) error {
	if len(text) > MaxTextBytes || !utf8.ValidString(text) {
		return ErrText
	}
	if f == nil || len(f.root.children) == 0 {
		return nil
	}
	text = normalizeText(text)
	var token [maxCandidateBytes]byte
	size, length := 0, 0
	dotted := false
	flush := func() error {
		defer func() { size, length, dotted = 0, 0, false }()
		if !dotted {
			return nil
		}
		if length > len(token) {
			return ErrCandidate
		}
		domain, err := canonical(string(token[:size]))
		if err == nil && f.matches(domain) {
			return ErrBlocked
		}
		return nil
	}
	for i := 0; i < len(text); {
		r, width := utf8.DecodeRuneInString(text[i:])
		if domainRune(r) || r == '.' {
			length += width
			if length <= len(token) {
				size += copy(token[size:], text[i:i+width])
			}
			dotted = dotted || r == '.'
			i += width
			continue
		}
		// Only whitespace next to a dot is joined, never spaces inside labels.
		if unicode.IsSpace(r) {
			j := i + width
			for j < len(text) {
				r2, w2 := utf8.DecodeRuneInString(text[j:])
				if !unicode.IsSpace(r2) {
					break
				}
				j += w2
			}
			if length > 0 && (size > 0 && token[size-1] == '.' || j < len(text) && text[j] == '.') {
				i = j
				continue
			}
			i = j
		} else {
			i += width
		}
		if err := flush(); err != nil {
			return err
		}
	}
	return flush()
}

func domainRune(r rune) bool {
	// Underscores and format controls deliberately stay inside the token: they
	// must not expose a valid suffix of an invalid or unrelated hostname.
	if r < 128 {
		return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_'
	}
	return r == '\uff0d' || r == '\ufe63' || unicode.Is(unicode.Pc, r) || unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r) || unicode.Is(unicode.Cf, r) || symbolLetters[r]
}

// Some lookup-valid characters (e.g. circled letters and ™) are classified as
// Unicode symbols, not letters. Compute their membership once at startup so
// emoji-heavy requests do not trigger one IDNA conversion per character.
func mappedSymbolLetters() map[rune]bool {
	result := make(map[rune]bool)
	add := func(r rune) {
		mapped, err := lookup.ToUnicode(string(r))
		if err != nil || mapped == "" {
			return
		}
		for _, c := range mapped {
			if !(unicode.IsLetter(c) || unicode.IsNumber(c) || unicode.IsMark(c) || c == '-') {
				return
			}
		}
		result[r] = true
	}
	for _, group := range unicode.S.R16 {
		for r := uint32(group.Lo); r <= uint32(group.Hi); r += uint32(group.Stride) {
			add(rune(r))
		}
	}
	for _, group := range unicode.S.R32 {
		for r := group.Lo; r <= group.Hi; r += group.Stride {
			add(rune(r))
		}
	}
	return result
}

func normalizeText(text string) string {
	if !strings.ContainsAny(text, "%[(。．｡\u200b\ufeff\u00ad") {
		return text
	}
	// Decode percent escapes once, including UTF-8 host bytes. Invalid escapes
	// stay literal; one invalid escape cannot disable inspection of other URLs.
	var decoded strings.Builder
	decoded.Grow(len(text))
	for i := 0; i < len(text); i++ {
		if text[i] == '%' && i+2 < len(text) {
			a, b := hex(text[i+1]), hex(text[i+2])
			if a >= 0 && b >= 0 {
				decoded.WriteByte(byte(a*16 + b))
				i += 2
				continue
			}
		}
		decoded.WriteByte(text[i])
	}
	var normalized strings.Builder
	normalized.Grow(decoded.Len())
	for _, r := range decoded.String() {
		switch r {
		case '\u3002', '\uff0e', '\uff61':
			normalized.WriteByte('.')
		case '\u200b', '\ufeff', '\u00ad': // Common invisible lookup-ignored obfuscation.
		default:
			normalized.WriteRune(r)
		}
	}
	return dotForms.Replace(normalized.String())
}

func hex(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	default:
		return -1
	}
}
