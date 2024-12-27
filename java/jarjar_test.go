// Copyright 2018 Google Inc. All rights reserved.
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

package java

import (
	"fmt"
	"testing"

	"android/soong/android"
)

func AssertJarJarRename(t *testing.T, result *android.TestResult, libName, original, expectedRename string) {
	module := result.ModuleForTests(libName, "android_common")

	provider, found := android.OtherModuleProvider(result.OtherModuleProviderAdaptor(), module.Module(), JarJarProvider)
	android.AssertBoolEquals(t, fmt.Sprintf("found provider (%s)", libName), true, found)

	renamed, found := provider.Rename[original]
	android.AssertBoolEquals(t, fmt.Sprintf("found rename (%s)", libName), true, found)
	android.AssertStringEquals(t, fmt.Sprintf("renamed (%s)", libName), expectedRename, renamed)
}

func TestJarJarRenameDifferentModules(t *testing.T) {
	t.Parallel()
	result := android.GroupFixturePreparers(
		prepareForJavaTest,
	).RunTestWithBp(t, `
		java_library {
			name: "their_lib",
			jarjar_rename: ["com.example.a"],
		}

		java_library {
			name: "boundary_lib",
			jarjar_prefix: "RENAME",
			static_libs: ["their_lib"],
		}

		java_library {
			name: "my_lib",
			static_libs: ["boundary_lib"],
		}
	`)

	original := "com.example.a"
	renamed := "RENAME.com.example.a"
	AssertJarJarRename(t, result, "their_lib", original, "")
	AssertJarJarRename(t, result, "boundary_lib", original, renamed)
	AssertJarJarRename(t, result, "my_lib", original, renamed)
}

func TestJarJarRenameSameModule(t *testing.T) {
	t.Parallel()
	result := android.GroupFixturePreparers(
		prepareForJavaTest,
	).RunTestWithBp(t, `
		java_library {
			name: "their_lib",
			jarjar_rename: ["com.example.a"],
			jarjar_prefix: "RENAME",
		}

		java_library {
			name: "my_lib",
			static_libs: ["their_lib"],
		}
	`)

	original := "com.example.a"
	renamed := "RENAME.com.example.a"
	AssertJarJarRename(t, result, "their_lib", original, renamed)
	AssertJarJarRename(t, result, "my_lib", original, renamed)
}
