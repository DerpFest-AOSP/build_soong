// Copyright 2024 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aconfig

import (
	"testing"

	"android/soong/android"
)

func TestTwoAconfigDeclarationsPerPackage(t *testing.T) {
	bp := `
		aconfig_declarations {
			name: "module_name.foo",
			package: "com.example.package",
			container: "com.android.foo",
			srcs: [
				"foo.aconfig",
			],
		}

		aconfig_declarations {
			name: "module_name.bar",
			package: "com.example.package",
			container: "com.android.foo",
			srcs: [
				"bar.aconfig",
			],
		}
	`
	errMsg := "Only one aconfig_declarations allowed for each package."
	android.GroupFixturePreparers(
		PrepareForTestWithAconfigBuildComponents).
		ExtendWithErrorHandler(android.FixtureExpectsOneErrorPattern(errMsg)).
		RunTestWithBp(t, bp)
}
