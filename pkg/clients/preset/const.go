/*
Copyright 2025.

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

package preset

import "time"

const (
	// Transport-level retries for transient network / Heimdall failures.
	presetHTTPRetryAttempts = 3

	// Application-level retries after receiving HTTP 429 from Preset SCIM.
	// attempt < presetRateLimitRetryAttempts allows this many retries after the first response.
	presetRateLimitRetryAttempts = 5

	presetRateLimitBaseBackoff = 5 * time.Second
	presetRateLimitMaxBackoff  = 60 * time.Second
	presetRateLimitExpFactor   = 2.0

	// Heimdall exponential backoff for non-429 transport retries.
	presetHTTPRetryInitialBackoff = 1 * time.Second
	presetHTTPRetryMaxBackoff     = 30 * time.Second
	presetHTTPRetryExpFactor      = 2.0
	presetHTTPRetryMaxJitter      = 500 * time.Millisecond

	logKeyAttempt = "attempt"
)
