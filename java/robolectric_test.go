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

package java

import (
	"runtime"
	"testing"

	"android/soong/android"
)

var prepareRobolectricRuntime = android.GroupFixturePreparers(
	android.FixtureRegisterWithContext(func(ctx android.RegistrationContext) {
		RegisterRobolectricBuildComponents(ctx)
	}),
	android.FixtureAddTextFile("robolectric/Android.bp", `
	java_library {
		name: "Robolectric_all-target_upstream",
		srcs: ["Robo.java"]
	}

	java_library {
		name: "mockito-robolectric-prebuilt",
		srcs: ["Mockito.java"]
	}

	java_library {
		name: "truth",
		srcs: ["Truth.java"]
	}

	java_library {
		name: "junitxml",
		srcs: ["JUnitXml.java"]
	}

	java_library_host {
		name: "robolectric-host-android_all",
		srcs: ["Runtime.java"]
	}

	android_robolectric_runtimes {
		name: "robolectric-android-all-prebuilts",
		jars: ["android-all/Runtime.jar"],
		lib: "robolectric-host-android_all",
	}
	`),
)

func TestRobolectricJniTest(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires linux")
	}

	ctx := android.GroupFixturePreparers(
		PrepareForIntegrationTestWithJava,
		prepareRobolectricRuntime,
	).RunTestWithBp(t, `
	android_app {
		name: "inst-target",
		srcs: ["App.java"],
		platform_apis: true,
	}

	cc_library_shared {
		name: "jni-lib1",
		host_supported: true,
		srcs: ["jni.cpp"],
	}

	android_robolectric_test {
		name: "robo-test",
		instrumentation_for: "inst-target",
		srcs: ["FooTest.java"],
		jni_libs: [
			"jni-lib1"
		],
	}
	`)

	CheckModuleHasDependency(t, ctx.TestContext, "robo-test", "android_common", "jni-lib1")

	// Check that the .so files make it into the output.
	module := ctx.ModuleForTests("robo-test", "android_common")
	module.Output(installPathPrefix + "/robo-test/lib64/jni-lib1.so")
}
