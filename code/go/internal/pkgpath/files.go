// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package pkgpath

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/joeshaw/multierror"

	"gopkg.in/yaml.v3"

	"github.com/elastic/package-spec/v3/code/go/internal/fspath"
)

// File represents a file in the package.
type File struct {
	os.FileInfo

	fsys    fspath.FS
	path    string
	content []byte
	parsed  any
}

func newFile(fsys fspath.FS, path string) (File, error) {
	stat, err := fs.Stat(fsys, path)
	if err != nil {
		return File{}, err
	}

	return File{
		FileInfo: stat,
		fsys:     fsys,
		path:     path,
	}, nil
}

// Files finds files for the given glob
func Files(fsys fspath.FS, glob string) ([]File, error) {
	paths, err := fs.Glob(fsys, glob)
	if err != nil {
		return nil, err
	}

	var errs multierror.Errors
	var files = make([]File, 0)
	for _, path := range paths {
		file, err := newFile(fsys, path)
		if err != nil {
			errs = append(errs, err)
		}
		files = append(files, file)
	}

	return files, errs.Err()
}

// Values returns values within the file matching the given path. Paths
// should be expressed using JSONPath syntax. This method is only supported
// for YAML and JSON files.
func (f File) Values(path string) (any, error) {
	parsed, err := f.getParsedContent()
	if err != nil {
		return nil, err
	}
	return jsonpath.Get(path, parsed)
}

// Path returns the complete path to the file.
func (f File) Path() string {
	return f.path
}

// ReadAll reads and returns the entire contents of the file.
func (f File) ReadAll() ([]byte, error) {
	return f.getContent()
}

// getContent reads the content in a lazy way, as no all Files are used to read
// the content of the file. The value is cached to avoid having to read it
// again in cases where the file is accessed multiple times.
func (f File) getContent() ([]byte, error) {
	if f.content == nil {
		content, err := fs.ReadFile(f.fsys, f.path)
		if err != nil {
			return nil, err
		}
		f.content = content
	}
	return f.content, nil
}

// getParsedContent parses the content in a lazy way, as no all Files are used to
// access its content using Values. The value is cached to avoid having to read it
// again in cases where the file is accessed multiple times.
func (f File) getParsedContent() (any, error) {
	if f.parsed == nil {
		content, err := f.getContent()
		if err != nil {
			return nil, err
		}

		fileExt := filepath.Ext(f.Name())
		var parsed any
		switch fileExt {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(content, &parsed); err != nil {
				return nil, fmt.Errorf("unmarshalling YAML file failed (path: %s): %w", f.fsys.Path(f.path), err)
			}
		case ".json":
			if err := json.Unmarshal(content, &parsed); err != nil {
				return nil, fmt.Errorf("unmarshalling JSON file failed (path: %s): %w", f.fsys.Path(f.path), err)
			}
		default:
			return nil, fmt.Errorf("cannot extract values from file type = %s", strings.TrimLeft(fileExt, "."))
		}
		f.parsed = parsed
	}
	return f.parsed, nil
}
