// Copyright 2019 Google Inc. All rights reserved.
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
	"io"
	"strconv"
	"strings"

	"android/soong/android"
	"android/soong/java/config"
	"android/soong/testing"
	"android/soong/tradefed"

	"github.com/google/blueprint/proptools"
)

func init() {
	android.RegisterModuleType("android_robolectric_test", RobolectricTestFactory)
	android.RegisterModuleType("android_robolectric_runtimes", robolectricRuntimesFactory)
}

var robolectricDefaultLibs = []string{
	"mockito-robolectric-prebuilt",
	"truth",
	// TODO(ccross): this is not needed at link time
	"junitxml",
}

const robolectricCurrentLib = "Robolectric_all-target"
const robolectricPrebuiltLibPattern = "platform-robolectric-%s-prebuilt"

var (
	roboCoverageLibsTag = dependencyTag{name: "roboCoverageLibs"}
	roboRuntimesTag     = dependencyTag{name: "roboRuntimes"}
)

type robolectricProperties struct {
	// The name of the android_app module that the tests will run against.
	Instrumentation_for *string

	// Additional libraries for which coverage data should be generated
	Coverage_libs []string

	Test_options struct {
		// Timeout in seconds when running the tests.
		Timeout *int64

		// Number of shards to use when running the tests.
		Shards *int64
	}

	// The version number of a robolectric prebuilt to use from prebuilts/misc/common/robolectric
	// instead of the one built from source in external/robolectric-shadows.
	Robolectric_prebuilt_version *string

	// Use /external/robolectric rather than /external/robolectric-shadows as the version of robolectric
	// to use.  /external/robolectric closely tracks github's master, and will fully replace /external/robolectric-shadows
	Upstream *bool
}

type robolectricTest struct {
	Library

	robolectricProperties robolectricProperties
	testProperties        testProperties

	libs  []string
	tests []string

	manifest    android.Path
	resourceApk android.Path

	combinedJar android.WritablePath

	roboSrcJar android.Path

	testConfig android.Path
	data       android.Paths

	forceOSType   android.OsType
	forceArchType android.ArchType
}

func (r *robolectricTest) TestSuites() []string {
	return r.testProperties.Test_suites
}

var _ android.TestSuiteModule = (*robolectricTest)(nil)

func (r *robolectricTest) DepsMutator(ctx android.BottomUpMutatorContext) {
	r.Library.DepsMutator(ctx)

	if r.robolectricProperties.Instrumentation_for != nil {
		ctx.AddVariationDependencies(nil, instrumentationForTag, String(r.robolectricProperties.Instrumentation_for))
	} else {
		ctx.PropertyErrorf("instrumentation_for", "missing required instrumented module")
	}

	if v := String(r.robolectricProperties.Robolectric_prebuilt_version); v != "" {
		ctx.AddVariationDependencies(nil, libTag, fmt.Sprintf(robolectricPrebuiltLibPattern, v))
	} else {
		if proptools.Bool(r.robolectricProperties.Upstream) {
			ctx.AddVariationDependencies(nil, libTag, robolectricCurrentLib+"_upstream")
		} else {
			ctx.AddVariationDependencies(nil, libTag, robolectricCurrentLib)
		}
	}

	ctx.AddVariationDependencies(nil, libTag, robolectricDefaultLibs...)

	ctx.AddVariationDependencies(nil, roboCoverageLibsTag, r.robolectricProperties.Coverage_libs...)

	ctx.AddFarVariationDependencies(ctx.Config().BuildOSCommonTarget.Variations(),
		roboRuntimesTag, "robolectric-android-all-prebuilts")
}

