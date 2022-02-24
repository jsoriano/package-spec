// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package semantic

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	ve "github.com/elastic/package-spec/code/go/internal/errors"
	"github.com/elastic/package-spec/code/go/internal/fspath"
)

const maxFieldsPerDataStream = 1024

// ValidateFieldsLimits verifies limits on fields.
func ValidateFieldsLimits(fsys fspath.FS) ve.ValidationErrors {
	counts := make(map[string]uint)
	countField := func(fieldsFile string, f field) ve.ValidationErrors {
		if len(f.Fields) > 0 {
			// Don't count groups
			return nil
		}

		dataStream, err := dataStreamFromFieldsPath(fsys.Path(), fieldsFile)
		if err != nil {
			return ve.ValidationErrors{err}
		}
		count, _ := counts[dataStream]
		counts[dataStream] = count + 1
		return nil
	}

	err := validateFields(fsys, countField)
	if err != nil {
		return err
	}

	var errs ve.ValidationErrors
	for dataStream, count := range counts {
		if count > maxFieldsPerDataStream {
			errs = append(errs, errors.Errorf("data stream %s has more than %d fields (%d)", dataStream, maxFieldsPerDataStream, count))
		}
	}
	return errs
}

func dataStreamFromFieldsPath(pkgRoot, fieldsFile string) (string, error) {
	dataStreamPath := filepath.Clean(filepath.Join(pkgRoot, "data_stream"))
	relPath, err := filepath.Rel(dataStreamPath, filepath.Clean(fieldsFile))
	if err != nil {
		return "", fmt.Errorf("looking for fields file (%s) in data streams path (%s): %w", fieldsFile, dataStreamPath, err)
	}

	parts := strings.SplitN(relPath, string(filepath.Separator), 2)
	if len(parts) != 2 {
		return "", errors.Errorf("could not find data stream for fields file %s", fieldsFile)
	}
	dataStream := parts[0]
	return dataStream, nil
}