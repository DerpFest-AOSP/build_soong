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

type vintfFragmentProperties struct {
	// Vintf fragment XML file.
	Src string `android:"path"`
}

type vintfFragmentModule struct {
	ModuleBase

	properties vintfFragmentProperties

	installDirPath InstallPath
	outputFilePath OutputPath
}

func init() {
	registerVintfFragmentComponents(InitRegistrationContext)
}

func registerVintfFragmentComponents(ctx RegistrationContext) {
	ctx.RegisterModuleType("vintf_fragment", vintfLibraryFactory)
}

// vintf_fragment module processes vintf fragment file and installs under etc/vintf/manifest.
// Vintf fragment files formerly listed in vintf_fragment property would be transformed into
// this module type.
func vintfLibraryFactory() Module {
	m := &vintfFragmentModule{}
	m.AddProperties(
		&m.properties,
	)
	InitAndroidArchModule(m, DeviceSupported, MultilibFirst)

	return m
}

func (m *vintfFragmentModule) GenerateAndroidBuildActions(ctx ModuleContext) {
	builder := NewRuleBuilder(pctx, ctx)
	srcVintfFragment := PathForModuleSrc(ctx, m.properties.Src)
	processedVintfFragment := PathForModuleOut(ctx, srcVintfFragment.Base())

	// Process vintf fragment source file with assemble_vintf tool
	builder.Command().
		Flag("VINTF_IGNORE_TARGET_FCM_VERSION=true").
		BuiltTool("assemble_vintf").
		FlagWithInput("-i ", srcVintfFragment).
		FlagWithOutput("-o ", processedVintfFragment)

	builder.Build("assemble_vintf", "Process vintf fragment "+processedVintfFragment.String())

	m.installDirPath = PathForModuleInstall(ctx, "etc", "vintf", "manifest")
	m.outputFilePath = processedVintfFragment.OutputPath

	ctx.InstallFile(m.installDirPath, processedVintfFragment.Base(), processedVintfFragment)
}

// Make this module visible to AndroidMK so it can be referenced from modules defined from Android.mk files
func (m *vintfFragmentModule) AndroidMkEntries() []AndroidMkEntries {
	return []AndroidMkEntries{{
		Class:      "ETC",
		OutputFile: OptionalPathForPath(m.outputFilePath),
		ExtraEntries: []AndroidMkExtraEntriesFunc{
			func(ctx AndroidMkExtraEntriesContext, entries *AndroidMkEntries) {
				entries.SetString("LOCAL_MODULE_PATH", m.installDirPath.String())
				entries.SetString("LOCAL_INSTALLED_MODULE_STEM", m.outputFilePath.Base())
			},
		},
	}}
}
