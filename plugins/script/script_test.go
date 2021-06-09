package script

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator/mocksender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// templateFile defines the contents of a template to be stored in a file, for testing.
type templateFile struct {
	name     string
	contents string
}

func createTestDir(files []templateFile) string {
	dir, err := ioutil.TempDir("", "template")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		f, err := os.Create(filepath.Join(dir, file.name))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		_, err = io.WriteString(f, file.contents)
		if err != nil {
			log.Fatal(err)
		}
	}
	return dir
}

func TestGetFiles(t *testing.T) {
	dir := createTestDir([]templateFile{
		{"10.json", ""},
		{"10.sh", ""},
		{"60-a.sh", ""},
		{"60-b-1.sh", ""},
		{"60-b-2.sh", ""},
	})
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)
	os.Chdir(dir)

	cases := []struct {
		path string
		want []string
	}{
		{"*", []string{"10.json", "10.sh", "60-a.sh", "60-b-1.sh", "60-b-2.sh"}},
		{"*.sh", []string{"10.sh", "60-a.sh", "60-b-1.sh", "60-b-2.sh"}},
		{"60-*.sh", []string{"60-a.sh", "60-b-1.sh", "60-b-2.sh"}},
		{"60-*-1.sh", []string{"60-b-1.sh"}},
		{"60-*-*.sh", []string{"60-b-1.sh", "60-b-2.sh"}},
	}

	ck := Check{}
	for _, c := range cases {
		ck.config = &checkConfig{filePath: c.path}
		got := ck.getFiles()
		assert.Equalf(t, c.want, got, "path %s", c.path)
	}

}

const sampleTextFormat = `[
{"metric":"test1", "value":1, counterType:"GUAGE", "tags":"a=1"},
{"metric":"test2", "value":2, counterType:"COUNTER", "tags":"a=2"},
{"metric":"test3", "value":3, counterType:"MONOTONIC_COUNT", "tags":"a=3"},
{"metric":"test4", "value":4, "tags":"a=4"}
]`

func TestCollect(t *testing.T) {
	flag.Set("v", "10")
	flag.Set("logtostderr", "true")
	flag.Parse()

	dir := createTestDir([]templateFile{
		{"test.sh", "cat"},
	})
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)

	check := new(Check)
	err := check.Configure([]byte(fmt.Sprintf(`
filePath: "/bin/sh"
params: "%s" 
stdin: '%s' 
`, filepath.Join(dir, "test.sh"), sampleTextFormat)), nil, "test")
	assert.Nil(t, err)

	sender := mocksender.NewMockSender(check.ID())

	sender.On("Gauge", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("Rate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("MonotonicCount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("Commit").Return()

	err = check.Run()
	assert.Nil(t, err)

	sender.AssertCalled(t, "Gauge", "go_gc_duration_seconds_quantile", float64(0.00010425500000000001), "", []string{"quantile:0"})
}
