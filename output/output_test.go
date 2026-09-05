// Copyright 2024 OWASP CRS Project
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

var testString = "test"

var format = `This is the %s`

// TODO:
// GitHub
// GitLab
// CodeBuild
// CircleCI
// Jenkins
var outputTest = []struct {
	oType    string
	expected string
}{
	// quiet and json don't emit anything through Printf (json is handled via zerolog).
	// github defaults to plain log text; annotations must be opted into.
	{"quiet", ""},
	{"github", "This is the test"},
	{"normal", "This is the test"},
	{"json", ""},
}

type outputTestSuite struct {
	suite.Suite
}

func (s *outputTestSuite) SetupSuite() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestOutputTestSuite(t *testing.T) {
	suite.Run(t, new(outputTestSuite))
}

func (s *outputTestSuite) TestOutput() {
	for i, test := range outputTest {
		var b bytes.Buffer
		o := NewOutput(test.oType, &b)

		err := o.Printf(format, testString)
		s.Require().NoError(err, "Error! in test %d", i)
		s.Equal(test.expected, b.String(), "unexpected output in test %d", i)
	}
}

func (s *outputTestSuite) TestNormalCatalogOutput() {
	var b bytes.Buffer

	normal := NewOutput("normal", &b)
	for _, v := range normalCatalog {
		normal.RawPrint(v)
		s.Equal(b.String(), v, "output is not equal")
		// reset buffer
		b.Reset()
	}
}

func (s *outputTestSuite) TestPlainCatalogOutput() {
	var b bytes.Buffer

	normal := NewOutput("normal", &b)
	for _, v := range createPlainCatalog(normalCatalog) {
		normal.RawPrint(v)
		s.Equal(b.String(), v, "output is not equal")
		// reset buffer
		b.Reset()
	}
}

func (s *outputTestSuite) TestGitHubAnnotationError() {
	var b bytes.Buffer
	o := NewOutput("github", &b)
	o.SetSeverity(AnnotationError)
	o.SetAnnotationsEnabled(true)

	err := o.Printf("- %s failed in %s", "920100-1", "5ms")
	s.Require().NoError(err)
	// file/line/endLine are deliberately omitted: go-ftw has no per-test
	// line info, and a file without a line renders as a misleading `#L0`.
	s.Equal("::error::- 920100-1 failed in 5ms", b.String())
}

func (s *outputTestSuite) TestGitHubAnnotationEscapesSpecialChars() {
	var b bytes.Buffer
	o := NewOutput("github", &b)
	o.SetAnnotationsEnabled(true)

	err := o.Printf("100%% done\nwith %s", "newline")
	s.Require().NoError(err)
	s.Equal("::notice::100%25 done%0Awith newline", b.String())
}

// Println's line break must remain a real newline: GitHub only parses one
// workflow command per line, so escaping it would glue every command into a
// single (rejected) annotation.
func (s *outputTestSuite) TestGitHubAnnotationPrintlnKeepsRealNewline() {
	var b bytes.Buffer
	o := NewOutput("github", &b)
	o.SetAnnotationsEnabled(true)

	err := o.Println("+ passed in %s", "5ms")
	s.Require().NoError(err)
	err = o.Println("- failed")
	s.Require().NoError(err)
	s.Equal("::notice::+ passed in 5ms\n::notice::- failed\n", b.String())
}

// GitHub copy drops terminal decorations: bullets, banners, kaomoji.
func (s *outputTestSuite) TestGitHubCatalogCopy() {
	var b bytes.Buffer
	o := NewOutput("github", &b)

	s.Equal("test 920100-1 failed in 5ms (RTT 2ms)", fmt.Sprintf(o.Message("- %s failed in %s (RTT %s)"), "920100-1", "5ms", "2ms"))
	s.Equal("run 10 total tests in 1m0s", fmt.Sprintf(o.Message("+ run %d total tests in %s"), 10, "1m0s"))
	s.Equal("ignored 69 tests", fmt.Sprintf(o.Message("^ ignored %d tests"), 69))
	s.Equal("All tests successful!", o.Message("\\o/ All tests successful!"))
}
