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

package aconfig

import (
	"slices"
	"strings"
	"testing"

	"android/soong/android"
)

func TestAconfigDeclarations(t *testing.T) {
	bp := `
		aconfig_declarations {
			name: "module_name",
			package: "com.example.package",
			container: "com.android.foo",
			exportable: true,
			srcs: [
				"foo.aconfig",
				"bar.aconfig",
			],
		}
	`
	result := runTest(t, android.FixtureExpectsNoErrors, bp)

	module := result.ModuleForTests("module_name", "").Module().(*DeclarationsModule)

	// Check that the provider has the right contents
	depData, _ := android.OtherModuleProvider(result, module, android.AconfigDeclarationsProviderKey)
	android.AssertStringEquals(t, "package", depData.Package, "com.example.package")
	android.AssertStringEquals(t, "container", depData.Container, "com.android.foo")
	android.AssertBoolEquals(t, "exportable", depData.Exportable, true)
	if !strings.HasSuffix(depData.IntermediateCacheOutputPath.String(), "/intermediate.pb") {
		t.Errorf("Missing intermediates proto path in provider: %s", depData.IntermediateCacheOutputPath.String())
	}
	if !strings.HasSuffix(depData.IntermediateDumpOutputPath.String(), "/intermediate.txt") {
		t.Errorf("Missing intermediates text path in provider: %s", depData.IntermediateDumpOutputPath.String())
	}
}

func TestAconfigDeclarationsWithExportableUnset(t *testing.T) {
	bp := `
		aconfig_declarations {
			name: "module_name",
			package: "com.example.package",
			container: "com.android.foo",
			srcs: [
				"foo.aconfig",
				"bar.aconfig",
			],
		}
	`
	result := runTest(t, android.FixtureExpectsNoErrors, bp)

	module := result.ModuleForTests("module_name", "").Module().(*DeclarationsModule)
	depData, _ := android.OtherModuleProvider(result, module, android.AconfigDeclarationsProviderKey)
	android.AssertBoolEquals(t, "exportable", depData.Exportable, false)
}

func TestAconfigDeclarationsWithContainer(t *testing.T) {
	bp := `
		aconfig_declarations {
			name: "module_name",
			package: "com.example.package",
			container: "com.android.foo",
			srcs: [
				"foo.aconfig",
			],
		}
	`
	result := runTest(t, android.FixtureExpectsNoErrors, bp)

	module := result.ModuleForTests("module_name", "")
	rule := module.Rule("aconfig")
	android.AssertStringEquals(t, "rule must contain container", rule.Args["container"], "--container com.android.foo")
}

func TestMandatoryProperties(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
		bp            string
	}{
		{
			name: "Srcs missing from aconfig_declarations",
			bp: `
				aconfig_declarations {
					name: "my_aconfig_declarations_foo",
					package: "com.example.package",
					container: "otherapex",
				}`,
			expectedError: `missing source files`,
		},
		{
			name: "Package missing from aconfig_declarations",
			bp: `
				aconfig_declarations {
					name: "my_aconfig_declarations_foo",
					container: "otherapex",
					srcs: ["foo.aconfig"],
				}`,
			expectedError: `missing package property`,
		},
		{
			name: "Container missing from aconfig_declarations",
			bp: `
				aconfig_declarations {
					name: "my_aconfig_declarations_foo",
					package: "com.example.package",
					srcs: ["foo.aconfig"],
				}`,
			expectedError: `missing container property`,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			errorHandler := android.FixtureExpectsAtLeastOneErrorMatchingPattern(test.expectedError)
			android.GroupFixturePreparers(PrepareForTestWithAconfigBuildComponents).
				ExtendWithErrorHandler(errorHandler).
				RunTestWithBp(t, test.bp)
		})
	}
}

func TestAssembleFileName(t *testing.T) {
	testCases := []struct {
		name          string
		releaseConfig string
		path          string
		expectedValue string
	}{
		{
			name:          "active release config",
			path:          "file.path",
			expectedValue: "file.path",
		},
		{
			name:          "release config FOO",
			releaseConfig: "FOO",
			path:          "file.path",
			expectedValue: "file-FOO.path",
		},
	}
	for _, test := range testCases {
		actualValue := assembleFileName(test.releaseConfig, test.path)
		if actualValue != test.expectedValue {
			t.Errorf("Expected %q found %q", test.expectedValue, actualValue)
		}
	}
}

func TestGenerateAndroidBuildActions(t *testing.T) {
	testCases := []struct {
		name         string
		buildFlags   map[string]string
		bp           string
		errorHandler android.FixtureErrorHandler
	}{
		{
			name: "generate extra",
			buildFlags: map[string]string{
				"RELEASE_ACONFIG_EXTRA_RELEASE_CONFIGS": "config2",
				"RELEASE_ACONFIG_VALUE_SETS":            "aconfig_value_set-config1",
				"RELEASE_ACONFIG_VALUE_SETS_config2":    "aconfig_value_set-config2",
			},
			bp: `
				aconfig_declarations {
					name: "module_name",
					package: "com.example.package",
					container: "com.android.foo",
					srcs: [
						"foo.aconfig",
						"bar.aconfig",
					],
				}
				aconfig_value_set {
					name: "aconfig_value_set-config1",
					values: []
				}
				aconfig_value_set {
					name: "aconfig_value_set-config2",
					values: []
				}
			`,
		},
	}
	for _, test := range testCases {
		fixture := PrepareForTest(t, addBuildFlagsForTest(test.buildFlags))
		if test.errorHandler != nil {
			fixture = fixture.ExtendWithErrorHandler(test.errorHandler)
		}
		result := fixture.RunTestWithBp(t, test.bp)
		module := result.ModuleForTests("module_name", "").Module().(*DeclarationsModule)
		depData, _ := android.OtherModuleProvider(result, module, android.AconfigReleaseDeclarationsProviderKey)
		expectedKeys := []string{""}
		for _, rc := range strings.Split(test.buildFlags["RELEASE_ACONFIG_EXTRA_RELEASE_CONFIGS"], " ") {
			expectedKeys = append(expectedKeys, rc)
		}
		slices.Sort(expectedKeys)
		actualKeys := []string{}
		for rc := range depData {
			actualKeys = append(actualKeys, rc)
		}
		slices.Sort(actualKeys)
		android.AssertStringEquals(t, "provider keys", strings.Join(expectedKeys, " "), strings.Join(actualKeys, " "))
		for _, rc := range actualKeys {
			if !strings.HasSuffix(depData[rc].IntermediateCacheOutputPath.String(), assembleFileName(rc, "/intermediate.pb")) {
				t.Errorf("Incorrect intermediates proto path in provider for release config %s: %s", rc, depData[rc].IntermediateCacheOutputPath.String())
			}
			if !strings.HasSuffix(depData[rc].IntermediateDumpOutputPath.String(), assembleFileName(rc, "/intermediate.txt")) {
				t.Errorf("Incorrect intermediates text path in provider for release config %s: %s", rc, depData[rc].IntermediateDumpOutputPath.String())
			}
		}
	}
}
