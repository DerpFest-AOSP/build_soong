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

package apex

import (
	"android/soong/android"
	"android/soong/java"
	"fmt"
	"testing"
)

var checkContainerMatch = func(t *testing.T, name string, container string, expected bool, actual bool) {
	errorMessage := fmt.Sprintf("module %s container %s value differ", name, container)
	android.AssertBoolEquals(t, errorMessage, expected, actual)
}

func TestApexDepsContainers(t *testing.T) {
	result := android.GroupFixturePreparers(
		prepareForApexTest,
		java.PrepareForTestWithJavaSdkLibraryFiles,
		java.FixtureWithLastReleaseApis("mybootclasspathlib", "bar"),
	).RunTestWithBp(t, `
		apex {
			name: "myapex",
			key: "myapex.key",
			bootclasspath_fragments: [
				"mybootclasspathfragment",
			],
			updatable: true,
			min_sdk_version: "30",
		}
		apex_key {
			name: "myapex.key",
			public_key: "testkey.avbpubkey",
			private_key: "testkey.pem",
		}
		bootclasspath_fragment {
			name: "mybootclasspathfragment",
			contents: [
				"mybootclasspathlib",
			],
			apex_available: [
				"myapex",
			],
			hidden_api: {
				split_packages: ["*"],
			},
		}
		java_sdk_library {
			name: "mybootclasspathlib",
			srcs: [
				"mybootclasspathlib.java",
			],
			apex_available: [
				"myapex",
			],
			compile_dex: true,
			static_libs: [
				"food",
				"baz",
			],
			libs: [
				"bar.stubs",
			],
			min_sdk_version: "30",
			sdk_version: "current",
		}
		java_library {
			name: "food",
			srcs:[
				"A.java",
			],
			apex_available: [
				"myapex",
			],
			min_sdk_version: "30",
			sdk_version: "core_current",
		}
		java_sdk_library {
			name: "bar",
			srcs:[
				"A.java",
			],
			min_sdk_version: "30",
			sdk_version: "core_current",
		}
		java_library {
			name: "baz",
			srcs:[
				"A.java",
			],
			apex_available: [
				"//apex_available:platform",
				"myapex",
			],
			min_sdk_version: "30",
			sdk_version: "core_current",
		}
	`)
	testcases := []struct {
		moduleName        string
		variant           string
		isSystemContainer bool
		isApexContainer   bool
	}{
		{
			moduleName:        "mybootclasspathlib",
			variant:           "android_common_myapex",
			isSystemContainer: true,
			isApexContainer:   true,
		},
		{
			moduleName:        "mybootclasspathlib.impl",
			variant:           "android_common_apex30",
			isSystemContainer: true,
			isApexContainer:   true,
		},
		{
			moduleName:        "mybootclasspathlib.stubs",
			variant:           "android_common",
			isSystemContainer: true,
			isApexContainer:   false,
		},
		{
			moduleName:        "food",
			variant:           "android_common_apex30",
			isSystemContainer: true,
			isApexContainer:   true,
		},
		{
			moduleName:        "bar",
			variant:           "android_common",
			isSystemContainer: true,
			isApexContainer:   false,
		},
		{
			moduleName:        "baz",
			variant:           "android_common_apex30",
			isSystemContainer: true,
			isApexContainer:   true,
		},
	}

	for _, c := range testcases {
		m := result.ModuleForTests(c.moduleName, c.variant)
		containers, _ := android.OtherModuleProvider(result.TestContext.OtherModuleProviderAdaptor(), m.Module(), android.ContainersInfoProvider)
		belongingContainers := containers.BelongingContainers()
		checkContainerMatch(t, c.moduleName, "system", c.isSystemContainer, android.InList(android.SystemContainer, belongingContainers))
		checkContainerMatch(t, c.moduleName, "apex", c.isApexContainer, android.InList(android.ApexContainer, belongingContainers))
	}
}

