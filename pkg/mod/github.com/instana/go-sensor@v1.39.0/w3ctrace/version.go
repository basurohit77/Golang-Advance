// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package w3ctrace

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

// Version represents the W3C trace context version. It defines the format of `traceparent` header
type Version uint8

const (
	// Version_Invalid represend an invalid  W3C Trace Context version
	Version_Invalid Version = iota
	// Version_0 represent the W3C Trace Context version 00
	Version_0
	// Version_Max is the latest version of W3C Trace Context supported by this package
	Version_Max = Version_0
)

// ParseVersion parses the version part of a `traceparent` header value. It returns ErrContextCorrupted
// if the version is malformed
func ParseVersion(s string) (Version, error) {
	if len(s) < 2 || (len(s) > 2 && s[2] != '-') {
		return Version_Invalid, ErrContextCorrupted
	}
	s = s[:2]

	if s == "ff" {
		return Version_Invalid, nil
	}

	ver, err := strconv.ParseUint(s, 16, 8)
	if err != nil {
		return Version_Invalid, ErrContextCorrupted
	}

	return Version(ver + 1), nil
}

// String returns string representation of a trace parent version. The returned value is compatible with the
// `traceparent` header format. The caller should take care of handling the Version_Unknown, otherwise this
// method will return "ff" which is considered invalid
func (ver Version) String() string {
	if ver == Version_Invalid {
		return "ff"
	}

	return fmt.Sprintf("%02x", uint8(ver)-1)
}

// parseParent parses the version-format string as described in https://www.w3.org/TR/trace-context/#version-format
func (ver Version) parseParent(s string) (Parent, error) {
	if ver == Version_Invalid {
		return Parent{Version: Version_Invalid}, ErrContextCorrupted
	}

	// If a higher version is detected, we try to parse it as the highest version
	// that is currently supported
	if ver > Version_Max {
		ver = Version_Max
	}

	switch ver {
	case Version_0:
		return parseV0Parent(s)
	default:
		return Parent{Version: ver}, ErrUnsupportedVersion
	}
}

// formatParent returns the version-format string for this version as described in
// https://www.w3.org/TR/trace-context/#version-format. The returned value is
// empty if the version is not supported or invalid
func (ver Version) formatParent(p Parent) string {
	// Construct the new traceparent field according to the highest version of
	// the specification known to the implementation
	if ver > Version_Max {
		ver = Version_Max
	}

	switch ver {
	case Version_0:
		return formatV0Parent(p)
	default:
		return ""
	}
}

// W3C Trace Context v0 version-format parsing/formatting
const (
	v0SampledFlag uint8 = 1 << iota
)

func parseV0Parent(s string) (Parent, error) {
	const (
		versionFormatLen = 55
		versionLen       = 2
		traceIDLen       = 32
		parentIDLen      = 16
		flagsLen         = 2
		separator        = '-'
		invalidTraceID   = "00000000000000000000000000000000"
		invalidParentID  = "0000000000000000"
	)

	// trim version part
	if len(s) < versionFormatLen || s[versionLen] != separator {
		return Parent{}, ErrContextCorrupted
	}
	_, s = s[:versionLen], s[versionLen+1:]

	// extract trace id
	if s[traceIDLen] != separator {
		return Parent{}, ErrContextCorrupted
	}
	traceID, s := s[:traceIDLen], s[traceIDLen+1:]

	if traceID == invalidTraceID || !isHex(traceID) {
		return Parent{}, ErrContextCorrupted
	}

	// extract parent id
	if s[parentIDLen] != separator {
		return Parent{}, ErrContextCorrupted
	}
	parentID, s := s[:parentIDLen], s[parentIDLen+1:]

	if parentID == invalidParentID || !isHex(parentID) {
		return Parent{}, ErrContextCorrupted
	}

	// extract and parse flags
	if len(s) > flagsLen && s[flagsLen] != separator {
		return Parent{}, ErrContextCorrupted
	}

	flags, err := strconv.ParseUint(s[:flagsLen], 16, 8)
	if err != nil {
		return Parent{}, ErrContextCorrupted
	}

	return Parent{
		Version:  Version_0,
		TraceID:  traceID,
		ParentID: parentID,
		Flags: Flags{
			Sampled: uint8(flags)&v0SampledFlag != 0,
		},
	}, nil
}

func formatV0Parent(p Parent) string {
	var flags uint8
	if p.Flags.Sampled {
		flags |= v0SampledFlag
	}

	return fmt.Sprintf("00-%032s-%016s-%02x", p.TraceID, p.ParentID, flags)
}

func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}
