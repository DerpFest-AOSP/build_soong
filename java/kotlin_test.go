// Copyright 2017 Google Inc. All rights reserved.
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
	"slices"
	"strconv"
	"strings"
	"testing"

	"android/soong/android"
)

func TestKotlin(t *testing.T) {
	bp := `
		java_library {
			name: "foo",
			srcs: ["a.java", "b.kt"],
			static_libs: ["quz"],
		}

		java_library {
			name: "bar",
			srcs: ["b.kt"],
			libs: ["foo"],
			static_libs: ["baz"],
		}

		java_library {
			name: "baz",
			srcs: ["c.java"],
			static_libs: ["quz"],
		}

		java_library {
			name: "quz",
			srcs: ["d.kt"],
		}`

	kotlinStdlibTurbineCombinedJars := []string{
		"out/soong/.intermediates/default/java/kotlin-stdlib/android_common/turbine-combined/kotlin-stdlib.jar",
		"out/soong/.intermediates/default/java/kotlin-stdlib-jdk7/android_common/turbine-combined/kotlin-stdlib-jdk7.jar",
		"out/soong/.intermediates/default/java/kotlin-stdlib-jdk8/android_common/turbine-combined/kotlin-stdlib-jdk8.jar",
		"out/soong/.intermediates/default/java/kotlin-annotations/android_common/turbine-combined/kotlin-annotations.jar",
	}

	kotlinStdlibTurbineJars := []string{
		"out/soong/.intermediates/default/java/kotlin-stdlib/android_common/turbine/kotlin-stdlib.jar",
		"out/soong/.intermediates/default/java/kotlin-stdlib-jdk7/android_common/turbine/kotlin-stdlib-jdk7.jar",
		"out/soong/.intermediates/default/java/kotlin-stdlib-jdk8/android_common/turbine/kotlin-stdlib-jdk8.jar",
		"out/soong/.intermediates/default/java/kotlin-annotations/android_common/turbine/kotlin-annotations.jar",
	}

	kotlinStdlibJavacJars := []string{
		"out/soong/.intermediates/default/java/kotlin-stdlib/android_common/javac/kotlin-stdlib.jar",
		"out/soong/.intermediates/default/java/kotlin-stdlib-jdk7/android_common/javac/kotlin-stdlib-jdk7.jar",
		"out/soong/.intermediates/default/java/kotlin-stdlib-jdk8/android_common/javac/kotlin-stdlib-jdk8.jar",
		"out/soong/.intermediates/default/java/kotlin-annotations/android_common/javac/kotlin-annotations.jar",
	}

	bootclasspathTurbineCombinedJars := []string{
		"out/soong/.intermediates/default/java/stable.core.platform.api.stubs/android_common/turbine-combined/stable.core.platform.api.stubs.jar",
		"out/soong/.intermediates/default/java/core-lambda-stubs/android_common/turbine-combined/core-lambda-stubs.jar",
	}

	bootclasspathTurbineJars := []string{
		"out/soong/.intermediates/default/java/stable.core.platform.api.stubs/android_common/turbine/stable.core.platform.api.stubs.jar",
		"out/soong/.intermediates/default/java/core-lambda-stubs/android_common/turbine/core-lambda-stubs.jar",
	}

	frameworkTurbineCombinedJars := []string{
		"out/soong/.intermediates/default/java/ext/android_common/turbine-combined/ext.jar",
		"out/soong/.intermediates/default/java/framework/android_common/turbine-combined/framework.jar",
	}

	frameworkTurbineJars := []string{
		"out/soong/.intermediates/default/java/ext/android_common/turbine/ext.jar",
		"out/soong/.intermediates/default/java/framework/android_common/turbine/framework.jar",
	}

	testCases := []struct {
		name string

		preparer android.FixturePreparer

		fooKotlincInputs        []string
		fooJavacInputs          []string
		fooKotlincClasspath     []string
		fooJavacClasspath       []string
		fooCombinedInputs       []string
		fooHeaderCombinedInputs []string

		barKotlincInputs        []string
		barKotlincClasspath     []string
		barCombinedInputs       []string
		barHeaderCombinedInputs []string
	}{
		{
			name:             "normal",
			preparer:         android.NullFixturePreparer,
			fooKotlincInputs: []string{"a.java", "b.kt"},
			fooJavacInputs:   []string{"a.java"},
			fooKotlincClasspath: slices.Concat(
				bootclasspathTurbineCombinedJars,
				frameworkTurbineCombinedJars,
				[]string{"out/soong/.intermediates/quz/android_common/turbine-combined/quz.jar"},
				kotlinStdlibTurbineCombinedJars,
			),
			fooJavacClasspath: slices.Concat(
				[]string{"out/soong/.intermediates/foo/android_common/kotlin_headers/foo.jar"},
				frameworkTurbineCombinedJars,
				[]string{"out/soong/.intermediates/quz/android_common/turbine-combined/quz.jar"},
				kotlinStdlibTurbineCombinedJars,
			),
			fooCombinedInputs: slices.Concat(
				[]string{
					"out/soong/.intermediates/foo/android_common/kotlin/foo.jar",
					"out/soong/.intermediates/foo/android_common/javac/foo.jar",
					"out/soong/.intermediates/quz/android_common/combined/quz.jar",
				},
				kotlinStdlibJavacJars,
			),
			fooHeaderCombinedInputs: slices.Concat(
				[]string{
					"out/soong/.intermediates/foo/android_common/turbine/foo.jar",
					"out/soong/.intermediates/foo/android_common/kotlin_headers/foo.jar",
					"out/soong/.intermediates/quz/android_common/turbine-combined/quz.jar",
				},
				kotlinStdlibTurbineCombinedJars,
			),

			barKotlincInputs: []string{"b.kt"},
			barKotlincClasspath: slices.Concat(
				bootclasspathTurbineCombinedJars,
				frameworkTurbineCombinedJars,
				[]string{
					"out/soong/.intermediates/foo/android_common/turbine-combined/foo.jar",
					"out/soong/.intermediates/baz/android_common/turbine-combined/baz.jar",
				},
				kotlinStdlibTurbineCombinedJars,
			),
			barCombinedInputs: slices.Concat(
				[]string{
					"out/soong/.intermediates/bar/android_common/kotlin/bar.jar",
					"out/soong/.intermediates/baz/android_common/combined/baz.jar",
				},
				kotlinStdlibJavacJars,
				[]string{},
			),
			barHeaderCombinedInputs: slices.Concat(
				[]string{
					"out/soong/.intermediates/bar/android_common/kotlin_headers/bar.jar",
					"out/soong/.intermediates/baz/android_common/turbine-combined/baz.jar",
				},
				kotlinStdlibTurbineCombinedJars,
			),
		},
		{
			name:             "transitive classpath",
			preparer:         PrepareForTestWithTransitiveClasspathEnabled,
			fooKotlincInputs: []string{"a.java", "b.kt"},
			fooJavacInputs:   []string{"a.java"},
			fooKotlincClasspath: slices.Concat(
				bootclasspathTurbineJars,
				frameworkTurbineJars,
				[]string{"out/soong/.intermediates/quz/android_common/kotlin_headers/quz.jar"},
				kotlinStdlibTurbineJars,
			),
			fooJavacClasspath: slices.Concat(
				[]string{"out/soong/.intermediates/foo/android_common/kotlin_headers/foo.jar"},
				frameworkTurbineJars,
				[]string{"out/soong/.intermediates/quz/android_common/kotlin_headers/quz.jar"},
				kotlinStdlibTurbineJars,
			),
			fooCombinedInputs: slices.Concat(
				[]string{
					"out/soong/.intermediates/foo/android_common/kotlin/foo.jar",
					"out/soong/.intermediates/foo/android_common/javac/foo.jar",
					"out/soong/.intermediates/quz/android_common/kotlin/quz.jar",
				},
				kotlinStdlibJavacJars,
			),
			fooHeaderCombinedInputs: slices.Concat(
				[]string{
					"out/soong/.intermediates/foo/android_common/turbine/foo.jar",
					"out/soong/.intermediates/foo/android_common/kotlin_headers/foo.jar",
					"out/soong/.intermediates/quz/android_common/kotlin_headers/quz.jar",
				},
				kotlinStdlibTurbineJars,
			),

			barKotlincInputs: []string{"b.kt"},
			barKotlincClasspath: slices.Concat(
				bootclasspathTurbineJars,
				frameworkTurbineJars,
				[]string{
					"out/soong/.intermediates/foo/android_common/turbine/foo.jar",
					"out/soong/.intermediates/foo/android_common/kotlin_headers/foo.jar",
					"out/soong/.intermediates/quz/android_common/kotlin_headers/quz.jar",
				},
				kotlinStdlibTurbineJars,
				[]string{"out/soong/.intermediates/baz/android_common/turbine/baz.jar"},
			),
			barCombinedInputs: slices.Concat(
				[]string{
					"out/soong/.intermediates/bar/android_common/kotlin/bar.jar",
					"out/soong/.intermediates/baz/android_common/javac/baz.jar",
					"out/soong/.intermediates/quz/android_common/kotlin/quz.jar",
				},
				kotlinStdlibJavacJars,
			),
			barHeaderCombinedInputs: slices.Concat(
				[]string{
					"out/soong/.intermediates/bar/android_common/kotlin_headers/bar.jar",
					"out/soong/.intermediates/baz/android_common/turbine/baz.jar",
					"out/soong/.intermediates/quz/android_common/kotlin_headers/quz.jar",
				},
				kotlinStdlibTurbineJars,
			),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := android.GroupFixturePreparers(
				PrepareForTestWithJavaDefaultModules,
				tt.preparer,
			).RunTestWithBp(t, bp)
			foo := result.ModuleForTests("foo", "android_common")
			fooKotlinc := foo.Rule("kotlinc")
			android.AssertPathsRelativeToTopEquals(t, "foo kotlinc inputs", tt.fooKotlincInputs, fooKotlinc.Inputs)

			fooKotlincClasspath := android.ContentFromFileRuleForTests(t, result.TestContext, foo.Output("kotlinc/classpath.rsp"))
			android.AssertStringPathsRelativeToTopEquals(t, "foo kotlinc classpath", result.Config, tt.fooKotlincClasspath, strings.Fields(fooKotlincClasspath))

			fooJavac := foo.Rule("javac")
			android.AssertPathsRelativeToTopEquals(t, "foo javac inputs", tt.fooJavacInputs, fooJavac.Inputs)

			fooJavacClasspath := fooJavac.Args["classpath"]
			android.AssertStringPathsRelativeToTopEquals(t, "foo javac classpath", result.Config, tt.fooJavacClasspath,
				strings.Split(strings.TrimPrefix(fooJavacClasspath, "-classpath "), ":"))

			fooCombinedJar := foo.Output("combined/foo.jar")
			android.AssertPathsRelativeToTopEquals(t, "foo combined inputs", tt.fooCombinedInputs, fooCombinedJar.Inputs)

			fooCombinedHeaderJar := foo.Output("turbine-combined/foo.jar")
			android.AssertPathsRelativeToTopEquals(t, "foo header combined inputs", tt.fooHeaderCombinedInputs, fooCombinedHeaderJar.Inputs)

			bar := result.ModuleForTests("bar", "android_common")
			barKotlinc := bar.Rule("kotlinc")
			android.AssertPathsRelativeToTopEquals(t, "bar kotlinc inputs", tt.barKotlincInputs, barKotlinc.Inputs)

			barKotlincClasspath := android.ContentFromFileRuleForTests(t, result.TestContext, bar.Output("kotlinc/classpath.rsp"))
			android.AssertStringPathsRelativeToTopEquals(t, "bar kotlinc classpath", result.Config, tt.barKotlincClasspath, strings.Fields(barKotlincClasspath))

			barCombinedJar := bar.Output("combined/bar.jar")
			android.AssertPathsRelativeToTopEquals(t, "bar combined inputs", tt.barCombinedInputs, barCombinedJar.Inputs)

			barCombinedHeaderJar := bar.Output("turbine-combined/bar.jar")
			android.AssertPathsRelativeToTopEquals(t, "bar header combined inputs", tt.barHeaderCombinedInputs, barCombinedHeaderJar.Inputs)
		})
	}
}

