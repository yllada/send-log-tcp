package main

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// =============================================================================
// SYSLOG MESSAGE FORMAT TESTS
// Covers the raw-pri passthrough mode, the message format resolution used for
// backward compatibility with templates saved before MessageFormat existed, and
// regression pins for the pre-existing RFC 5424 / RFC 3164 builders.
// =============================================================================

// rfc5424TimestampPattern matches the timestamp produced by buildRFC5424Message
// ("2006-01-02T15:04:05.000Z07:00" rendered in UTC).
var rfc5424TimestampPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`)

// rfc3164Pattern matches the full RFC 3164 line ("Jan  2 15:04:05" keeps a
// padded day, so the day separator may be one or two spaces).
var rfc3164Pattern = regexp.MustCompile(`^<134>[A-Z][a-z]{2} {1,2}\d{1,2} \d{2}:\d{2}:\d{2} myhost myapp: hello world$`)

// =============================================================================
// RAW-PRI FORMAT TESTS
// =============================================================================

func TestBuildSyslogMessage_RawPRIPrependsPriorityOnly(t *testing.T) {
	// Real ESET ESMC / ERAServer line: the vendor payload already carries the
	// full RFC 5424 header, so the app must only supply the PRI.
	const esetMessage = `1 2022-04-11T10:41:05.300Z esmc ERAServer 11804 - - {"event_type":"FirewallAggregated_Event"}`
	const want = `<12>1 2022-04-11T10:41:05.300Z esmc ERAServer 11804 - - {"event_type":"FirewallAggregated_Event"}`

	config := SyslogConfig{
		Facility:      1, // user
		Severity:      4, // warning
		MessageFormat: MessageFormatRawPRI,
	}

	got, err := buildSyslogMessage(config, esetMessage)
	if err != nil {
		t.Fatalf("buildSyslogMessage returned unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("raw-pri output mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestBuildSyslogMessage_RawPRINoSpaceAfterPriority(t *testing.T) {
	config := SyslogConfig{
		Facility:      16,
		Severity:      6,
		MessageFormat: MessageFormatRawPRI,
	}

	got, err := buildSyslogMessage(config, "payload")
	if err != nil {
		t.Fatalf("buildSyslogMessage returned unexpected error: %v", err)
	}

	const prefix = "<134>"
	if !strings.HasPrefix(got, prefix) {
		t.Fatalf("expected output to start with %q, got %q", prefix, got)
	}

	remainder := strings.TrimPrefix(got, prefix)
	if remainder != "payload" {
		t.Errorf("expected message bytes immediately after '>', got %q", remainder)
	}
	if len(remainder) > 0 && remainder[0] == ' ' {
		t.Errorf("raw-pri must not insert a space after '>', got %q", got)
	}
}

func TestBuildSyslogMessage_RawPRIPreservesBytesVerbatim(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "json with backslash escapes",
			message: `{"path":"C:\\Windows\\System32","quote":"he said \"hi\""}`,
		},
		{
			name:    "non ascii multibyte",
			message: `usuario conectó desde 東京 — ok ✓`,
		},
		{
			name:    "leading and trailing spaces",
			message: "   padded payload   ",
		},
		{
			name:    "embedded line feed",
			message: "first line\nsecond line",
		},
		{
			name:    "embedded carriage return is not stripped",
			message: "first line\r\nsecond line",
		},
		{
			name:    "tab characters",
			message: "col1\tcol2\tcol3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := SyslogConfig{
				Facility:      1,
				Severity:      4,
				MessageFormat: MessageFormatRawPRI,
			}

			got, err := buildSyslogMessage(config, tt.message)
			if err != nil {
				t.Fatalf("buildSyslogMessage returned unexpected error: %v", err)
			}

			want := "<12>" + tt.message
			if got != want {
				t.Errorf("bytes not preserved verbatim\n got: %q\nwant: %q", got, want)
			}
			if len(got) != len("<12>")+len(tt.message) {
				t.Errorf("byte length changed: got %d bytes, want %d", len(got), len("<12>")+len(tt.message))
			}
		})
	}
}

func TestBuildSyslogMessage_RawPRIIgnoresHostnameAndAppname(t *testing.T) {
	const message = `1 2022-04-11T10:41:05.300Z esmc ERAServer 11804 - - {"k":"v"}`

	config := SyslogConfig{
		Facility:      1,
		Severity:      4,
		Hostname:      "JUNK-HOSTNAME",
		Appname:       "JUNK-APPNAME",
		MessageFormat: MessageFormatRawPRI,
	}

	got, err := buildSyslogMessage(config, message)
	if err != nil {
		t.Fatalf("buildSyslogMessage returned unexpected error: %v", err)
	}

	if want := "<12>" + message; got != want {
		t.Errorf("raw-pri output mismatch\n got: %q\nwant: %q", got, want)
	}
	if strings.Contains(got, "JUNK-HOSTNAME") {
		t.Error("raw-pri must not inject Hostname")
	}
	if strings.Contains(got, "JUNK-APPNAME") {
		t.Error("raw-pri must not inject Appname")
	}
}

func TestBuildRawPRIMessage_PriorityAndMessageOnly(t *testing.T) {
	tests := []struct {
		name     string
		priority uint8
		message  string
		want     string
	}{
		{name: "lowest priority", priority: 0, message: "payload", want: "<0>payload"},
		{name: "eset priority", priority: 12, message: "1 x", want: "<12>1 x"},
		{name: "highest valid priority", priority: 191, message: "payload", want: "<191>payload"},
		{name: "empty message keeps only pri", priority: 12, message: "", want: "<12>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildRawPRIMessage(tt.priority, tt.message); got != tt.want {
				t.Errorf("buildRawPRIMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// PRIORITY BOUNDARY TESTS
// =============================================================================

func TestBuildSyslogMessage_PriorityBoundary(t *testing.T) {
	tests := []struct {
		name     string
		facility uint8
		severity uint8
		wantErr  bool
		want     string
	}{
		{
			name:     "facility 23 severity 7 is the maximum valid priority",
			facility: 23,
			severity: 7,
			wantErr:  false,
			want:     "<191>payload",
		},
		{
			name:     "facility 23 severity 0 is valid",
			facility: 23,
			severity: 0,
			wantErr:  false,
			want:     "<184>payload",
		},
		{
			name:     "facility 24 exceeds the maximum priority",
			facility: 24,
			severity: 0,
			wantErr:  true,
		},
		{
			name:     "facility 24 severity 7 exceeds the maximum priority",
			facility: 24,
			severity: 7,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := SyslogConfig{
				Facility:      tt.facility,
				Severity:      tt.severity,
				MessageFormat: MessageFormatRawPRI,
			}

			got, err := buildSyslogMessage(config, "payload")

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got message %q", got)
				}
				if !strings.Contains(err.Error(), "invalid priority") {
					t.Errorf("expected an 'invalid priority' error, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// MESSAGE FORMAT RESOLUTION TESTS (backward compatibility)
// =============================================================================

func TestResolveMessageFormat_ExplicitValueWins(t *testing.T) {
	tests := []struct {
		name       string
		format     MessageFormat
		useRFC5424 bool
		want       MessageFormat
	}{
		{
			name:       "explicit raw-pri wins over UseRFC5424 true",
			format:     MessageFormatRawPRI,
			useRFC5424: true,
			want:       MessageFormatRawPRI,
		},
		{
			name:       "explicit raw-pri wins over UseRFC5424 false",
			format:     MessageFormatRawPRI,
			useRFC5424: false,
			want:       MessageFormatRawPRI,
		},
		{
			name:       "explicit rfc3164 wins over UseRFC5424 true",
			format:     MessageFormatRFC3164,
			useRFC5424: true,
			want:       MessageFormatRFC3164,
		},
		{
			name:       "explicit rfc5424 wins over UseRFC5424 false",
			format:     MessageFormatRFC5424,
			useRFC5424: false,
			want:       MessageFormatRFC5424,
		},
		{
			name:       "empty format falls back to UseRFC5424 true",
			format:     "",
			useRFC5424: true,
			want:       MessageFormatRFC5424,
		},
		{
			name:       "empty format falls back to UseRFC5424 false",
			format:     "",
			useRFC5424: false,
			want:       MessageFormatRFC3164,
		},
		{
			name:       "unrecognized format falls back to UseRFC5424 true",
			format:     "totally-unknown",
			useRFC5424: true,
			want:       MessageFormatRFC5424,
		},
		{
			name:       "unrecognized format falls back to UseRFC5424 false",
			format:     "totally-unknown",
			useRFC5424: false,
			want:       MessageFormatRFC3164,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveMessageFormat(tt.format, tt.useRFC5424); got != tt.want {
				t.Errorf("resolveMessageFormat(%q, %v) = %q, want %q", tt.format, tt.useRFC5424, got, tt.want)
			}
		})
	}
}

func TestBuildSyslogMessage_LegacyUseRFC5424FallbackSelectsFormat(t *testing.T) {
	tests := []struct {
		name       string
		useRFC5424 bool
		wantPrefix string
	}{
		{
			name:       "legacy template with UseRFC5424 true still emits RFC 5424",
			useRFC5424: true,
			wantPrefix: "<134>1 ",
		},
		{
			name:       "legacy template with UseRFC5424 false still emits RFC 3164",
			useRFC5424: false,
			wantPrefix: "<134>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := SyslogConfig{
				Facility:   16,
				Severity:   6,
				Hostname:   "myhost",
				Appname:    "myapp",
				UseRFC5424: tt.useRFC5424,
				// MessageFormat intentionally left empty: this is a template
				// stored before the MessageFormat field existed.
			}

			got, err := buildSyslogMessage(config, "hello world")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("expected prefix %q, got %q", tt.wantPrefix, got)
			}
			if tt.useRFC5424 && strings.HasPrefix(got, "<134>1 ") == false {
				t.Errorf("expected the RFC 5424 version digit, got %q", got)
			}
			if !tt.useRFC5424 && strings.HasPrefix(got, "<134>1 ") {
				t.Errorf("RFC 3164 must not emit a version digit, got %q", got)
			}
		})
	}
}

// =============================================================================
// REGRESSION PINS FOR THE EXISTING FORMATS
// These builders embed time.Now(), so assertions target structure instead of
// exact string equality to stay deterministic.
// =============================================================================

func TestBuildSyslogMessage_RFC5424StructureUnchanged(t *testing.T) {
	config := SyslogConfig{
		Facility:      16,
		Severity:      6,
		Hostname:      "myhost",
		Appname:       "myapp",
		MessageFormat: MessageFormatRFC5424,
	}

	got, err := buildSyslogMessage(config, "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// <PRI>VERSION SP TIMESTAMP SP HOSTNAME SP APP-NAME SP PROCID SP MSGID SP STRUCTURED-DATA SP MSG
	fields := strings.SplitN(got, " ", 8)
	if len(fields) != 8 {
		t.Fatalf("expected 8 space-separated header fields, got %d from %q", len(fields), got)
	}
	if fields[0] != "<134>1" {
		t.Errorf("field 0 = %q, want %q", fields[0], "<134>1")
	}
	if !rfc5424TimestampPattern.MatchString(fields[1]) {
		t.Errorf("timestamp %q does not match the RFC 5424 pattern", fields[1])
	}
	if fields[2] != "myhost" {
		t.Errorf("hostname = %q, want %q", fields[2], "myhost")
	}
	if fields[3] != "myapp" {
		t.Errorf("appname = %q, want %q", fields[3], "myapp")
	}
	for i := 4; i <= 6; i++ {
		if fields[i] != "-" {
			t.Errorf("field %d = %q, want %q (procid/msgid/structured-data are nil)", i, fields[i], "-")
		}
	}
	if fields[7] != "hello world" {
		t.Errorf("msg = %q, want %q", fields[7], "hello world")
	}
}

func TestBuildSyslogMessage_RFC5424DefaultsMissingFieldsToNil(t *testing.T) {
	config := SyslogConfig{
		Facility:      16,
		Severity:      6,
		MessageFormat: MessageFormatRFC5424,
	}

	got, err := buildSyslogMessage(config, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fields := strings.SplitN(got, " ", 8)
	if len(fields) != 8 {
		t.Fatalf("expected 8 space-separated header fields, got %d from %q", len(fields), got)
	}
	if fields[2] != "-" {
		t.Errorf("empty hostname should render as %q, got %q", "-", fields[2])
	}
	if fields[3] != "-" {
		t.Errorf("empty appname should render as %q, got %q", "-", fields[3])
	}
}

func TestBuildSyslogMessage_RFC3164StructureUnchanged(t *testing.T) {
	config := SyslogConfig{
		Facility:      16,
		Severity:      6,
		Hostname:      "myhost",
		Appname:       "myapp",
		MessageFormat: MessageFormatRFC3164,
	}

	got, err := buildSyslogMessage(config, "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !rfc3164Pattern.MatchString(got) {
		t.Errorf("message %q does not match the RFC 3164 pattern %q", got, rfc3164Pattern.String())
	}
	if strings.HasPrefix(got, "<134>1 ") {
		t.Errorf("RFC 3164 must not emit a version digit, got %q", got)
	}
}

// =============================================================================
// CONFIG VALIDATION TESTS
// =============================================================================

func TestValidateConfig_MessageFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  MessageFormat
		wantErr bool
	}{
		{name: "empty format is valid (legacy templates)", format: "", wantErr: false},
		{name: "rfc5424 is valid", format: MessageFormatRFC5424, wantErr: false},
		{name: "rfc3164 is valid", format: MessageFormatRFC3164, wantErr: false},
		{name: "raw-pri is valid", format: MessageFormatRawPRI, wantErr: false},
		{name: "unknown format is rejected", format: "rfc9999", wantErr: true},
		{name: "wrong case is rejected", format: "RFC5424", wantErr: true},
		{name: "whitespace only is rejected", format: " ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := SyslogConfig{
				Address:       "127.0.0.1",
				Port:          "514",
				Protocol:      "tcp",
				MessageFormat: tt.format,
			}

			err := validateConfig(&config)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			msg := err.Error()
			if !strings.Contains(msg, "message format") {
				t.Errorf("error should mention the message format, got %q", msg)
			}
			if r := []rune(msg)[0]; unicode.IsUpper(r) {
				t.Errorf("error strings must start lowercase, got %q", msg)
			}
		})
	}
}

func TestValidateConfig_PreservesExistingBehaviour(t *testing.T) {
	config := SyslogConfig{
		Address:  "127.0.0.1",
		Port:     "514",
		Protocol: "tcp",
	}

	if err := validateConfig(&config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.FramingMethod != RecommendedFramingMethod() {
		t.Errorf("FramingMethod = %q, want %q", config.FramingMethod, RecommendedFramingMethod())
	}
	if config.Appname != "sendlog" {
		t.Errorf("Appname = %q, want %q", config.Appname, "sendlog")
	}
	if config.Hostname == "" {
		t.Error("Hostname should be defaulted, got empty string")
	}
	if config.MessageFormat != "" {
		t.Errorf("validateConfig must not invent a MessageFormat, got %q", config.MessageFormat)
	}
}

// =============================================================================
// FRAMING AGREEMENT TEST
// The octet-counting frame length must be computed from the exact raw-pri bytes.
// =============================================================================

func TestFrame_OctetCountMatchesRawPRIByteLength(t *testing.T) {
	const esetMessage = `1 2022-04-11T10:41:05.300Z esmc ERAServer 11804 - - {"event_type":"FirewallAggregated_Event","severity":"Warning","hostname":"东京-01"}`

	config := SyslogConfig{
		Facility:      1,
		Severity:      4,
		MessageFormat: MessageFormatRawPRI,
	}

	syslogMsg, err := buildSyslogMessage(config, esetMessage)
	if err != nil {
		t.Fatalf("buildSyslogMessage returned unexpected error: %v", err)
	}

	framer := NewFramer(FramingConfig{Method: OctetCounting, ValidateUTF8: true})
	framed, err := framer.Frame(syslogMsg)
	if err != nil {
		t.Fatalf("Frame returned unexpected error: %v", err)
	}

	prefix, payload, found := strings.Cut(string(framed), " ")
	if !found {
		t.Fatalf("framed output has no octet-count prefix: %q", string(framed))
	}

	wantLen := len([]byte(syslogMsg))
	if prefix != strconv.Itoa(wantLen) {
		t.Errorf("octet count prefix = %q, want %q", prefix, strconv.Itoa(wantLen))
	}
	if payload != syslogMsg {
		t.Errorf("framed payload = %q, want %q", payload, syslogMsg)
	}
	if len([]byte(payload)) != wantLen {
		t.Errorf("framed payload is %d bytes, want %d", len([]byte(payload)), wantLen)
	}
}
