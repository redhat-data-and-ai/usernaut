/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package snowflake

import (
	"net/url"
	"unicode"
)

// quoteSnowflakeIdentifier wraps the identifier in double quotes and URL-encodes
// it if it doesn't conform to Snowflake's unquoted identifier rules (e.g. starts
// with a digit). See https://docs.snowflake.com/en/sql-reference/identifiers-syntax
func quoteSnowflakeIdentifier(name string) string {
	if needsQuoting(name) {
		return url.PathEscape(`"` + name + `"`)
	}
	return name
}

func needsQuoting(name string) bool {
	if len(name) == 0 {
		return true
	}
	first := rune(name[0])
	if !unicode.IsLetter(first) && first != '_' {
		return true
	}
	for _, ch := range name[1:] {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' && ch != '$' {
			return true
		}
	}
	return false
}
