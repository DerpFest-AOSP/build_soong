// Copyright (C) 2021 The Android Open Source Project
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

package apex

import (
	"android/soong/android"
)

type DeapexerProperties struct {
	// List of common modules that may need access to files exported by this module.
	//
	// A common module in this sense is one that is not arch specific but uses a common variant for
	// all architectures, e.g. java.
	CommonModules []string

	// List of modules that use an embedded .prof to guide optimization of the equivalent dexpreopt artifact
	// This is a subset of CommonModules
	DexpreoptProfileGuidedModules []string

	// List of files exported from the .apex file by this module
	//
	// Each entry is a path from the apex root, e.g. javalib/core-libart.jar.
	ExportedFiles []string
}

type SelectedApexProperties struct {
	// The path to the apex selected for use by this module.
	//
	// Is tagged as `android:"path"` because it will usually contain a string of the form ":<module>"
	// and is tagged as "`blueprint:"mutate"` because it is only initialized in a LoadHook not an
	// Android.bp file.
	Selected_apex *string `android:"path" blueprint:"mutated"`
}

// deapex creates the build rules to deapex a prebuilt .apex file
// it returns a pointer to a DeapexerInfo object
func deapex(ctx android.ModuleContext, apexFile android.Path, deapexerProps DeapexerProperties) *android.DeapexerInfo {
	// Create and remember the directory into which the .apex file's contents will be unpacked.
	deapexerOutput := android.PathForModuleOut(ctx, "deapexer")

	exports := make(map[string]android.WritablePath)

	// Create mappings from apex relative path to the extracted file's path.
	exportedPaths := make(android.Paths, 0, len(exports))
	for _, path := range deapexerProps.ExportedFiles {
		// Populate the exports that this makes available.
		extractedPath := deapexerOutput.Join(ctx, path)
		exports[path] = extractedPath
		exportedPaths = append(exportedPaths, extractedPath)
	}

	// If the prebuilt_apex exports any files then create a build rule that unpacks the apex using
	// deapexer and verifies that all the required files were created. Also, make the mapping from
	// apex relative path to extracted file path available for other modules.
	if len(exports) > 0 {
		// Make the information available for other modules.
		di := android.NewDeapexerInfo(ctx.ModuleName(), exports, deapexerProps.CommonModules)
		di.AddDexpreoptProfileGuidedExportedModuleNames(deapexerProps.DexpreoptProfileGuidedModules...)

		// Create a sorted list of the files that this exports.
		exportedPaths = android.SortedUniquePaths(exportedPaths)

		// The apex needs to export some files so create a ninja rule to unpack the apex and check that
		// the required files are present.
		builder := android.NewRuleBuilder(pctx, ctx)
		command := builder.Command()
		command.
			Tool(android.PathForSource(ctx, "build/soong/scripts/unpack-prebuilt-apex.sh")).
			BuiltTool("deapexer").
			BuiltTool("debugfs").
			BuiltTool("fsck.erofs").
			Input(apexFile).
			Text(deapexerOutput.String())
		for _, p := range exportedPaths {
			command.Output(p.(android.WritablePath))
		}
		builder.Build("deapexer", "deapex "+ctx.ModuleName())
		return &di
	}
	return nil
}
