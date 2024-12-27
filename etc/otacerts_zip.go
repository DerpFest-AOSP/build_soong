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

package etc

import (
	"android/soong/android"

	"github.com/google/blueprint/proptools"
)

func init() {
	RegisterOtacertsZipBuildComponents(android.InitRegistrationContext)
}

func RegisterOtacertsZipBuildComponents(ctx android.RegistrationContext) {
	ctx.RegisterModuleType("otacerts_zip", otacertsZipFactory)
}

type otacertsZipProperties struct {
	// Make this module available when building for recovery.
	// Only the recovery partition is available.
	Recovery_available *bool

	// Optional subdirectory under which the zip file is installed into.
	Relative_install_path *string

	// Optional name for the installed file. If unspecified, otacerts.zip is used.
	Filename *string
}

type otacertsZipModule struct {
	android.ModuleBase

	properties otacertsZipProperties
	outputPath android.OutputPath
}

// otacerts_zip collects key files defined in PRODUCT_DEFAULT_DEV_CERTIFICATE
// and PRODUCT_EXTRA_OTA_KEYS for system or PRODUCT_EXTRA_RECOVERY_KEYS for
// recovery image. The output file (otacerts.zip by default) is installed into
// the relative_install_path directory under the etc directory of the target
// partition.
func otacertsZipFactory() android.Module {
	module := &otacertsZipModule{}
	module.AddProperties(&module.properties)
	android.InitAndroidArchModule(module, android.DeviceSupported, android.MultilibCommon)
	return module
}

var _ android.ImageInterface = (*otacertsZipModule)(nil)

func (m *otacertsZipModule) ImageMutatorBegin(ctx android.BaseModuleContext) {}

func (m *otacertsZipModule) VendorVariantNeeded(ctx android.BaseModuleContext) bool {
	return false
}

func (m *otacertsZipModule) ProductVariantNeeded(ctx android.BaseModuleContext) bool {
	return false
}

func (m *otacertsZipModule) CoreVariantNeeded(ctx android.BaseModuleContext) bool {
	return !m.ModuleBase.InstallInRecovery()
}

func (m *otacertsZipModule) RamdiskVariantNeeded(ctx android.BaseModuleContext) bool {
	return false
}

func (m *otacertsZipModule) VendorRamdiskVariantNeeded(ctx android.BaseModuleContext) bool {
	return false
}

func (m *otacertsZipModule) DebugRamdiskVariantNeeded(ctx android.BaseModuleContext) bool {
	return false
}

func (m *otacertsZipModule) RecoveryVariantNeeded(ctx android.BaseModuleContext) bool {
	return proptools.Bool(m.properties.Recovery_available) || m.ModuleBase.InstallInRecovery()
}

func (m *otacertsZipModule) ExtraImageVariations(ctx android.BaseModuleContext) []string {
	return nil
}

func (m *otacertsZipModule) SetImageVariation(ctx android.BaseModuleContext, variation string) {
}

func (m *otacertsZipModule) InRecovery() bool {
	return m.ModuleBase.InRecovery() || m.ModuleBase.InstallInRecovery()
}

func (m *otacertsZipModule) InstallInRecovery() bool {
	return m.InRecovery()
}

func (m *otacertsZipModule) outputFileName() string {
	// Use otacerts.zip if not specified.
	return proptools.StringDefault(m.properties.Filename, "otacerts.zip")
}

func (m *otacertsZipModule) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	// Read .x509.pem file defined in PRODUCT_DEFAULT_DEV_CERTIFICATE or the default test key.
	pem, _ := ctx.Config().DefaultAppCertificate(ctx)
	// Read .x509.pem files listed  in PRODUCT_EXTRA_OTA_KEYS or PRODUCT_EXTRA_RECOVERY_KEYS.
	extras := ctx.Config().ExtraOtaKeys(ctx, m.InRecovery())
	srcPaths := append([]android.SourcePath{pem}, extras...)
	m.outputPath = android.PathForModuleOut(ctx, m.outputFileName()).OutputPath

	rule := android.NewRuleBuilder(pctx, ctx)
	cmd := rule.Command().BuiltTool("soong_zip").
		FlagWithOutput("-o ", m.outputPath).
		Flag("-j ").
		Flag("-symlinks=false ")
	for _, src := range srcPaths {
		cmd.FlagWithInput("-f ", src)
	}
	rule.Build(ctx.ModuleName(), "Generating the otacerts zip file")

	installPath := android.PathForModuleInstall(ctx, "etc", proptools.String(m.properties.Relative_install_path))
	ctx.InstallFile(installPath, m.outputFileName(), m.outputPath)
}

func (m *otacertsZipModule) AndroidMkEntries() []android.AndroidMkEntries {
	nameSuffix := ""
	if m.InRecovery() {
		nameSuffix = ".recovery"
	}
	return []android.AndroidMkEntries{android.AndroidMkEntries{
		Class:      "ETC",
		SubName:    nameSuffix,
		OutputFile: android.OptionalPathForPath(m.outputPath),
	}}
}