func TestKapt(t *testing.T) {
	bp := `
		java_library {
			name: "foo",
			srcs: ["a.java", "b.kt"],
			plugins: ["bar", "baz"],
			errorprone: {
				extra_check_modules: ["my_check"],
			},
		}

		java_plugin {
			name: "bar",
			processor_class: "com.bar",
			srcs: ["b.java"],
		}

		java_plugin {
			name: "baz",
			processor_class: "com.baz",
			srcs: ["b.java"],
		}

		java_plugin {
			name: "my_check",
			srcs: ["b.java"],
		}
	`
	t.Run("", func(t *testing.T) {
		ctx, _ := testJava(t, bp)

		buildOS := ctx.Config().BuildOS.String()

		foo := ctx.ModuleForTests("foo", "android_common")
		kaptStubs := foo.Rule("kapt")
		turbineApt := foo.Description("turbine apt")
		kotlinc := foo.Rule("kotlinc")
		javac := foo.Rule("javac")

		bar := ctx.ModuleForTests("bar", buildOS+"_common").Rule("javac").Output.String()
		baz := ctx.ModuleForTests("baz", buildOS+"_common").Rule("javac").Output.String()

		// Test that the kotlin and java sources are passed to kapt and kotlinc
		if len(kaptStubs.Inputs) != 2 || kaptStubs.Inputs[0].String() != "a.java" || kaptStubs.Inputs[1].String() != "b.kt" {
			t.Errorf(`foo kapt inputs %v != ["a.java", "b.kt"]`, kaptStubs.Inputs)
		}
		if len(kotlinc.Inputs) != 2 || kotlinc.Inputs[0].String() != "a.java" || kotlinc.Inputs[1].String() != "b.kt" {
			t.Errorf(`foo kotlinc inputs %v != ["a.java", "b.kt"]`, kotlinc.Inputs)
		}

		// Test that only the java sources are passed to turbine-apt and javac
		if len(turbineApt.Inputs) != 1 || turbineApt.Inputs[0].String() != "a.java" {
			t.Errorf(`foo turbine apt inputs %v != ["a.java"]`, turbineApt.Inputs)
		}
		if len(javac.Inputs) != 1 || javac.Inputs[0].String() != "a.java" {
			t.Errorf(`foo inputs %v != ["a.java"]`, javac.Inputs)
		}

		// Test that the kapt stubs jar is a dependency of turbine-apt
		if !inList(kaptStubs.Output.String(), turbineApt.Implicits.Strings()) {
			t.Errorf("expected %q in turbine-apt implicits %v", kaptStubs.Output.String(), kotlinc.Implicits.Strings())
		}

		// Test that the turbine-apt srcjar is a dependency of kotlinc and javac rules
		if !inList(turbineApt.Output.String(), kotlinc.Implicits.Strings()) {
			t.Errorf("expected %q in kotlinc implicits %v", turbineApt.Output.String(), kotlinc.Implicits.Strings())
		}
		if !inList(turbineApt.Output.String(), javac.Implicits.Strings()) {
			t.Errorf("expected %q in javac implicits %v", turbineApt.Output.String(), javac.Implicits.Strings())
		}

		// Test that the turbine-apt srcjar is extracted by the kotlinc and javac rules
		if kotlinc.Args["srcJars"] != turbineApt.Output.String() {
			t.Errorf("expected %q in kotlinc srcjars %v", turbineApt.Output.String(), kotlinc.Args["srcJars"])
		}
		if javac.Args["srcJars"] != turbineApt.Output.String() {
			t.Errorf("expected %q in javac srcjars %v", turbineApt.Output.String(), kotlinc.Args["srcJars"])
		}

		// Test that the processors are passed to kapt
		expectedProcessorPath := "-P plugin:org.jetbrains.kotlin.kapt3:apclasspath=" + bar +
			" -P plugin:org.jetbrains.kotlin.kapt3:apclasspath=" + baz
		if kaptStubs.Args["kaptProcessorPath"] != expectedProcessorPath {
			t.Errorf("expected kaptProcessorPath %q, got %q", expectedProcessorPath, kaptStubs.Args["kaptProcessorPath"])
		}
		expectedProcessor := "-P plugin:org.jetbrains.kotlin.kapt3:processors=com.bar -P plugin:org.jetbrains.kotlin.kapt3:processors=com.baz"
		if kaptStubs.Args["kaptProcessor"] != expectedProcessor {
			t.Errorf("expected kaptProcessor %q, got %q", expectedProcessor, kaptStubs.Args["kaptProcessor"])
		}

		// Test that the processors are passed to turbine-apt
		expectedProcessorPath = "--processorpath " + bar + " " + baz
		if !strings.Contains(turbineApt.Args["turbineFlags"], expectedProcessorPath) {
			t.Errorf("expected turbine-apt processorpath %q, got %q", expectedProcessorPath, turbineApt.Args["turbineFlags"])
		}
		expectedProcessor = "--processors com.bar com.baz"
		if !strings.Contains(turbineApt.Args["turbineFlags"], expectedProcessor) {
			t.Errorf("expected turbine-apt processor %q, got %q", expectedProcessor, turbineApt.Args["turbineFlags"])
		}

		// Test that the processors are not passed to javac
		if javac.Args["processorpath"] != "" {
			t.Errorf("expected processorPath '', got %q", javac.Args["processorpath"])
		}
		if javac.Args["processor"] != "-proc:none" {
			t.Errorf("expected processor '-proc:none', got %q", javac.Args["processor"])
		}
	})

	t.Run("errorprone", func(t *testing.T) {
		env := map[string]string{
			"RUN_ERROR_PRONE": "true",
		}

		result := android.GroupFixturePreparers(
			PrepareForTestWithJavaDefaultModules,
			android.FixtureMergeEnv(env),
		).RunTestWithBp(t, bp)

		buildOS := result.Config.BuildOS.String()

		kapt := result.ModuleForTests("foo", "android_common").Rule("kapt")
		javac := result.ModuleForTests("foo", "android_common").Description("javac")
		errorprone := result.ModuleForTests("foo", "android_common").Description("errorprone")

		bar := result.ModuleForTests("bar", buildOS+"_common").Description("javac").Output.String()
		baz := result.ModuleForTests("baz", buildOS+"_common").Description("javac").Output.String()
		myCheck := result.ModuleForTests("my_check", buildOS+"_common").Description("javac").Output.String()

		// Test that the errorprone plugins are not passed to kapt
		expectedProcessorPath := "-P plugin:org.jetbrains.kotlin.kapt3:apclasspath=" + bar +
			" -P plugin:org.jetbrains.kotlin.kapt3:apclasspath=" + baz
		if kapt.Args["kaptProcessorPath"] != expectedProcessorPath {
			t.Errorf("expected kaptProcessorPath %q, got %q", expectedProcessorPath, kapt.Args["kaptProcessorPath"])
		}
		expectedProcessor := "-P plugin:org.jetbrains.kotlin.kapt3:processors=com.bar -P plugin:org.jetbrains.kotlin.kapt3:processors=com.baz"
		if kapt.Args["kaptProcessor"] != expectedProcessor {
			t.Errorf("expected kaptProcessor %q, got %q", expectedProcessor, kapt.Args["kaptProcessor"])
		}

		// Test that the errorprone plugins are not passed to javac
		if javac.Args["processorpath"] != "" {
			t.Errorf("expected processorPath '', got %q", javac.Args["processorpath"])
		}
		if javac.Args["processor"] != "-proc:none" {
			t.Errorf("expected processor '-proc:none', got %q", javac.Args["processor"])
		}

		// Test that the errorprone plugins are passed to errorprone
		expectedProcessorPath = "-processorpath " + myCheck
		if errorprone.Args["processorpath"] != expectedProcessorPath {
			t.Errorf("expected processorpath %q, got %q", expectedProcessorPath, errorprone.Args["processorpath"])
		}
		if errorprone.Args["processor"] != "-proc:none" {
			t.Errorf("expected processor '-proc:none', got %q", errorprone.Args["processor"])
		}
	})
}

