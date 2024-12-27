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

package android

import (
	"strings"
	"testing"
)

func TestVintfManifestBuildAction(t *testing.T) {
	bp := `
	vintf_fragment {
		name: "test_vintf_fragment",
		src: "test_vintf_file",
	}
	`

	testResult := PrepareForTestWithAndroidBuildComponents.RunTestWithBp(t, bp)

	vintfFragmentBuild := testResult.TestContext.ModuleForTests("test_vintf_fragment", "android_arm64_armv8-a").Rule("assemble_vintf")
	if !strings.Contains(vintfFragmentBuild.RuleParams.Command, "assemble_vintf") {
		t.Errorf("Vintf_manifest build command does not process with assemble_vintf : " + vintfFragmentBuild.RuleParams.Command)
	}
}
