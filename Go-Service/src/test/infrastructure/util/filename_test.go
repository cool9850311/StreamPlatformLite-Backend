package infrastructure

import (
	"strings"
	"testing"
	"unicode/utf8"

	domainErrors "Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/infrastructure/util"

	"github.com/stretchr/testify/assert"
)

// TestSanitizeDownloadFilename covers plan table A-1.
//
// Every row may assert an exact result (want) and/or a set of extra
// expectations (check). Two invariants from contract A are asserted for every
// row regardless: the result is non-empty and ends with ".mp4" case
// insensitively.
func TestSanitizeDownloadFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		check func(t *testing.T, got string)
	}{
		{
			name:  "normal CJK title is preserved",
			input: "阿豹的ASMR",
			want:  "阿豹的ASMR.mp4",
		},
		{
			name:  "path traversal separators and edge dots are removed",
			input: "../../etc/passwd",
			want:  "etcpasswd.mp4",
		},
		{
			name:  "both separator flavours are removed",
			input: `a/b\c`,
			check: func(t *testing.T, got string) {
				assert.NotContains(t, got, "/")
				assert.NotContains(t, got, `\`)
			},
		},
		{
			name:  "leading dash is deliberately preserved (display name only)",
			input: "-y",
			want:  "-y.mp4",
		},
		{
			name:  "CR and LF are stripped (header injection)",
			input: "evil\r\nSet-Cookie: x",
			check: func(t *testing.T, got string) {
				assert.NotContains(t, got, "\r")
				assert.NotContains(t, got, "\n")
			},
		},
		{
			name:  "NUL byte is removed",
			input: "evil\x00title",
			check: func(t *testing.T, got string) {
				assert.NotContains(t, got, "\x00")
			},
		},
		{
			name:  "empty title falls back to the default",
			input: "",
			want:  util.DefaultRecordingFilename,
		},
		{
			name:  "whitespace only title falls back to the default",
			input: "   ",
			want:  util.DefaultRecordingFilename,
		},
		{
			name:  "single dot falls back to the default",
			input: ".",
			want:  util.DefaultRecordingFilename,
		},
		{
			name:  "double dot falls back to the default",
			input: "..",
			want:  util.DefaultRecordingFilename,
		},
		{
			name:  "over long title is truncated by rune without breaking UTF-8",
			input: strings.Repeat("阿", 300),
			check: func(t *testing.T, got string) {
				assert.LessOrEqual(t, utf8.RuneCountInString(got), util.MaxDownloadFilenameRunes,
					"result must fit within MaxDownloadFilenameRunes including the extension")
				assert.True(t, utf8.ValidString(got), "truncation must not split a multi-byte rune")
				assert.True(t, strings.HasSuffix(got, ".mp4"))
			},
		},
		{
			name:  "existing mp4 extension is not duplicated",
			input: "名稱.mp4",
			want:  "名稱.mp4",
			check: func(t *testing.T, got string) {
				assert.NotEqual(t, "名稱.mp4.mp4", got)
				assert.NotEqual(t, "名稱mp4.mp4", got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := util.SanitizeDownloadFilename(tt.input)

			// Contract A invariants, asserted for every case in the table.
			assert.NotEmpty(t, got, "result must never be empty")
			assert.True(t, strings.HasSuffix(strings.ToLower(got), ".mp4"),
				"result must always end with .mp4 (case insensitively), got %q", got)

			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

// TestValidateLivestreamTitle covers plan table A-2. The error value must be
// the package level domainErrors.ErrInvalidInput so the controller can map it
// to HTTP 400 with an == comparison.
func TestValidateLivestreamTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "normal title passes",
			input:   "阿豹的 ASMR 直播",
			wantErr: nil,
		},
		{
			name:    "title longer than MaxTitleRunes is rejected",
			input:   strings.Repeat("a", util.MaxTitleRunes+1),
			wantErr: domainErrors.ErrInvalidInput,
		},
		{
			name:    "title containing a control character is rejected",
			input:   "evil\r\nSet-Cookie: x",
			wantErr: domainErrors.ErrInvalidInput,
		},
		{
			name:    "empty title passes because Title is optional",
			input:   "",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := util.ValidateLivestreamTitle(tt.input)

			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			// Deliberately == and not errors.Is: the controller relies on the
			// exact sentinel value.
			assert.True(t, err == tt.wantErr, "expected %v, got %v", tt.wantErr, err)
		})
	}
}