func TestKaptEncodeFlags(t *testing.T) {
	// Compares the kaptEncodeFlags against the results of the example implementation at
	// https://kotlinlang.org/docs/reference/kapt.html#apjavac-options-encoding
	tests := []struct {
		in  [][2]string
		out string
	}{
		{
			// empty input
			in:  [][2]string{},
			out: "rO0ABXcEAAAAAA==",
		},
		{
			// common input
			in: [][2]string{
				{"-source", "1.8"},
				{"-target", "1.8"},
			},
			out: "rO0ABXcgAAAAAgAHLXNvdXJjZQADMS44AActdGFyZ2V0AAMxLjg=",
		},
		{
			// input that serializes to a 255 byte block
			in: [][2]string{
				{"-source", "1.8"},
				{"-target", "1.8"},
				{"a", strings.Repeat("b", 218)},
			},
			out: "rO0ABXf/AAAAAwAHLXNvdXJjZQADMS44AActdGFyZ2V0AAMxLjgAAWEA2mJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJi",
		},
		{
			// input that serializes to a 256 byte block
			in: [][2]string{
				{"-source", "1.8"},
				{"-target", "1.8"},
				{"a", strings.Repeat("b", 219)},
			},
			out: "rO0ABXoAAAEAAAAAAwAHLXNvdXJjZQADMS44AActdGFyZ2V0AAMxLjgAAWEA22JiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYg==",
		},
		{
			// input that serializes to a 257 byte block
			in: [][2]string{
				{"-source", "1.8"},
				{"-target", "1.8"},
				{"a", strings.Repeat("b", 220)},
			},
			out: "rO0ABXoAAAEBAAAAAwAHLXNvdXJjZQADMS44AActdGFyZ2V0AAMxLjgAAWEA3GJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmI=",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := kaptEncodeFlags(test.in)
			if got != test.out {
				t.Errorf("\nwant %q\n got %q", test.out, got)
			}
		})
	}
}

