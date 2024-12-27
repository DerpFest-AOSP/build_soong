// Copyright (C) 2021 The Android Open Source Project
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

var PrepareForTestWithAconfigBuildComponents = android.FixtureRegisterWithContext(RegisterBuildComponents)

func runTest(t *testing.T, errorHandler android.FixtureErrorHandler, bp string) *android.TestResult {
	return PrepareForTest(t).
		ExtendWithErrorHandler(errorHandler).
		RunTestWithBp(t, bp)
}

func PrepareForTest(t *testing.T, preparers ...android.FixturePreparer) android.FixturePreparer {
	preparers = append([]android.FixturePreparer{PrepareForTestWithAconfigBuildComponents}, preparers...)
	return android.GroupFixturePreparers(preparers...)
}

func addBuildFlagsForTest(buildFlags map[string]string) android.FixturePreparer {
	return android.GroupFixturePreparers(
		android.FixtureModifyProductVariables(func(vars android.FixtureProductVariables) {
			if vars.BuildFlags == nil {
				vars.BuildFlags = make(map[string]string)
			}
			for k, v := range buildFlags {
				vars.BuildFlags[k] = v
			}
		}),
	)
}
