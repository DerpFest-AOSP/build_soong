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
	"path/filepath"
	"slices"
	"strings"

	"android/soong/android"

	"github.com/google/blueprint"
)

type AconfigReleaseConfigValue struct {
	ReleaseConfig string
	Values        []string `blueprint:"mutated"`
}

type DeclarationsModule struct {
	android.ModuleBase
	android.DefaultableModuleBase
	blueprint.IncrementalModule

	// Properties for "aconfig_declarations"
	properties struct {
		// aconfig files, relative to this Android.bp file
		Srcs []string `android:"path"`

		// Release config flag package
		Package string

		// Values for release configs / RELEASE_ACONFIG_VALUE_SETS
		// The current release config is `ReleaseConfig: ""`, others
		// are from RELEASE_ACONFIG_EXTRA_RELEASE_CONFIGS.
		ReleaseConfigValues []AconfigReleaseConfigValue

		// Container(system/vendor/apex) that this module belongs to
		Container string

		// The flags will only be repackaged if this prop is true.
		Exportable bool
	}
}

func DeclarationsFactory() android.Module {
	module := &DeclarationsModule{}

	android.InitAndroidModule(module)
	android.InitDefaultableModule(module)
	module.AddProperties(&module.properties)

	return module
}

type implicitValuesTagType struct {
	blueprint.BaseDependencyTag

	// The release config name for these values.
	// Empty string for the actual current release config.
	ReleaseConfig string
}

var implicitValuesTag = implicitValuesTagType{}

func (module *DeclarationsModule) DepsMutator(ctx android.BottomUpMutatorContext) {
	// Validate Properties
	if len(module.properties.Srcs) == 0 {
		ctx.PropertyErrorf("srcs", "missing source files")
		return
	}
	if len(module.properties.Package) == 0 {
		ctx.PropertyErrorf("package", "missing package property")
	}
	if len(module.properties.Container) == 0 {
		ctx.PropertyErrorf("container", "missing container property")
	}

	// treating system_ext as system partition as we are combining them as one container
	// TODO remove this logic once we start enforcing that system_ext cannot be specified as
	// container in the container field.
	if module.properties.Container == "system_ext" {
		module.properties.Container = "system"
	}

	// Add a dependency on the aconfig_value_sets defined in
	// RELEASE_ACONFIG_VALUE_SETS, and add any aconfig_values that
	// match our package.
	valuesFromConfig := ctx.Config().ReleaseAconfigValueSets()
	if len(valuesFromConfig) > 0 {
		ctx.AddDependency(ctx.Module(), implicitValuesTag, valuesFromConfig...)
	}
	for rcName, valueSets := range ctx.Config().ReleaseAconfigExtraReleaseConfigsValueSets() {
		if len(valueSets) > 0 {
			ctx.AddDependency(ctx.Module(), implicitValuesTagType{ReleaseConfig: rcName}, valueSets...)
		}
	}
}

func joinAndPrefix(prefix string, values []string) string {
	var sb strings.Builder
	for _, v := range values {
		sb.WriteString(prefix)
		sb.WriteString(v)
	}
	return sb.String()
}

func optionalVariable(prefix string, value string) string {
	var sb strings.Builder
	if value != "" {
		sb.WriteString(prefix)
		sb.WriteString(value)
	}
	return sb.String()
}

// Assemble the actual filename.
// If `rcName` is not empty, then insert "-{rcName}" into the path before the
// file extension.
func assembleFileName(rcName, path string) string {
	if rcName == "" {
		return path
	}
	dir, file := filepath.Split(path)
	rcName = "-" + rcName
	ext := filepath.Ext(file)
	base := file[:len(file)-len(ext)]
	return dir + base + rcName + ext
}

func (module *DeclarationsModule) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	// Determine which release configs we are processing.
	//
	// We always process the current release config (empty string).
	// We may have been told to also create artifacts for some others.
	configs := append([]string{""}, ctx.Config().ReleaseAconfigExtraReleaseConfigs()...)
	slices.Sort(configs)

	values := make(map[string][]string)
	valuesFiles := make(map[string][]android.Path, 0)
	providerData := android.AconfigReleaseDeclarationsProviderData{}
	ctx.VisitDirectDeps(func(dep android.Module) {
		if depData, ok := android.OtherModuleProvider(ctx, dep, valueSetProviderKey); ok {
			depTag := ctx.OtherModuleDependencyTag(dep)
			for _, config := range configs {
				tag := implicitValuesTagType{ReleaseConfig: config}
				if depTag == tag {
					paths, ok := depData.AvailablePackages[module.properties.Package]
					if ok {
						valuesFiles[config] = append(valuesFiles[config], paths...)
						for _, path := range paths {
							values[config] = append(values[config], path.String())
						}
					}
				}
			}
		}
	})
	for _, config := range configs {
		module.properties.ReleaseConfigValues = append(module.properties.ReleaseConfigValues, AconfigReleaseConfigValue{
			ReleaseConfig: config,
			Values:        values[config],
		})

		// Intermediate format
		declarationFiles := android.PathsForModuleSrc(ctx, module.properties.Srcs)
		intermediateCacheFilePath := android.PathForModuleOut(ctx, assembleFileName(config, "intermediate.pb"))
		var defaultPermission string
		defaultPermission = ctx.Config().ReleaseAconfigFlagDefaultPermission()
		if config != "" {
			if confPerm, ok := ctx.Config().GetBuildFlag("RELEASE_ACONFIG_FLAG_DEFAULT_PERMISSION_" + config); ok {
				defaultPermission = confPerm
			}
		}
		inputFiles := make([]android.Path, len(declarationFiles))
		copy(inputFiles, declarationFiles)
		inputFiles = append(inputFiles, valuesFiles[config]...)
		args := map[string]string{
			"release_version":    ctx.Config().ReleaseVersion(),
			"package":            module.properties.Package,
			"declarations":       android.JoinPathsWithPrefix(declarationFiles, "--declarations "),
			"values":             joinAndPrefix(" --values ", values[config]),
			"default-permission": optionalVariable(" --default-permission ", defaultPermission),
		}
		if len(module.properties.Container) > 0 {
			args["container"] = "--container " + module.properties.Container
		}
		ctx.Build(pctx, android.BuildParams{
			Rule:        aconfigRule,
			Output:      intermediateCacheFilePath,
			Inputs:      inputFiles,
			Description: "aconfig_declarations",
			Args:        args,
		})

		intermediateDumpFilePath := android.PathForModuleOut(ctx, assembleFileName(config, "intermediate.txt"))
		ctx.Build(pctx, android.BuildParams{
			Rule:        aconfigTextRule,
			Output:      intermediateDumpFilePath,
			Inputs:      android.Paths{intermediateCacheFilePath},
			Description: "aconfig_text",
		})

		providerData[config] = android.AconfigDeclarationsProviderData{
			Package:                     module.properties.Package,
			Container:                   module.properties.Container,
			Exportable:                  module.properties.Exportable,
			IntermediateCacheOutputPath: intermediateCacheFilePath,
			IntermediateDumpOutputPath:  intermediateDumpFilePath,
		}
	}
	android.SetProvider(ctx, android.AconfigDeclarationsProviderKey, providerData[""])
	android.SetProvider(ctx, android.AconfigReleaseDeclarationsProviderKey, providerData)
}

var _ blueprint.Incremental = &DeclarationsModule{}