func TestKotlinCompose(t *testing.T) {
	result := android.GroupFixturePreparers(
		PrepareForTestWithJavaDefaultModules,
	).RunTestWithBp(t, `
		java_library {
			name: "androidx.compose.runtime_runtime",
		}

		kotlin_plugin {
			name: "androidx.compose.compiler_compiler-hosted-plugin",
		}

		java_library {
			name: "withcompose",
			srcs: ["a.kt"],
			plugins: ["plugin"],
			static_libs: ["androidx.compose.runtime_runtime"],
		}

		java_library {
			name: "nocompose",
			srcs: ["a.kt"],
		}

		java_plugin {
			name: "plugin",
		}
	`)

	buildOS := result.Config.BuildOS.String()

	composeCompiler := result.ModuleForTests("androidx.compose.compiler_compiler-hosted-plugin", buildOS+"_common").Rule("combineJar").Output
	withCompose := result.ModuleForTests("withcompose", "android_common")
	noCompose := result.ModuleForTests("nocompose", "android_common")

	android.AssertStringListContains(t, "missing compose compiler dependency",
		withCompose.Rule("kotlinc").Implicits.Strings(), composeCompiler.String())

	android.AssertStringDoesContain(t, "missing compose compiler plugin",
		withCompose.VariablesForTestsRelativeToTop()["kotlincFlags"], "-Xplugin="+composeCompiler.String())

	android.AssertStringListContains(t, "missing kapt compose compiler dependency",
		withCompose.Rule("kapt").Implicits.Strings(), composeCompiler.String())

	android.AssertStringListDoesNotContain(t, "unexpected compose compiler dependency",
		noCompose.Rule("kotlinc").Implicits.Strings(), composeCompiler.String())

	android.AssertStringDoesNotContain(t, "unexpected compose compiler plugin",
		noCompose.VariablesForTestsRelativeToTop()["kotlincFlags"], "-Xplugin="+composeCompiler.String())
}

