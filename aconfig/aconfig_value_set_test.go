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
	"testing"

	"android/soong/android"

	"github.com/google/blueprint"
)

func TestAconfigValueSet(t *testing.T) {
	bp := `
				aconfig_values {
					name: "one",
					srcs: [ "blah.aconfig_values" ],
					package: "foo.package"
				}

				aconfig_value_set {
					name: "module_name",
          values: [ "one" ],
				}
			`
	result := runTest(t, android.FixtureExpectsNoErrors, bp)

	module := result.ModuleForTests("module_name", "").Module().(*ValueSetModule)

	// Check that the provider has the right contents
	depData, _ := android.OtherModuleProvider(result, module, valueSetProviderKey)
	android.AssertStringEquals(t, "AvailablePackages", "blah.aconfig_values", depData.AvailablePackages["foo.package"][0].String())
}

func TestAconfigValueSetBpGlob(t *testing.T) {
	result := android.GroupFixturePreparers(
		PrepareForTestWithAconfigBuildComponents,
		android.FixtureMergeMockFs(
			map[string][]byte{
				// .../some_release/android.foo/
				"some_release/android.foo/Android.bp": []byte(`
				aconfig_values {
					name: "aconfig-values-platform_build_release-some_release-android.foo-all",
					package: "android.foo",
					srcs: [
						"*.textproto",
					],
				}
				`),
				"some_release/android.foo/flag.textproto": nil,

				// .../some_release/android.bar/
				"some_release/android.bar/Android.bp": []byte(`
				aconfig_values {
					name: "aconfig-values-platform_build_release-some_release-android.bar-all",
					package: "android.bar",
					srcs: [
						"*.textproto",
					],
				}
				`),
				"some_release/android.bar/flag.textproto": nil,

				// .../some_release/
				"some_release/Android.bp": []byte(`
				aconfig_value_set {
					name: "aconfig_value_set-platform_build_release-some_release",
					srcs: [
						"*/Android.bp",
					],
				}
				`),
			},
		),
	).RunTest(t)

	checkModuleHasDependency := func(name, variant, dep string) bool {
		t.Helper()
		module := result.ModuleForTests(name, variant).Module()
		depFound := false
		result.VisitDirectDeps(module, func(m blueprint.Module) {
			if m.Name() == dep {
				depFound = true
			}
		})
		return depFound
	}
	android.AssertBoolEquals(t,
		"aconfig_value_set expected to depend on aconfig_value via srcs",
		true,
		checkModuleHasDependency(
			"aconfig_value_set-platform_build_release-some_release",
			"",
			"aconfig-values-platform_build_release-some_release-android.foo-all",
		),
	)
	android.AssertBoolEquals(t,
		"aconfig_value_set expected to depend on aconfig_value via srcs",
		true,
		checkModuleHasDependency(
			"aconfig_value_set-platform_build_release-some_release",
			"",
			"aconfig-values-platform_build_release-some_release-android.bar-all",
		),
	)
}

func TestAconfigValueSetBpGlobError(t *testing.T) {
	android.GroupFixturePreparers(
		PrepareForTestWithAconfigBuildComponents,
		android.FixtureMergeMockFs(
			map[string][]byte{
				// .../some_release/android.bar/
				"some_release/android.bar/Android.bp": []byte(`
				aconfig_values {
					name: "aconfig-values-platform_build_release-some_release-android_bar-all",
					package: "android.bar",
					srcs: [
						"*.textproto",
					],
				}
				`),
				"some_release/android.bar/flag.textproto": nil,

				// .../some_release/
				"some_release/Android.bp": []byte(`
				aconfig_value_set {
					name: "aconfig_value_set-platform_build_release-some_release",
					srcs: [
						"*/Android.bp",
					],
				}
				`),
			},
		),
	).ExtendWithErrorHandler(android.FixtureExpectsOneErrorPattern(
		`module "aconfig_value_set-platform_build_release-some_release": module ` +
			`"aconfig-values-platform_build_release-some_release-android.bar-all" not found. ` +
			`Rename the aconfig_values module defined in "some_release/android.bar/Android.bp" ` +
			`to "aconfig-values-platform_build_release-some_release-android.bar-all"`),
	).RunTest(t)
}
