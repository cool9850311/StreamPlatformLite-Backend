package util

import (
	"strings"
	"unicode"
	"unicode/utf8"

	domainErrors "Go-Service/src/main/domain/entity/errors"
)

// DefaultRecordingFilename is the display name used when a livestream title
// cannot produce anything meaningful (empty, whitespace only, "." or "..").
const DefaultRecordingFilename = "recording.mp4"

// MaxDownloadFilenameRunes bounds the length of the value returned by
// SanitizeDownloadFilename, counted in runes and INCLUDING the ".mp4"
// extension. The base name is therefore truncated to
// MaxDownloadFilenameRunes - len(".mp4") runes before the extension is
// appended, so the returned string is never longer than this limit.
const MaxDownloadFilenameRunes = 100

// mp4Ext is the extension every download display name ends with.
const mp4Ext = ".mp4"

// SanitizeDownloadFilename turns a user supplied livestream title into a
// filename that is safe to place in a Content-Disposition header or in the
// "filename" field of the record status response.
//
// This function is NOT a security boundary. The security of the recording
// pipeline comes from the fact that the title never reaches the filesystem at
// all: recordings are always written to a compile time constant name and the
// conversion library no longer accepts a filename argument (see plan section
// 5.1, "storage name / display name decoupling"). What is done here is header
// hygiene and user experience only - keeping CR/LF out of a response header
// and keeping the browser's "Save as" dialog sane. Never rely on it to keep a
// value away from exec or from a path join.
//
// The returned value is always non-empty and always ends with ".mp4" (the
// original casing is kept when the title already supplied the extension).
func SanitizeDownloadFilename(title string) string {
	// 1+2. Drop path separators and every control character. unicode.IsControl
	// covers C0 (NUL, TAB, CR, LF, ...) and C1, so no hand written blocklist is
	// needed. Invalid UTF-8 bytes are mapped to U+FFFD by strings.Map, which
	// keeps the result valid UTF-8.
	cleaned := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || unicode.IsControl(r) {
			return -1
		}
		return r
	}, title)

	// 3. Trim surrounding whitespace and dots. Dots are trimmed only at the
	// edges, so an interior extension such as "名稱.mp4" survives untouched and
	// is later recognised by the extension check below.
	cleaned = trimFilenameEdges(cleaned)

	// 4. Nothing usable left (this also covers "." and ".." since the dot trim
	// above reduces them to an empty string).
	if cleaned == "" || cleaned == "." || cleaned == ".." {
		return DefaultRecordingFilename
	}

	// 6a. Split off an existing ".mp4" so the extension is never duplicated and
	// never chopped in half by the truncation below. The comparison is case
	// insensitive but the original casing is preserved in the result. Slicing by
	// bytes is safe here because ".mp4" is pure ASCII and therefore can never
	// overlap a UTF-8 continuation byte.
	base, ext := cleaned, mp4Ext
	if len(cleaned) >= len(mp4Ext) && strings.EqualFold(cleaned[len(cleaned)-len(mp4Ext):], mp4Ext) {
		base, ext = cleaned[:len(cleaned)-len(mp4Ext)], cleaned[len(cleaned)-len(mp4Ext):]
	}

	// 5. Truncate the base by runes - never by bytes - so a multi-byte sequence
	// is never split. The extension is appended afterwards and its length is
	// already reserved by maxBase.
	maxBase := MaxDownloadFilenameRunes - utf8.RuneCountInString(ext)
	if utf8.RuneCountInString(base) > maxBase {
		base = string([]rune(base)[:maxBase])
		// Truncation can expose new trailing dots or spaces, e.g. "abc..." would
		// otherwise become "abc....mp4".
		base = trimFilenameEdges(base)
	}

	if base == "" {
		return DefaultRecordingFilename
	}

	// 6b/7. Always non-empty, always ends with ".mp4", never ".mp4.mp4".
	return base + ext
}

// trimFilenameEdges removes leading and trailing whitespace and dots. Leading
// dots would produce a hidden file on unix hosts and trailing dots are silently
// dropped on Windows, so neither belongs in a display name.
func trimFilenameEdges(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == '.'
	})
}

// MaxTitleRunes bounds the livestream title length, counted in runes.
const MaxTitleRunes = 200

// ValidateLivestreamTitle rejects livestream titles that are too long or that
// contain control characters.
//
// This is a hygiene layer, explicitly NOT the primary defence (plan section
// 5.4, layer 3): rows created before this validation existed are still in the
// database and any future caller could bypass it, so the storage path must stay
// safe on its own. An empty title is accepted because Title is optional and
// that is the pre-existing behaviour.
//
// The error is returned as the package level domainErrors.ErrInvalidInput value
// so callers can compare it with == and map it to HTTP 400.
func ValidateLivestreamTitle(title string) error {
	if title == "" {
		return nil
	}

	if utf8.RuneCountInString(title) > MaxTitleRunes {
		return domainErrors.ErrInvalidInput
	}

	for _, r := range title {
		if unicode.IsControl(r) {
			return domainErrors.ErrInvalidInput
		}
	}

	return nil
}
