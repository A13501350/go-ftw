// Copyright 2024 OWASP CRS Project
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"bytes"
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
	{"quiet", ""},
	{"github", "::notice::This is the test"},
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

func (s *outputTestSuite) TestGitHubAnnotationWithFile() {
	var b bytes.Buffer
	o := NewOutput("github", &b)
	o.SetCurrentTestFile("tests/920100.yaml")
	o.SetSeverity(AnnotationError)

	err := o.Printf("- %s failed in %s", "920100-1", "5ms")
	s.Require().NoError(err)
	s.Equal("::error file=tests/920100.yaml::- 920100-1 failed in 5ms", b.String())
}

func (s *outputTestSuite) TestGitHubAnnotationEscapesSpecialChars() {
	var b bytes.Buffer
	o := NewOutput("github", &b)

	err := o.Printf("100%% done\nwith %s", "newline")
	s.Require().NoError(err)
	s.Equal("::notice::100%25 done%0Awith newline", b.String())
}
