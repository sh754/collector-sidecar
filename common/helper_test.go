// This file is part of Graylog.
//
// Graylog is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Graylog is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Graylog.  If not, see <http://www.gnu.org/licenses/>.

package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestGetCollectorIdFromExistingFile(t *testing.T) {
	content := []byte(" 2135792e-8556-4bf0-8aef-503f29890b09 \n")
	tmpfile, err := ioutil.TempFile("", "test-node-id")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	result := GetCollectorId("file:/" + tmpfile.Name())

	expect := "2135792e-8556-4bf0-8aef-503f29890b09"
	if !strings.Contains(result, expect) {
		t.Fail()
	}
}

func TestGetCollectorIdFromNonExistingFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-node-id")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	tmpfile := filepath.Join(dir, "node-id")
	result := GetCollectorId("file:/" + tmpfile)
	match, err := regexp.Match("^[0-9a-f]{8}-", []byte(result))
	if err != nil {
		t.Fatal(err)
	}

	if !match {
		t.Fail()
	}
}

func TestEncloseWithWithoutAction(t *testing.T) {
	content := "/some regex/"
	result := EncloseWith(content, "/")

	if result != content {
		t.Fail()
	}
}

func TestEncloseWithPrepend(t *testing.T) {
	content := "some regex/"
	result := EncloseWith(content, "/")

	if result != "/"+content {
		t.Fail()
	}
}

func TestEncloseWithAppend(t *testing.T) {
	content := "/some regex"
	result := EncloseWith(content, "/")

	if result != content+"/" {
		t.Fail()
	}
}

func TestEncloseWithoutData(t *testing.T) {
	content := ""
	result := EncloseWith(content, "/")

	if result != "" {
		t.Fail()
	}
}

func TestPathMatch(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-path-match")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	execfile := filepath.Join(dir, "myexec")
	patternList := []string{"/usr/bin/moo", "/sbin/bar"}
	result, err := PathMatch(execfile, patternList)
	if err != nil || result.Match {
		t.Fatalf("'%s' should not match patternList %v err '%v' result '%v'", execfile, patternList, err, result)
	}
	if result.DoesExist {
		t.Fatalf("DoesExist should not be set")
	}

	os.Create(execfile)
	patternList = []string{"/usr/bin/moo", "/sbin/bar", execfile}
	result, err = PathMatch(execfile, patternList)
	if err != nil || !result.Match {
		t.Fatalf("'%s' should match patternList %v err '%v' result '%v'", execfile, patternList, err, result)
	}
	if !result.DoesExist {
		t.Fatalf("DoesExist should be set")
	}

	patternList = []string{"/usr/bin/moo", "/sbin/bar", dir + "/*"}
	result, err = PathMatch(execfile, patternList)
	if err != nil || !result.Match {
		t.Fatalf("'%s' should match globbing patternList %v err '%v' result '%v'",
			execfile, patternList, err, result)
	}

}

func TestPathMatchSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}

	dir, err := ioutil.TempDir("", "test-path-match")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	execfile := filepath.Join(dir, "myexec")
	symlink := filepath.Join(dir, "symlink")
	if err = os.Symlink(execfile, symlink); err != nil {
		t.Fatal()
	}
	patternList := []string{"moo", "bar", execfile}
	result, err := PathMatch(symlink, patternList)
	if err == nil {
		t.Fatalf("broken symlinks should report an error")
	}

	os.Create(execfile)
	result, err = PathMatch(symlink, patternList)
	if err != nil || !result.Match {
		t.Fatalf("'%s' should match patternList %v err '%v' result '%v'", execfile, patternList, err, result)
	}
	if !result.IsLink {
		t.Fatalf("result.IsLink is false")
	}
	if result.Path != execfile {
		t.Fatalf("result.Path did not contain resolved symlink")
	}
}
