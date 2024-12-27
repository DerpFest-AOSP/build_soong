// Copyright 2016 Google Inc. All rights reserved.
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

package cc

import (
	"path/filepath"

	"android/soong/android"

	"github.com/google/blueprint"
)

var (
	preprocessNdkHeader = pctx.AndroidStaticRule("preprocessNdkHeader",
		blueprint.RuleParams{
			Command:     "$preprocessor -o $out $in",
			CommandDeps: []string{"$preprocessor"},
		},
		"preprocessor")
)

// Returns the NDK base include path for use with sdk_version current. Usable with -I.
func getCurrentIncludePath(ctx android.PathContext) android.OutputPath {
	return getNdkSysrootBase(ctx).Join(ctx, "usr/include")
}

type headerProperties struct {
	// Base directory of the headers being installed. As an example:
	//
	// ndk_headers {
	//     name: "foo",
	//     from: "include",
	//     to: "",
	//     srcs: ["include/foo/bar/baz.h"],
	// }
	//
	// Will install $SYSROOT/usr/include/foo/bar/baz.h. If `from` were instead
	// "include/foo", it would have installed $SYSROOT/usr/include/bar/baz.h.
	From *string

	// Install path within the sysroot. This is relative to usr/include.
	To *string

	// List of headers to install. Glob compatible. Common case is "include/**/*.h".
	Srcs []string `android:"path"`

	// Source paths that should be excluded from the srcs glob.
	Exclude_srcs []string `android:"path"`

	// Path to the NOTICE file associated with the headers.
	License *string `android:"path"`

	// Set to true if the headers installed by this module should skip
	// verification. This step ensures that each header is self-contained (can
	// be #included alone) and is valid C. This should not be disabled except in
	// rare cases. Outside bionic and external, if you're using this option
	// you've probably made a mistake.
	Skip_verification *bool
}

type headerModule struct {
	android.ModuleBase

	properties headerProperties

	srcPaths     android.Paths
	installPaths android.Paths
	licensePath  android.Path
}

func getHeaderInstallDir(ctx android.ModuleContext, header android.Path, from string,
	to string) android.OutputPath {
	// Output path is the sysroot base + "usr/include" + to directory + directory component
	// of the file without the leading from directory stripped.
	//
	// Given:
	// sysroot base = "ndk/sysroot"
	// from = "include/foo"
	// to = "bar"
	// header = "include/foo/woodly/doodly.h"
	// output path = "ndk/sysroot/usr/include/bar/woodly/doodly.h"

	// full/platform/path/to/include/foo
	fullFromPath := android.PathForModuleSrc(ctx, from)

	// full/platform/path/to/include/foo/woodly
	headerDir := filepath.Dir(header.String())

	// woodly
	strippedHeaderDir, err := filepath.Rel(fullFromPath.String(), headerDir)
	if err != nil {
		ctx.ModuleErrorf("filepath.Rel(%q, %q) failed: %s", headerDir,
			fullFromPath.String(), err)
	}

	// full/platform/path/to/sysroot/usr/include/bar/woodly
	installDir := getCurrentIncludePath(ctx).Join(ctx, to, strippedHeaderDir)

	// full/platform/path/to/sysroot/usr/include/bar/woodly/doodly.h
	return installDir
}

func (m *headerModule) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	if String(m.properties.License) == "" {
		ctx.PropertyErrorf("license", "field is required")
	}

	m.licensePath = android.PathForModuleSrc(ctx, String(m.properties.License))

	m.srcPaths = android.PathsForModuleSrcExcludes(ctx, m.properties.Srcs, m.properties.Exclude_srcs)
	for _, header := range m.srcPaths {
		installDir := getHeaderInstallDir(ctx, header, String(m.properties.From),
			String(m.properties.To))
		installPath := installDir.Join(ctx, header.Base())
		ctx.Build(pctx, android.BuildParams{
			Rule:   android.Cp,
			Input:  header,
			Output: installPath,
		})
		m.installPaths = append(m.installPaths, installPath)
	}

	if len(m.installPaths) == 0 {
		ctx.ModuleErrorf("srcs %q matched zero files", m.properties.Srcs)
	}
}

// ndk_headers installs the sets of ndk headers defined in the srcs property
// to the sysroot base + "usr/include" + to directory + directory component.
// ndk_headers requires the license file to be specified. Example:
//
//	Given:
//	sysroot base = "ndk/sysroot"
//	from = "include/foo"
//	to = "bar"
//	header = "include/foo/woodly/doodly.h"
//	output path = "ndk/sysroot/usr/include/bar/woodly/doodly.h"
func NdkHeadersFactory() android.Module {
	module := &headerModule{}
	module.AddProperties(&module.properties)
	android.InitAndroidModule(module)
	return module
}

// preprocessed_ndk_header {
//
//	name: "foo",
//	preprocessor: "foo.sh",
//	srcs: [...],
//	to: "android",
//
// }
//
// Will invoke the preprocessor as:
//
//	$preprocessor -o $SYSROOT/usr/include/android/needs_preproc.h $src
//
// For each src in srcs.
type preprocessedHeadersProperties struct {
	// The preprocessor to run. Must be a program inside the source directory
	// with no dependencies.
	Preprocessor *string

	// Source path to the files to be preprocessed.
	Srcs []string

	// Source paths that should be excluded from the srcs glob.
	Exclude_srcs []string

	// Install path within the sysroot. This is relative to usr/include.
	To *string

	// Path to the NOTICE file associated with the headers.
	License *string

	// Set to true if the headers installed by this module should skip
	// verification. This step ensures that each header is self-contained (can
	// be #included alone) and is valid C. This should not be disabled except in
	// rare cases. Outside bionic and external, if you're using this option
	// you've probably made a mistake.
	Skip_verification *bool
}

type preprocessedHeadersModule struct {
	android.ModuleBase

	properties preprocessedHeadersProperties

	srcPaths     android.Paths
	installPaths android.Paths
	licensePath  android.Path
}

func (m *preprocessedHeadersModule) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	if String(m.properties.License) == "" {
		ctx.PropertyErrorf("license", "field is required")
	}

	preprocessor := android.PathForModuleSrc(ctx, String(m.properties.Preprocessor))
	m.licensePath = android.PathForModuleSrc(ctx, String(m.properties.License))

	m.srcPaths = android.PathsForModuleSrcExcludes(ctx, m.properties.Srcs, m.properties.Exclude_srcs)
	installDir := getCurrentIncludePath(ctx).Join(ctx, String(m.properties.To))
	for _, src := range m.srcPaths {
		installPath := installDir.Join(ctx, src.Base())
		m.installPaths = append(m.installPaths, installPath)

		ctx.Build(pctx, android.BuildParams{
			Rule:        preprocessNdkHeader,
			Description: "preprocess " + src.Rel(),
			Input:       src,
			Output:      installPath,
			Args: map[string]string{
				"preprocessor": preprocessor.String(),
			},
		})
	}

	if len(m.installPaths) == 0 {
		ctx.ModuleErrorf("srcs %q matched zero files", m.properties.Srcs)
	}
}

// preprocessed_ndk_headers preprocesses all the ndk headers listed in the srcs
// property by executing the command defined in the preprocessor property.
func preprocessedNdkHeadersFactory() android.Module {
	module := &preprocessedHeadersModule{}

	module.AddProperties(&module.properties)

	android.InitAndroidModule(module)

	return module
}
