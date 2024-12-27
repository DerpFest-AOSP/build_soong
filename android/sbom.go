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

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/google/blueprint"
)

var (
	// Command line tool to generate SBOM in Soong
	genSbom = pctx.HostBinToolVariable("genSbom", "gen_sbom")

	// Command to generate SBOM in Soong.
	genSbomRule = pctx.AndroidStaticRule("genSbomRule", blueprint.RuleParams{
		Command:     "rm -rf $out && ${genSbom} --output_file ${out} --metadata ${in} --product_out ${productOut} --soong_out ${soongOut} --build_version \"$$(cat ${buildFingerprintFile})\" --product_mfr \"${productManufacturer}\" --json",
		CommandDeps: []string{"${genSbom}"},
	}, "productOut", "soongOut", "buildFingerprintFile", "productManufacturer")
)

func init() {
	RegisterSbomSingleton(InitRegistrationContext)
}

func RegisterSbomSingleton(ctx RegistrationContext) {
	ctx.RegisterParallelSingletonType("sbom_singleton", sbomSingletonFactory)
}

// sbomSingleton is used to generate build actions of generating SBOM of products.
type sbomSingleton struct {
	sbomFile OutputPath
}

func sbomSingletonFactory() Singleton {
	return &sbomSingleton{}
}

// Generates SBOM of products
func (this *sbomSingleton) GenerateBuildActions(ctx SingletonContext) {
	if !ctx.Config().HasDeviceProduct() {
		return
	}
	// Get all METADATA files and add them as implicit input
	metadataFileListFile := PathForArbitraryOutput(ctx, ".module_paths", "METADATA.list")
	f, err := ctx.Config().fs.Open(metadataFileListFile.String())
	if err != nil {
		panic(err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	allMetadataFiles := strings.Split(string(b), "\n")
	implicits := []Path{metadataFileListFile}
	for _, path := range allMetadataFiles {
		implicits = append(implicits, PathForSource(ctx, path))
	}
	prodVars := ctx.Config().productVariables
	buildFingerprintFile := PathForArbitraryOutput(ctx, "target", "product", String(prodVars.DeviceName), "build_fingerprint.txt")
	implicits = append(implicits, buildFingerprintFile)

	// Add installed_files.stamp as implicit input, which depends on all installed files of the product.
	installedFilesStamp := PathForOutput(ctx, "compliance-metadata", ctx.Config().DeviceProduct(), "installed_files.stamp")
	implicits = append(implicits, installedFilesStamp)

	metadataDb := PathForOutput(ctx, "compliance-metadata", ctx.Config().DeviceProduct(), "compliance-metadata.db")
	this.sbomFile = PathForOutput(ctx, "sbom", ctx.Config().DeviceProduct(), "sbom.spdx.json")
	ctx.Build(pctx, BuildParams{
		Rule:      genSbomRule,
		Input:     metadataDb,
		Implicits: implicits,
		Output:    this.sbomFile,
		Args: map[string]string{
			"productOut":           filepath.Join(ctx.Config().OutDir(), "target", "product", String(prodVars.DeviceName)),
			"soongOut":             ctx.Config().soongOutDir,
			"buildFingerprintFile": buildFingerprintFile.String(),
			"productManufacturer":  ctx.Config().ProductVariables().ProductManufacturer,
		},
	})

	if !ctx.Config().UnbundledBuildApps() {
		// When building SBOM of products, phony rule "sbom" is for generating product SBOM in Soong.
		ctx.Build(pctx, BuildParams{
			Rule:   blueprint.Phony,
			Inputs: []Path{this.sbomFile},
			Output: PathForPhony(ctx, "sbom"),
		})
	}
}

func (this *sbomSingleton) MakeVars(ctx MakeVarsContext) {
	// When building SBOM of products
	if !ctx.Config().UnbundledBuildApps() {
		ctx.DistForGoalWithFilename("droid", this.sbomFile, "sbom/sbom.spdx.json")
	}
}
