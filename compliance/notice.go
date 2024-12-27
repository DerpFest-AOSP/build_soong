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

package compliance

import (
	"path/filepath"

	"android/soong/android"
	"github.com/google/blueprint"
)

func init() {
	RegisterNoticeXmlBuildComponents(android.InitRegistrationContext)
}

var PrepareForTestWithNoticeXmlBuildComponents = android.GroupFixturePreparers(
	android.FixtureRegisterWithContext(RegisterNoticeXmlBuildComponents),
)

var PrepareForTestWithNoticeXml = android.GroupFixturePreparers(
	PrepareForTestWithNoticeXmlBuildComponents,
)

func RegisterNoticeXmlBuildComponents(ctx android.RegistrationContext) {
	ctx.RegisterModuleType("notice_xml", NoticeXmlFactory)
}

var (
	pctx = android.NewPackageContext("android/soong/compliance")

	genNoticeXml = pctx.HostBinToolVariable("genNoticeXml", "gen_notice_xml")

	// Command to generate NOTICE.xml.gz for a partition
	genNoticeXmlRule = pctx.AndroidStaticRule("genNoticeXmlRule", blueprint.RuleParams{
		Command: "rm -rf $out && " +
			"${genNoticeXml} --output_file ${out} --metadata ${in} --partition ${partition} --product_out ${productOut} --soong_out ${soongOut}",
		CommandDeps: []string{"${genNoticeXml}"},
	}, "partition", "productOut", "soongOut")
)

func NoticeXmlFactory() android.Module {
	m := &NoticeXmlModule{}
	m.AddProperties(&m.props)
	android.InitAndroidArchModule(m, android.DeviceSupported, android.MultilibFirst)
	return m
}

type NoticeXmlModule struct {
	android.ModuleBase

	props noticeXmlProperties

	outputFile  android.OutputPath
	installPath android.InstallPath
}

type noticeXmlProperties struct {
	Partition_name string
}

func (nx *NoticeXmlModule) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	output := android.PathForModuleOut(ctx, "NOTICE.xml.gz")
	metadataDb := android.PathForOutput(ctx, "compliance-metadata", ctx.Config().DeviceProduct(), "compliance-metadata.db")
	ctx.Build(pctx, android.BuildParams{
		Rule:   genNoticeXmlRule,
		Input:  metadataDb,
		Output: output,
		Args: map[string]string{
			"productOut": filepath.Join(ctx.Config().OutDir(), "target", "product", ctx.Config().DeviceName()),
			"soongOut":   ctx.Config().SoongOutDir(),
			"partition":  nx.props.Partition_name,
		},
	})

	nx.outputFile = output.OutputPath

	if android.Bool(ctx.Config().ProductVariables().UseSoongSystemImage) {
		nx.installPath = android.PathForModuleInPartitionInstall(ctx, nx.props.Partition_name, "etc")
		ctx.InstallFile(nx.installPath, "NOTICE.xml.gz", nx.outputFile)
	}
}

func (nx *NoticeXmlModule) AndroidMkEntries() []android.AndroidMkEntries {
	return []android.AndroidMkEntries{{
		Class:      "ETC",
		OutputFile: android.OptionalPathForPath(nx.outputFile),
	}}
}