func (r *robolectricTest) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	r.forceOSType = ctx.Config().BuildOS
	r.forceArchType = ctx.Config().BuildArch

	r.testConfig = tradefed.AutoGenTestConfig(ctx, tradefed.AutoGenTestConfigOptions{
		TestConfigProp:         r.testProperties.Test_config,
		TestConfigTemplateProp: r.testProperties.Test_config_template,
		TestSuites:             r.testProperties.Test_suites,
		AutoGenConfig:          r.testProperties.Auto_gen_config,
		DeviceTemplate:         "${RobolectricTestConfigTemplate}",
		HostTemplate:           "${RobolectricTestConfigTemplate}",
	})
	r.data = android.PathsForModuleSrc(ctx, r.testProperties.Data)

	roboTestConfig := android.PathForModuleGen(ctx, "robolectric").
		Join(ctx, "com/android/tools/test_config.properties")

	var ok bool
	var instrumentedApp *AndroidApp

	// TODO: this inserts paths to built files into the test, it should really be inserting the contents.
	instrumented := ctx.GetDirectDepsWithTag(instrumentationForTag)

	if len(instrumented) == 1 {
		instrumentedApp, ok = instrumented[0].(*AndroidApp)
		if !ok {
			ctx.PropertyErrorf("instrumentation_for", "dependency must be an android_app")
		}
	} else if !ctx.Config().AllowMissingDependencies() {
		panic(fmt.Errorf("expected exactly 1 instrumented dependency, got %d", len(instrumented)))
	}

	if instrumentedApp != nil {
		r.manifest = instrumentedApp.mergedManifestFile
		r.resourceApk = instrumentedApp.outputFile

		generateRoboTestConfig(ctx, roboTestConfig, instrumentedApp)
		r.extraResources = android.Paths{roboTestConfig}
	}

	r.Library.GenerateAndroidBuildActions(ctx)

	roboSrcJar := android.PathForModuleGen(ctx, "robolectric", ctx.ModuleName()+".srcjar")

	if instrumentedApp != nil {
		r.generateRoboSrcJar(ctx, roboSrcJar, instrumentedApp)
		r.roboSrcJar = roboSrcJar
	}

	roboTestConfigJar := android.PathForModuleOut(ctx, "robolectric_samedir", "samedir_config.jar")
	generateSameDirRoboTestConfigJar(ctx, roboTestConfigJar)

	combinedJarJars := android.Paths{
		// roboTestConfigJar comes first so that its com/android/tools/test_config.properties
		// overrides the one from r.extraResources.  The r.extraResources one can be removed
		// once the Make test runner is removed.
		roboTestConfigJar,
		r.outputFile,
	}

	if instrumentedApp != nil {
		combinedJarJars = append(combinedJarJars, instrumentedApp.implementationAndResourcesJar)
	}

	handleLibDeps := func(dep android.Module) {
		m, _ := android.OtherModuleProvider(ctx, dep, JavaInfoProvider)
		r.libs = append(r.libs, ctx.OtherModuleName(dep))
		if !android.InList(ctx.OtherModuleName(dep), config.FrameworkLibraries) {
			combinedJarJars = append(combinedJarJars, m.ImplementationAndResourcesJars...)
		}
	}

	for _, dep := range ctx.GetDirectDepsWithTag(libTag) {
		handleLibDeps(dep)
	}
	for _, dep := range ctx.GetDirectDepsWithTag(sdkLibTag) {
		handleLibDeps(dep)
	}

	r.combinedJar = android.PathForModuleOut(ctx, "robolectric_combined", r.outputFile.Base())
	TransformJarsToJar(ctx, r.combinedJar, "combine jars", combinedJarJars, android.OptionalPath{},
		false, nil, nil)

	// TODO: this could all be removed if tradefed was used as the test runner, it will find everything
	// annotated as a test and run it.
	for _, src := range r.uniqueSrcFiles {
		s := src.Rel()
		if !strings.HasSuffix(s, "Test.java") && !strings.HasSuffix(s, "Test.kt") {
			continue
		} else if strings.HasSuffix(s, "/BaseRobolectricTest.java") {
			continue
		} else {
			s = strings.TrimPrefix(s, "src/")
		}
		r.tests = append(r.tests, s)
	}

	installPath := android.PathForModuleInstall(ctx, r.BaseModuleName())
	var installDeps android.InstallPaths

	if r.manifest != nil {
		r.data = append(r.data, r.manifest)
		installedManifest := ctx.InstallFile(installPath, ctx.ModuleName()+"-AndroidManifest.xml", r.manifest)
		installDeps = append(installDeps, installedManifest)
	}

	if r.resourceApk != nil {
		r.data = append(r.data, r.resourceApk)
		installedResourceApk := ctx.InstallFile(installPath, ctx.ModuleName()+".apk", r.resourceApk)
		installDeps = append(installDeps, installedResourceApk)
	}

	runtimes := ctx.GetDirectDepWithTag("robolectric-android-all-prebuilts", roboRuntimesTag)
	for _, runtime := range runtimes.(*robolectricRuntimes).runtimes {
		installDeps = append(installDeps, runtime)
	}

	installedConfig := ctx.InstallFile(installPath, ctx.ModuleName()+".config", r.testConfig)
	installDeps = append(installDeps, installedConfig)

	for _, data := range android.PathsForModuleSrc(ctx, r.testProperties.Data) {
		installedData := ctx.InstallFile(installPath, data.Rel(), data)
		installDeps = append(installDeps, installedData)
	}

	r.installFile = ctx.InstallFile(installPath, ctx.ModuleName()+".jar", r.combinedJar, installDeps...)
	android.SetProvider(ctx, testing.TestModuleProviderKey, testing.TestModuleProviderData{})
}

