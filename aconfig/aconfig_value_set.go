// Copyright 2023 Google Inc. All rights reserved.
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
	"android/soong/android"
	"fmt"
	"strings"

	"github.com/google/blueprint"
)

// Properties for "aconfig_value_set"
type ValueSetModule struct {
	android.ModuleBase
	android.DefaultableModuleBase

	properties struct {
		// aconfig_values modules
		Values []string

		// Paths to the Android.bp files where the aconfig_values modules are defined.
		Srcs []string
	}
}

func ValueSetFactory() android.Module {
	module := &ValueSetModule{}

	android.InitAndroidModule(module)
	android.InitDefaultableModule(module)
	module.AddProperties(&module.properties)

	return module
}

// Dependency tag for values property
type valueSetType struct {
	blueprint.BaseDependencyTag
}

var valueSetTag = valueSetType{}

// Provider published by aconfig_value_set
type valueSetProviderData struct {
	// The package of each of the
	// (map of package --> aconfig_module)
	AvailablePackages map[string]android.Paths
}

var valueSetProviderKey = blueprint.NewProvider[valueSetProviderData]()

func (module *ValueSetModule) FindAconfigValuesFromSrc(ctx android.BottomUpMutatorContext) map[string]android.Path {
	moduleDir := ctx.ModuleDir()
	srcs := android.PathsForModuleSrcExcludes(ctx, module.properties.Srcs, []string{ctx.BlueprintsFile()})

	aconfigValuesPrefix := strings.Replace(module.Name(), "aconfig_value_set", "aconfig-values", 1)
	moduleNamesSrcMap := make(map[string]android.Path)
	for _, src := range srcs {
		subDir := strings.TrimPrefix(src.String(), moduleDir+"/")
		packageName, _, found := strings.Cut(subDir, "/")
		if found {
			moduleName := fmt.Sprintf("%s-%s-all", aconfigValuesPrefix, packageName)
			moduleNamesSrcMap[moduleName] = src
		}
	}
	return moduleNamesSrcMap
}

func (module *ValueSetModule) DepsMutator(ctx android.BottomUpMutatorContext) {

	// TODO: b/366285733 - Replace the file path based solution with more robust solution.
	aconfigValuesMap := module.FindAconfigValuesFromSrc(ctx)
	for _, moduleName := range android.SortedKeys(aconfigValuesMap) {
		if ctx.OtherModuleExists(moduleName) {
			ctx.AddDependency(ctx.Module(), valueSetTag, moduleName)
		} else {
			ctx.ModuleErrorf("module %q not found. Rename the aconfig_values module defined in %q to %q", moduleName, aconfigValuesMap[moduleName], moduleName)
		}
	}

	deps := ctx.AddDependency(ctx.Module(), valueSetTag, module.properties.Values...)
	for _, dep := range deps {
		_, ok := dep.(*ValuesModule)
		if !ok {
			ctx.PropertyErrorf("values", "values must be a aconfig_values module")
			return
		}
	}
}

func (module *ValueSetModule) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	// Accumulate the packages of the values modules listed, and set that as an
	// valueSetProviderKey provider that aconfig modules can read and use
	// to append values to their aconfig actions.
	packages := make(map[string]android.Paths)
	ctx.VisitDirectDeps(func(dep android.Module) {
		if depData, ok := android.OtherModuleProvider(ctx, dep, valuesProviderKey); ok {
			srcs := make([]android.Path, len(depData.Values))
			copy(srcs, depData.Values)
			packages[depData.Package] = srcs
		}

	})
	android.SetProvider(ctx, valueSetProviderKey, valueSetProviderData{
		AvailablePackages: packages,
	})
}
