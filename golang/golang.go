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

// Package golang wraps the blueprint blueprint_go_binary and bootstrap_go_binary module types in versions
// that implement android.Module that are used when building in Soong.  This simplifies the code in Soong
// so it can always assume modules are an android.Module.
// The original blueprint blueprint_go_binary and bootstrap_go_binary module types are still used during
// bootstrapping, so the Android.bp entries for these module types must be compatible with both the
// original blueprint module types and these wrapped module types.
package golang

import (
	"android/soong/android"
	"github.com/google/blueprint"
	"github.com/google/blueprint/bootstrap"
)

func init() {
	// Wrap the blueprint Go module types with Soong ones that interoperate with the rest of the Soong modules.
	bootstrap.GoModuleTypesAreWrapped()
	RegisterGoModuleTypes(android.InitRegistrationContext)
}

func RegisterGoModuleTypes(ctx android.RegistrationContext) {
	ctx.RegisterModuleType("bootstrap_go_package", goPackageModuleFactory)
	ctx.RegisterModuleType("blueprint_go_binary", goBinaryModuleFactory)
}

// A GoPackage is a module for building Go packages.
type GoPackage struct {
	android.ModuleBase
	bootstrap.GoPackage
}

func goPackageModuleFactory() android.Module {
	module := &GoPackage{}
	module.AddProperties(module.Properties()...)
	android.InitAndroidArchModule(module, android.HostSupported, android.MultilibFirst)
	return module
}

func (g *GoPackage) GenerateBuildActions(ctx blueprint.ModuleContext) {
	// The embedded ModuleBase and bootstrap.GoPackage each implement GenerateBuildActions,
	// the delegation has to be implemented manually to disambiguate.  Call ModuleBase's
	// GenerateBuildActions, which will call GenerateAndroidBuildActions, which will call
	// bootstrap.GoPackage.GenerateBuildActions.
	g.ModuleBase.GenerateBuildActions(ctx)
}

func (g *GoPackage) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	g.GoPackage.GenerateBuildActions(ctx.BlueprintModuleContext())
}

// A GoBinary is a module for building executable binaries from Go sources.
type GoBinary struct {
	android.ModuleBase
	bootstrap.GoBinary

	outputFile android.Path
}

func goBinaryModuleFactory() android.Module {
	module := &GoBinary{}
	module.AddProperties(module.Properties()...)
	android.InitAndroidArchModule(module, android.HostSupportedNoCross, android.MultilibFirst)
	return module
}

func (g *GoBinary) GenerateBuildActions(ctx blueprint.ModuleContext) {
	// The embedded ModuleBase and bootstrap.GoBinary each implement GenerateBuildActions,
	// the delegation has to be implemented manually to disambiguate.  Call ModuleBase's
	// GenerateBuildActions, which will call GenerateAndroidBuildActions, which will call
	// bootstrap.GoBinary.GenerateBuildActions.
	g.ModuleBase.GenerateBuildActions(ctx)
}

func (g *GoBinary) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	// Install the file in Soong instead of blueprint so that Soong knows about the install rules.
	g.GoBinary.SetSkipInstall()

	// Run the build actions from the wrapped blueprint bootstrap module.
	g.GoBinary.GenerateBuildActions(ctx.BlueprintModuleContext())

	// Translate the bootstrap module's string path into a Path
	outputFile := android.PathForArbitraryOutput(ctx, android.Rel(ctx, ctx.Config().OutDir(), g.IntermediateFile())).WithoutRel()
	g.outputFile = outputFile

	// Don't create install rules for modules used by bootstrap, the install command line will differ from
	// what was used during bootstrap, which will cause ninja to rebuild the module on the next run,
	// triggering reanalysis.
	if !usedByBootstrap(ctx.ModuleName()) {
		installPath := ctx.InstallFile(android.PathForModuleInstall(ctx, "bin"), ctx.ModuleName(), outputFile)

		// Modules in an unexported namespace have no install rule, only add modules in the exported namespaces
		// to the blueprint_tools phony rules.
		if !ctx.Config().KatiEnabled() || g.ExportedToMake() {
			ctx.Phony("blueprint_tools", installPath)
		}
	}

	ctx.SetOutputFiles(android.Paths{outputFile}, "")
}

func usedByBootstrap(name string) bool {
	switch name {
	case "loadplugins", "soong_build":
		return true
	default:
		return false
	}
}

func (g *GoBinary) HostToolPath() android.OptionalPath {
	return android.OptionalPathForPath(g.outputFile)
}

func (g *GoBinary) AndroidMkEntries() []android.AndroidMkEntries {
	return []android.AndroidMkEntries{
		{
			Class:      "EXECUTABLES",
			OutputFile: android.OptionalPathForPath(g.outputFile),
			Include:    "$(BUILD_SYSTEM)/soong_cc_rust_prebuilt.mk",
		},
	}
}