func generateRoboTestConfig(ctx android.ModuleContext, outputFile android.WritablePath,
	instrumentedApp *AndroidApp) {
	rule := android.NewRuleBuilder(pctx, ctx)

	manifest := instrumentedApp.mergedManifestFile
	resourceApk := instrumentedApp.outputFile

	rule.Command().Text("rm -f").Output(outputFile)
	rule.Command().
		Textf(`echo "android_merged_manifest=%s" >>`, manifest.String()).Output(outputFile).Text("&&").
		Textf(`echo "android_resource_apk=%s" >>`, resourceApk.String()).Output(outputFile).
		// Make it depend on the files to which it points so the test file's timestamp is updated whenever the
		// contents change
		Implicit(manifest).
		Implicit(resourceApk)

	rule.Build("generate_test_config", "generate test_config.properties")
}

func generateSameDirRoboTestConfigJar(ctx android.ModuleContext, outputFile android.ModuleOutPath) {
	rule := android.NewRuleBuilder(pctx, ctx)

	outputDir := outputFile.InSameDir(ctx)
	configFile := outputDir.Join(ctx, "com/android/tools/test_config.properties")
	rule.Temporary(configFile)
	rule.Command().Text("rm -f").Output(outputFile).Output(configFile)
	rule.Command().Textf("mkdir -p $(dirname %s)", configFile.String())
	rule.Command().
		Text("(").
		Textf(`echo "android_merged_manifest=%s-AndroidManifest.xml" &&`, ctx.ModuleName()).
		Textf(`echo "android_resource_apk=%s.apk"`, ctx.ModuleName()).
		Text(") >>").Output(configFile)
	rule.Command().
		BuiltTool("soong_zip").
		FlagWithArg("-C ", outputDir.String()).
		FlagWithInput("-f ", configFile).
		FlagWithOutput("-o ", outputFile)

	rule.Build("generate_test_config_samedir", "generate test_config.properties")
}

func (r *robolectricTest) generateRoboSrcJar(ctx android.ModuleContext, outputFile android.WritablePath,
	instrumentedApp *AndroidApp) {

	srcJarArgs := android.CopyOf(instrumentedApp.srcJarArgs)
	srcJarDeps := append(android.Paths(nil), instrumentedApp.srcJarDeps...)

	for _, m := range ctx.GetDirectDepsWithTag(roboCoverageLibsTag) {
		if dep, ok := android.OtherModuleProvider(ctx, m, JavaInfoProvider); ok {
			srcJarArgs = append(srcJarArgs, dep.SrcJarArgs...)
			srcJarDeps = append(srcJarDeps, dep.SrcJarDeps...)
		}
	}

	TransformResourcesToJar(ctx, outputFile, srcJarArgs, srcJarDeps)
}

func (r *robolectricTest) AndroidMkEntries() []android.AndroidMkEntries {
	entriesList := r.Library.AndroidMkEntries()
	entries := &entriesList[0]
	entries.ExtraEntries = append(entries.ExtraEntries,
		func(ctx android.AndroidMkExtraEntriesContext, entries *android.AndroidMkEntries) {
			entries.SetBool("LOCAL_UNINSTALLABLE_MODULE", true)
			entries.AddStrings("LOCAL_COMPATIBILITY_SUITE", "robolectric-tests")
			if r.testConfig != nil {
				entries.SetPath("LOCAL_FULL_TEST_CONFIG", r.testConfig)
			}
		})

	entries.ExtraFooters = []android.AndroidMkExtraFootersFunc{
		func(w io.Writer, name, prefix, moduleDir string) {
			if s := r.robolectricProperties.Test_options.Shards; s != nil && *s > 1 {
				numShards := int(*s)
				shardSize := (len(r.tests) + numShards - 1) / numShards
				shards := android.ShardStrings(r.tests, shardSize)
				for i, shard := range shards {
					r.writeTestRunner(w, name, "Run"+name+strconv.Itoa(i), shard)
				}

				// TODO: add rules to dist the outputs of the individual tests, or combine them together?
				fmt.Fprintln(w, "")
				fmt.Fprintln(w, ".PHONY:", "Run"+name)
				fmt.Fprintln(w, "Run"+name, ": \\")
				for i := range shards {
					fmt.Fprintln(w, "   ", "Run"+name+strconv.Itoa(i), "\\")
				}
				fmt.Fprintln(w, "")
			} else {
				r.writeTestRunner(w, name, "Run"+name, r.tests)
			}
		},
	}

	return entriesList
}