func TestNonUpdatableApexDepsContainers(t *testing.T) {
	result := android.GroupFixturePreparers(
		prepareForApexTest,
		java.PrepareForTestWithJavaSdkLibraryFiles,
		java.FixtureWithLastReleaseApis("mybootclasspathlib", "bar"),
	).RunTestWithBp(t, `
		apex {
			name: "myapex",
			key: "myapex.key",
			bootclasspath_fragments: [
				"mybootclasspathfragment",
			],
			updatable: false,
		}
		apex_key {
			name: "myapex.key",
			public_key: "testkey.avbpubkey",
			private_key: "testkey.pem",
		}
		bootclasspath_fragment {
			name: "mybootclasspathfragment",
			contents: [
				"mybootclasspathlib",
			],
			apex_available: [
				"myapex",
			],
			hidden_api: {
				split_packages: ["*"],
			},
		}
		java_sdk_library {
			name: "mybootclasspathlib",
			srcs: [
				"mybootclasspathlib.java",
			],
			apex_available: [
				"myapex",
			],
			compile_dex: true,
			static_libs: [
				"food",
			],
			libs: [
				"bar.stubs",
			],
			sdk_version: "current",
		}
		java_library {
			name: "food",
			srcs:[
				"A.java",
			],
			apex_available: [
				"myapex",
			],
			sdk_version: "core_current",
		}
		java_sdk_library {
			name: "bar",
			srcs:[
				"A.java",
			],
			sdk_version: "none",
			system_modules: "none",
		}
	`)
	testcases := []struct {
		moduleName        string
		variant           string
		isSystemContainer bool
		isApexContainer   bool
	}{
		{
			moduleName:        "mybootclasspathlib",
			variant:           "android_common_myapex",
			isSystemContainer: true,
			isApexContainer:   true,
		},
		{
			moduleName:        "mybootclasspathlib.impl",
			variant:           "android_common_apex10000",
			isSystemContainer: true,
			isApexContainer:   true,
		},
		{
			moduleName:        "mybootclasspathlib.stubs",
			variant:           "android_common",
			isSystemContainer: true,
			isApexContainer:   false,
		},
		{
			moduleName:        "food",
			variant:           "android_common_apex10000",
			isSystemContainer: true,
			isApexContainer:   true,
		},
		{
			moduleName:        "bar",
			variant:           "android_common",
			isSystemContainer: true,
			isApexContainer:   false,
		},
	}

	for _, c := range testcases {
		m := result.ModuleForTests(c.moduleName, c.variant)
		containers, _ := android.OtherModuleProvider(result.TestContext.OtherModuleProviderAdaptor(), m.Module(), android.ContainersInfoProvider)
		belongingContainers := containers.BelongingContainers()
		checkContainerMatch(t, c.moduleName, "system", c.isSystemContainer, android.InList(android.SystemContainer, belongingContainers))
		checkContainerMatch(t, c.moduleName, "apex", c.isApexContainer, android.InList(android.ApexContainer, belongingContainers))
	}
}

func TestUpdatableAndNonUpdatableApexesIdenticalMinSdkVersion(t *testing.T) {
	result := android.GroupFixturePreparers(
		prepareForApexTest,
		java.PrepareForTestWithJavaSdkLibraryFiles,
		android.FixtureMergeMockFs(android.MockFS{
			"system/sepolicy/apex/myapex_non_updatable-file_contexts": nil,
			"system/sepolicy/apex/myapex_updatable-file_contexts":     nil,
		}),
	).RunTestWithBp(t, `
		apex {
			name: "myapex_non_updatable",
			key: "myapex_non_updatable.key",
			java_libs: [
				"foo",
			],
			updatable: false,
			min_sdk_version: "30",
		}
		apex_key {
			name: "myapex_non_updatable.key",
			public_key: "testkey.avbpubkey",
			private_key: "testkey.pem",
		}

		apex {
			name: "myapex_updatable",
			key: "myapex_updatable.key",
			java_libs: [
				"foo",
			],
			updatable: true,
			min_sdk_version: "30",
		}
		apex_key {
			name: "myapex_updatable.key",
			public_key: "testkey.avbpubkey",
			private_key: "testkey.pem",
		}

		java_library {
			name: "foo",
			srcs:[
				"A.java",
			],
			apex_available: [
				"myapex_non_updatable",
				"myapex_updatable",
			],
			min_sdk_version: "30",
			sdk_version: "current",
		}
	`)

	fooApexVariant := result.ModuleForTests("foo", "android_common_apex30")
	containers, _ := android.OtherModuleProvider(result.TestContext.OtherModuleProviderAdaptor(), fooApexVariant.Module(), android.ContainersInfoProvider)
	belongingContainers := containers.BelongingContainers()
	checkContainerMatch(t, "foo", "system", true, android.InList(android.SystemContainer, belongingContainers))
	checkContainerMatch(t, "foo", "apex", true, android.InList(android.ApexContainer, belongingContainers))
}