func TestKotlinPlugin(t *testing.T) {
	result := android.GroupFixturePreparers(
		PrepareForTestWithJavaDefaultModules,
	).RunTestWithBp(t, `
		kotlin_plugin {
			name: "kotlin_plugin",
		}

		java_library {
			name: "with_kotlin_plugin",
			srcs: ["a.kt"],
			plugins: ["plugin"],
			kotlin_plugins: ["kotlin_plugin"],
		}

		java_library {
			name: "no_kotlin_plugin",
			srcs: ["a.kt"],
		}

		java_plugin {
			name: "plugin",
		}
	`)

	buildOS := result.Config.BuildOS.String()

	kotlinPlugin := result.ModuleForTests("kotlin_plugin", buildOS+"_common").Rule("combineJar").Output
	withKotlinPlugin := result.ModuleForTests("with_kotlin_plugin", "android_common")
	noKotlinPlugin := result.ModuleForTests("no_kotlin_plugin", "android_common")

	android.AssertStringListContains(t, "missing plugin compiler dependency",
		withKotlinPlugin.Rule("kotlinc").Implicits.Strings(), kotlinPlugin.String())

	android.AssertStringDoesContain(t, "missing kotlin plugin",
		withKotlinPlugin.VariablesForTestsRelativeToTop()["kotlincFlags"], "-Xplugin="+kotlinPlugin.String())

	android.AssertStringListContains(t, "missing kapt kotlin plugin dependency",
		withKotlinPlugin.Rule("kapt").Implicits.Strings(), kotlinPlugin.String())

	android.AssertStringListDoesNotContain(t, "unexpected kotlin plugin dependency",
		noKotlinPlugin.Rule("kotlinc").Implicits.Strings(), kotlinPlugin.String())

	android.AssertStringDoesNotContain(t, "unexpected kotlin plugin",
		noKotlinPlugin.VariablesForTestsRelativeToTop()["kotlincFlags"], "-Xplugin="+kotlinPlugin.String())
}