func (r *robolectricTest) writeTestRunner(w io.Writer, module, name string, tests []string) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "include $(CLEAR_VARS)", " # java.robolectricTest")
	fmt.Fprintln(w, "LOCAL_MODULE :=", name)
	android.AndroidMkEmitAssignList(w, "LOCAL_JAVA_LIBRARIES", []string{module}, r.libs)
	fmt.Fprintln(w, "LOCAL_TEST_PACKAGE :=", String(r.robolectricProperties.Instrumentation_for))
	if r.roboSrcJar != nil {
		fmt.Fprintln(w, "LOCAL_INSTRUMENT_SRCJARS :=", r.roboSrcJar.String())
	}
	android.AndroidMkEmitAssignList(w, "LOCAL_ROBOTEST_FILES", tests)
	if t := r.robolectricProperties.Test_options.Timeout; t != nil {
		fmt.Fprintln(w, "LOCAL_ROBOTEST_TIMEOUT :=", *t)
	}
	if v := String(r.robolectricProperties.Robolectric_prebuilt_version); v != "" {
		fmt.Fprintf(w, "-include prebuilts/misc/common/robolectric/%s/run_robotests.mk\n", v)
	} else {
		fmt.Fprintln(w, "-include external/robolectric-shadows/run_robotests.mk")
	}
}

// An android_robolectric_test module compiles tests against the Robolectric framework that can run on the local host
// instead of on a device.  It also generates a rule with the name of the module prefixed with "Run" that can be
// used to run the tests.  Running the tests with build rule will eventually be deprecated and replaced with atest.
//
// The test runner considers any file listed in srcs whose name ends with Test.java to be a test class, unless
// it is named BaseRobolectricTest.java.  The path to the each source file must exactly match the package
// name, or match the package name when the prefix "src/" is removed.
func RobolectricTestFactory() android.Module {
	module := &robolectricTest{}

	module.addHostProperties()
	module.AddProperties(
		&module.Module.deviceProperties,
		&module.robolectricProperties,
		&module.testProperties)

	module.Module.dexpreopter.isTest = true
	module.Module.linter.properties.Lint.Test = proptools.BoolPtr(true)

	module.testProperties.Test_suites = []string{"robolectric-tests"}

	InitJavaModule(module, android.DeviceSupported)
	return module
}

func (r *robolectricTest) InstallInTestcases() bool { return true }
func (r *robolectricTest) InstallForceOS() (*android.OsType, *android.ArchType) {
	return &r.forceOSType, &r.forceArchType
}

func robolectricRuntimesFactory() android.Module {
	module := &robolectricRuntimes{}
	module.AddProperties(&module.props)
	android.InitAndroidArchModule(module, android.HostSupportedNoCross, android.MultilibCommon)
	return module
}

type robolectricRuntimesProperties struct {
	Jars []string `android:"path"`
	Lib  *string
}

type robolectricRuntimes struct {
	android.ModuleBase

	props robolectricRuntimesProperties

	runtimes []android.InstallPath

	forceOSType   android.OsType
	forceArchType android.ArchType
}

func (r *robolectricRuntimes) TestSuites() []string {
	return []string{"robolectric-tests"}
}

var _ android.TestSuiteModule = (*robolectricRuntimes)(nil)

func (r *robolectricRuntimes) DepsMutator(ctx android.BottomUpMutatorContext) {
	if !ctx.Config().AlwaysUsePrebuiltSdks() && r.props.Lib != nil {
		ctx.AddVariationDependencies(nil, libTag, String(r.props.Lib))
	}
}

func (r *robolectricRuntimes) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	if ctx.Target().Os != ctx.Config().BuildOSCommonTarget.Os {
		return
	}

	r.forceOSType = ctx.Config().BuildOS
	r.forceArchType = ctx.Config().BuildArch

	files := android.PathsForModuleSrc(ctx, r.props.Jars)

	androidAllDir := android.PathForModuleInstall(ctx, "android-all")
	for _, from := range files {
		installedRuntime := ctx.InstallFile(androidAllDir, from.Base(), from)
		r.runtimes = append(r.runtimes, installedRuntime)
	}

	if !ctx.Config().AlwaysUsePrebuiltSdks() && r.props.Lib != nil {
		runtimeFromSourceModule := ctx.GetDirectDepWithTag(String(r.props.Lib), libTag)
		if runtimeFromSourceModule == nil {
			if ctx.Config().AllowMissingDependencies() {
				ctx.AddMissingDependencies([]string{String(r.props.Lib)})
			} else {
				ctx.PropertyErrorf("lib", "missing dependency %q", String(r.props.Lib))
			}
			return
		}
		runtimeFromSourceJar := android.OutputFileForModule(ctx, runtimeFromSourceModule, "")

		// "TREE" name is essential here because it hooks into the "TREE" name in
		// Robolectric's SdkConfig.java that will always correspond to the NEWEST_SDK
		// in Robolectric configs.
		runtimeName := "android-all-current-robolectric-r0.jar"
		installedRuntime := ctx.InstallFile(androidAllDir, runtimeName, runtimeFromSourceJar)
		r.runtimes = append(r.runtimes, installedRuntime)
	}
}

func (r *robolectricRuntimes) InstallInTestcases() bool { return true }
func (r *robolectricRuntimes) InstallForceOS() (*android.OsType, *android.ArchType) {
	return &r.forceOSType, &r.forceArchType
}
