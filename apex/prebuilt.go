// Copyright (C) 2019 The Android Open Source Project
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
	"strconv"
	"strings"

	"android/soong/android"
	"android/soong/dexpreopt"
	"android/soong/java"
	"android/soong/provenance"

	"github.com/google/blueprint"
	"github.com/google/blueprint/proptools"
)

var (
	extractMatchingApex = pctx.StaticRule(
		"extractMatchingApex",
		blueprint.RuleParams{
			Command: `rm -rf "$out" && ` +
				`${extract_apks} -o "${out}" -allow-prereleased=${allow-prereleased} ` +
				`-sdk-version=${sdk-version} -skip-sdk-check=${skip-sdk-check} -abis=${abis} ` +
				`-screen-densities=all -extract-single ` +
				`${in}`,
			CommandDeps: []string{"${extract_apks}"},
		},
		"abis", "allow-prereleased", "sdk-version", "skip-sdk-check")
	decompressApex = pctx.StaticRule("decompressApex", blueprint.RuleParams{
		Command:     `rm -rf $out && ${deapexer} decompress --copy-if-uncompressed --input ${in} --output ${out}`,
		CommandDeps: []string{"${deapexer}"},
		Description: "decompress $out",
	})
)

type prebuilt interface {
	isForceDisabled() bool
	InstallFilename() string
}

type prebuiltCommon struct {
	android.ModuleBase
	java.Dexpreopter
	prebuilt android.Prebuilt

	// Properties common to both prebuilt_apex and apex_set.
	prebuiltCommonProperties *PrebuiltCommonProperties

	installDir      android.InstallPath
	installFilename string
	installedFile   android.InstallPath
	outputApex      android.WritablePath

	// fragment for this apex for apexkeys.txt
	apexKeysPath android.WritablePath

	// Installed locations of symlinks for backward compatibility.
	compatSymlinks android.InstallPaths

	hostRequired        []string
	requiredModuleNames []string
}

type sanitizedPrebuilt interface {
	hasSanitizedSource(sanitizer string) bool
}

type PrebuiltCommonProperties struct {
	SelectedApexProperties

	// Canonical name of this APEX. Used to determine the path to the activated APEX on
	// device (/apex/<apex_name>). If unspecified, follows the name property.
	Apex_name *string

	// Name of the source APEX that gets shadowed by this prebuilt
	// e.g. com.mycompany.android.myapex
	// If unspecified, follows the naming convention that the source apex of
	// the prebuilt is Name() without "prebuilt_" prefix
	Source_apex_name *string

	ForceDisable bool `blueprint:"mutated"`

	// whether the extracted apex file is installable.
	Installable *bool

	// optional name for the installed apex. If unspecified, name of the
	// module is used as the file name
	Filename *string

	// names of modules to be overridden. Listed modules can only be other binaries
	// (in Make or Soong).
	// This does not completely prevent installation of the overridden binaries, but if both
	// binaries would be installed by default (in PRODUCT_PACKAGES) the other binary will be removed
	// from PRODUCT_PACKAGES.
	Overrides []string

	// List of bootclasspath fragments inside this prebuilt APEX bundle and for which this APEX
	// bundle will create an APEX variant.
	Exported_bootclasspath_fragments []string

	// List of systemserverclasspath fragments inside this prebuilt APEX bundle and for which this
	// APEX bundle will create an APEX variant.
	Exported_systemserverclasspath_fragments []string

	// Path to the .prebuilt_info file of the prebuilt apex.
	// In case of mainline modules, the .prebuilt_info file contains the build_id that was used to
	// generate the prebuilt.
	Prebuilt_info *string `android:"path"`
}

// initPrebuiltCommon initializes the prebuiltCommon structure and performs initialization of the
// module that is common to Prebuilt and ApexSet.
func (p *prebuiltCommon) initPrebuiltCommon(module android.Module, properties *PrebuiltCommonProperties) {
	p.prebuiltCommonProperties = properties
	android.InitSingleSourcePrebuiltModule(module.(android.PrebuiltInterface), properties, "Selected_apex")
	android.InitAndroidMultiTargetsArchModule(module, android.DeviceSupported, android.MultilibCommon)
}

func (p *prebuiltCommon) ApexVariationName() string {
	return proptools.StringDefault(p.prebuiltCommonProperties.Apex_name, p.BaseModuleName())
}

func (p *prebuiltCommon) BaseModuleName() string {
	return proptools.StringDefault(p.prebuiltCommonProperties.Source_apex_name, p.ModuleBase.BaseModuleName())
}

func (p *prebuiltCommon) Prebuilt() *android.Prebuilt {
	return &p.prebuilt
}

func (p *prebuiltCommon) isForceDisabled() bool {
	return p.prebuiltCommonProperties.ForceDisable
}

func (p *prebuiltCommon) checkForceDisable(ctx android.ModuleContext) bool {
	forceDisable := false

	// Force disable the prebuilts when we are doing unbundled build. We do unbundled build
	// to build the prebuilts themselves.
	forceDisable = forceDisable || ctx.Config().UnbundledBuild()

	// b/137216042 don't use prebuilts when address sanitizer is on, unless the prebuilt has a sanitized source
	sanitized := ctx.Module().(sanitizedPrebuilt)
	forceDisable = forceDisable || (android.InList("address", ctx.Config().SanitizeDevice()) && !sanitized.hasSanitizedSource("address"))
	forceDisable = forceDisable || (android.InList("hwaddress", ctx.Config().SanitizeDevice()) && !sanitized.hasSanitizedSource("hwaddress"))

	if forceDisable && p.prebuilt.SourceExists() {
		p.prebuiltCommonProperties.ForceDisable = true
		return true
	}
	return false
}

func (p *prebuiltCommon) InstallFilename() string {
	return proptools.StringDefault(p.prebuiltCommonProperties.Filename, p.BaseModuleName()+imageApexSuffix)
}

func (p *prebuiltCommon) Name() string {
	return p.prebuilt.Name(p.ModuleBase.Name())
}

func (p *prebuiltCommon) Overrides() []string {
	return p.prebuiltCommonProperties.Overrides
}

func (p *prebuiltCommon) installable() bool {
	return proptools.BoolDefault(p.prebuiltCommonProperties.Installable, true)
}

// To satisfy java.DexpreopterInterface
func (p *prebuiltCommon) IsInstallable() bool {
	return p.installable()
}

// initApexFilesForAndroidMk initializes the prebuiltCommon.requiredModuleNames field with the install only deps of the prebuilt apex
func (p *prebuiltCommon) initApexFilesForAndroidMk(ctx android.ModuleContext) {
	// If this apex contains a system server jar, then the dexpreopt artifacts should be added as required
	for _, install := range p.Dexpreopter.DexpreoptBuiltInstalledForApex() {
		p.requiredModuleNames = append(p.requiredModuleNames, install.FullModuleName())
	}
}

// If this prebuilt has system server jar, create the rules to dexpreopt it and install it alongside the prebuilt apex
func (p *prebuiltCommon) dexpreoptSystemServerJars(ctx android.ModuleContext, di *android.DeapexerInfo) {
	if di == nil {
		return
	}
	// If this prebuilt apex has not been selected, return
	if p.IsHideFromMake() {
		return
	}
	// Use apex_name to determine the api domain of this prebuilt apex
	apexName := p.ApexVariationName()
	// TODO: do not compute twice
	dc := dexpreopt.GetGlobalConfig(ctx)
	systemServerJarList := dc.AllApexSystemServerJars(ctx)

	for i := 0; i < systemServerJarList.Len(); i++ {
		sscpApex := systemServerJarList.Apex(i)
		sscpJar := systemServerJarList.Jar(i)
		if apexName != sscpApex {
			continue
		}
		p.Dexpreopter.DexpreoptPrebuiltApexSystemServerJars(ctx, sscpJar, di)
	}
}

func (p *prebuiltCommon) addRequiredModules(entries *android.AndroidMkEntries) {
	entries.AddStrings("LOCAL_REQUIRED_MODULES", p.requiredModuleNames...)
}

func (p *prebuiltCommon) AndroidMkEntries() []android.AndroidMkEntries {
	entriesList := []android.AndroidMkEntries{
		{
			Class:         "ETC",
			OutputFile:    android.OptionalPathForPath(p.outputApex),
			Include:       "$(BUILD_PREBUILT)",
			Host_required: p.hostRequired,
			ExtraEntries: []android.AndroidMkExtraEntriesFunc{
				func(ctx android.AndroidMkExtraEntriesContext, entries *android.AndroidMkEntries) {
					entries.SetString("LOCAL_MODULE_PATH", p.installDir.String())
					entries.SetString("LOCAL_MODULE_STEM", p.installFilename)
					entries.SetPath("LOCAL_SOONG_INSTALLED_MODULE", p.installedFile)
					entries.SetString("LOCAL_SOONG_INSTALL_PAIRS", p.outputApex.String()+":"+p.installedFile.String())
					entries.AddStrings("LOCAL_SOONG_INSTALL_SYMLINKS", p.compatSymlinks.Strings()...)
					entries.SetBoolIfTrue("LOCAL_UNINSTALLABLE_MODULE", !p.installable())
					entries.AddStrings("LOCAL_OVERRIDES_MODULES", p.prebuiltCommonProperties.Overrides...)
					entries.SetString("LOCAL_APEX_KEY_PATH", p.apexKeysPath.String())
					p.addRequiredModules(entries)
				},
			},
		},
	}

	// Add the dexpreopt artifacts to androidmk
	for _, install := range p.Dexpreopter.DexpreoptBuiltInstalledForApex() {
		entriesList = append(entriesList, install.ToMakeEntries())
	}

	return entriesList
}

func (p *prebuiltCommon) hasExportedDeps() bool {
	return len(p.prebuiltCommonProperties.Exported_bootclasspath_fragments) > 0 ||
		len(p.prebuiltCommonProperties.Exported_systemserverclasspath_fragments) > 0
}

// prebuiltApexContentsDeps adds dependencies onto the prebuilt apex module's contents.
func (p *prebuiltCommon) prebuiltApexContentsDeps(ctx android.BottomUpMutatorContext) {
	module := ctx.Module()

	for _, dep := range p.prebuiltCommonProperties.Exported_bootclasspath_fragments {
		prebuiltDep := android.PrebuiltNameFromSource(dep)
		ctx.AddDependency(module, exportedBootclasspathFragmentTag, prebuiltDep)
	}

	for _, dep := range p.prebuiltCommonProperties.Exported_systemserverclasspath_fragments {
		prebuiltDep := android.PrebuiltNameFromSource(dep)
		ctx.AddDependency(module, exportedSystemserverclasspathFragmentTag, prebuiltDep)
	}
}

// Implements android.DepInInSameApex
func (p *prebuiltCommon) DepIsInSameApex(ctx android.BaseModuleContext, dep android.Module) bool {
	tag := ctx.OtherModuleDependencyTag(dep)
	_, ok := tag.(exportedDependencyTag)
	return ok
}

// apexInfoMutator marks any modules for which this apex exports a file as requiring an apex
// specific variant and checks that they are supported.
//
// The apexMutator will ensure that the ApexInfo objects passed to BuildForApex(ApexInfo) are
// associated with the apex specific variant using the ApexInfoProvider for later retrieval.
//
// Unlike the source apex module type the prebuilt_apex module type cannot share compatible variants
// across prebuilt_apex modules. That is because there is no way to determine whether two
// prebuilt_apex modules that export files for the same module are compatible. e.g. they could have
// been built from different source at different times or they could have been built with different
// build options that affect the libraries.
//
// While it may be possible to provide sufficient information to determine whether two prebuilt_apex
// modules were compatible it would be a lot of work and would not provide much benefit for a couple
// of reasons:
//   - The number of prebuilt_apex modules that will be exporting files for the same module will be
//     low as the prebuilt_apex only exports files for the direct dependencies that require it and
//     very few modules are direct dependencies of multiple prebuilt_apex modules, e.g. there are a
//     few com.android.art* apex files that contain the same contents and could export files for the
//     same modules but only one of them needs to do so. Contrast that with source apex modules which
//     need apex specific variants for every module that contributes code to the apex, whether direct
//     or indirect.
//   - The build cost of a prebuilt_apex variant is generally low as at worst it will involve some
//     extra copying of files. Contrast that with source apex modules that has to build each variant
//     from source.
func (p *prebuiltCommon) apexInfoMutator(mctx android.TopDownMutatorContext) {

	// Collect direct dependencies into contents.
	contents := make(map[string]android.ApexMembership)

	// Collect the list of dependencies.
	var dependencies []android.ApexModule
	mctx.WalkDeps(func(child, parent android.Module) bool {
		// If the child is not in the same apex as the parent then exit immediately and do not visit
		// any of the child's dependencies.
		if !android.IsDepInSameApex(mctx, parent, child) {
			return false
		}

		tag := mctx.OtherModuleDependencyTag(child)
		depName := mctx.OtherModuleName(child)
		if exportedTag, ok := tag.(exportedDependencyTag); ok {
			propertyName := exportedTag.name

			// It is an error if the other module is not a prebuilt.
			if !android.IsModulePrebuilt(child) {
				mctx.PropertyErrorf(propertyName, "%q is not a prebuilt module", depName)
				return false
			}

			// It is an error if the other module is not an ApexModule.
			if _, ok := child.(android.ApexModule); !ok {
				mctx.PropertyErrorf(propertyName, "%q is not usable within an apex", depName)
				return false
			}
		}

		// Ignore any modules that do not implement ApexModule as they cannot have an APEX specific
		// variant.
		if _, ok := child.(android.ApexModule); !ok {
			return false
		}

		// Strip off the prebuilt_ prefix if present before storing content to ensure consistent
		// behavior whether there is a corresponding source module present or not.
		depName = android.RemoveOptionalPrebuiltPrefix(depName)

		// Remember if this module was added as a direct dependency.
		direct := parent == mctx.Module()
		contents[depName] = contents[depName].Add(direct)

		// Add the module to the list of dependencies that need to have an APEX variant.
		dependencies = append(dependencies, child.(android.ApexModule))

		return true
	})

	// Create contents for the prebuilt_apex and store it away for later use.
	apexContents := android.NewApexContents(contents)
	android.SetProvider(mctx, android.ApexBundleInfoProvider, android.ApexBundleInfo{
		Contents: apexContents,
	})

	// Create an ApexInfo for the prebuilt_apex.
	apexVariationName := p.ApexVariationName()
	apexInfo := android.ApexInfo{
		ApexVariationName: apexVariationName,
		InApexVariants:    []string{apexVariationName},
		InApexModules:     []string{p.BaseModuleName()}, // BaseModuleName() to avoid the prebuilt_ prefix.
		ApexContents:      []*android.ApexContents{apexContents},
		ForPrebuiltApex:   true,
	}

	// Mark the dependencies of this module as requiring a variant for this module.
	for _, am := range dependencies {
		am.BuildForApex(apexInfo)
	}
}

type Prebuilt struct {
	prebuiltCommon

	properties PrebuiltProperties

	inputApex android.Path

	provenanceMetaDataFile android.OutputPath
}

type ApexFileProperties struct {
	// the path to the prebuilt .apex file to import.
	//
	// This cannot be marked as `android:"arch_variant"` because the `prebuilt_apex` is only mutated
	// for android_common. That is so that it will have the same arch variant as, and so be compatible
	// with, the source `apex` module type that it replaces.
	Src  proptools.Configurable[string] `android:"path,replace_instead_of_append"`
	Arch struct {
		Arm struct {
			Src *string `android:"path"`
		}
		Arm64 struct {
			Src *string `android:"path"`
		}
		Riscv64 struct {
			Src *string `android:"path"`
		}
		X86 struct {
			Src *string `android:"path"`
		}
		X86_64 struct {
			Src *string `android:"path"`
		}
	}
}

// prebuiltApexSelector selects the correct prebuilt APEX file for the build target.
//
// The ctx parameter can be for any module not just the prebuilt module so care must be taken not
// to use methods on it that are specific to the current module.
//
// See the ApexFileProperties.Src property.
func (p *ApexFileProperties) prebuiltApexSelector(ctx android.BaseModuleContext, prebuilt android.Module) string {
	multiTargets := prebuilt.MultiTargets()
	if len(multiTargets) != 1 {
		ctx.OtherModuleErrorf(prebuilt, "compile_multilib shouldn't be \"both\" for prebuilt_apex")
		return ""
	}
	var src string
	switch multiTargets[0].Arch.ArchType {
	case android.Arm:
		src = String(p.Arch.Arm.Src)
	case android.Arm64:
		src = String(p.Arch.Arm64.Src)
	case android.Riscv64:
		src = String(p.Arch.Riscv64.Src)
		// HACK: fall back to arm64 prebuilts, the riscv64 ones don't exist yet.
		if src == "" {
			src = String(p.Arch.Arm64.Src)
		}
	case android.X86:
		src = String(p.Arch.X86.Src)
	case android.X86_64:
		src = String(p.Arch.X86_64.Src)
	}
	if src == "" {
		src = p.Src.GetOrDefault(ctx, "")
	}

	if src == "" {
		if ctx.Config().AllowMissingDependencies() {
			ctx.AddMissingDependencies([]string{ctx.OtherModuleName(prebuilt)})
		} else {
			ctx.OtherModuleErrorf(prebuilt, "prebuilt_apex does not support %q", multiTargets[0].Arch.String())
		}
		// Drop through to return an empty string as the src (instead of nil) to avoid the prebuilt
		// logic from reporting a more general, less useful message.
	}

	return src
}

type PrebuiltProperties struct {
	ApexFileProperties

	PrebuiltCommonProperties
}

func (a *Prebuilt) hasSanitizedSource(sanitizer string) bool {
	return false
}

// prebuilt_apex imports an `.apex` file into the build graph as if it was built with apex.
func PrebuiltFactory() android.Module {
	module := &Prebuilt{}
	module.AddProperties(&module.properties)
	module.prebuiltCommon.prebuiltCommonProperties = &module.properties.PrebuiltCommonProperties

	// init the module as a prebuilt
	// even though this module type has srcs, use `InitPrebuiltModuleWithoutSrcs`, since the existing
	// InitPrebuiltModule* are not friendly with Sources of Configurable type.
	// The actual src will be evaluated in GenerateAndroidBuildActions.
	android.InitPrebuiltModuleWithoutSrcs(module)
	android.InitAndroidMultiTargetsArchModule(module, android.DeviceSupported, android.MultilibCommon)

	return module
}

func (p *prebuiltCommon) getDeapexerPropertiesIfNeeded(ctx android.ModuleContext) DeapexerProperties {
	// Compute the deapexer properties from the transitive dependencies of this module.
	commonModules := []string{}
	dexpreoptProfileGuidedModules := []string{}
	exportedFiles := []string{}
	ctx.WalkDeps(func(child, parent android.Module) bool {
		tag := ctx.OtherModuleDependencyTag(child)

		// If the child is not in the same apex as the parent then ignore it and all its children.
		if !android.IsDepInSameApex(ctx, parent, child) {
			return false
		}

		name := java.ModuleStemForDeapexing(child)
		if _, ok := tag.(android.RequiresFilesFromPrebuiltApexTag); ok {
			commonModules = append(commonModules, name)

			extract := child.(android.RequiredFilesFromPrebuiltApex)
			requiredFiles := extract.RequiredFilesFromPrebuiltApex(ctx)
			exportedFiles = append(exportedFiles, requiredFiles...)

			if extract.UseProfileGuidedDexpreopt() {
				dexpreoptProfileGuidedModules = append(dexpreoptProfileGuidedModules, name)
			}

			// Visit the dependencies of this module just in case they also require files from the
			// prebuilt apex.
			return true
		}

		return false
	})

	// Create properties for deapexer module.
	deapexerProperties := DeapexerProperties{
		// Remove any duplicates from the common modules lists as a module may be included via a direct
		// dependency as well as transitive ones.
		CommonModules:                 android.SortedUniqueStrings(commonModules),
		DexpreoptProfileGuidedModules: android.SortedUniqueStrings(dexpreoptProfileGuidedModules),
	}

	// Populate the exported files property in a fixed order.
	deapexerProperties.ExportedFiles = android.SortedUniqueStrings(exportedFiles)
	return deapexerProperties
}

func prebuiltApexExportedModuleName(ctx android.BottomUpMutatorContext, name string) string {
	// The prebuilt_apex should be depending on prebuilt modules but as this runs after
	// prebuilt_rename the prebuilt module may or may not be using the prebuilt_ prefixed named. So,
	// check to see if the prefixed name is in use first, if it is then use that, otherwise assume
	// the unprefixed name is the one to use. If the unprefixed one turns out to be a source module
	// and not a renamed prebuilt module then that will be detected and reported as an error when
	// processing the dependency in ApexInfoMutator().
	prebuiltName := android.PrebuiltNameFromSource(name)
	if ctx.OtherModuleExists(prebuiltName) {
		name = prebuiltName
	}
	return name
}

type exportedDependencyTag struct {
	blueprint.BaseDependencyTag
	name string
}

// Mark this tag so dependencies that use it are excluded from visibility enforcement.
//
// This does allow any prebuilt_apex to reference any module which does open up a small window for
// restricted visibility modules to be referenced from the wrong prebuilt_apex. However, doing so
// avoids opening up a much bigger window by widening the visibility of modules that need files
// provided by the prebuilt_apex to include all the possible locations they may be defined, which
// could include everything below vendor/.
//
// A prebuilt_apex that references a module via this tag will have to contain the appropriate files
// corresponding to that module, otherwise it will fail when attempting to retrieve the files from
// the .apex file. It will also have to be included in the module's apex_available property too.
// That makes it highly unlikely that a prebuilt_apex would reference a restricted module
// incorrectly.
func (t exportedDependencyTag) ExcludeFromVisibilityEnforcement() {}

func (t exportedDependencyTag) RequiresFilesFromPrebuiltApex() {}

var _ android.RequiresFilesFromPrebuiltApexTag = exportedDependencyTag{}

var (
	exportedBootclasspathFragmentTag         = exportedDependencyTag{name: "exported_bootclasspath_fragments"}
	exportedSystemserverclasspathFragmentTag = exportedDependencyTag{name: "exported_systemserverclasspath_fragments"}
)

func (p *Prebuilt) ComponentDepsMutator(ctx android.BottomUpMutatorContext) {
	p.prebuiltApexContentsDeps(ctx)
}

var _ ApexInfoMutator = (*Prebuilt)(nil)

func (p *Prebuilt) ApexInfoMutator(mctx android.TopDownMutatorContext) {
	p.apexInfoMutator(mctx)
}

// creates the build rules to deapex the prebuilt, and returns a deapexerInfo
func (p *prebuiltCommon) getDeapexerInfo(ctx android.ModuleContext, apexFile android.Path) *android.DeapexerInfo {
	if !p.hasExportedDeps() {
		// nothing to do
		return nil
	}
	deapexerProps := p.getDeapexerPropertiesIfNeeded(ctx)
	return deapex(ctx, apexFile, deapexerProps)
}

// Set a provider containing information about the jars and .prof provided by the apex
// Apexes built from prebuilts retrieve this information by visiting its internal deapexer module
// Used by dex_bootjars to generate the boot image
func (p *prebuiltCommon) provideApexExportsInfo(ctx android.ModuleContext, di *android.DeapexerInfo) {
	if di == nil {
		return
	}
	javaModuleToDexPath := map[string]android.Path{}
	for _, commonModule := range di.GetExportedModuleNames() {
		if dex := di.PrebuiltExportPath(java.ApexRootRelativePathToJavaLib(commonModule)); dex != nil {
			javaModuleToDexPath[commonModule] = dex
		}
	}

	exports := android.ApexExportsInfo{
		ApexName:                      p.ApexVariationName(),
		ProfilePathOnHost:             di.PrebuiltExportPath(java.ProfileInstallPathInApex),
		LibraryNameToDexJarPathOnHost: javaModuleToDexPath,
	}
	android.SetProvider(ctx, android.ApexExportsInfoProvider, exports)
}

// Set prebuiltInfoProvider. This will be used by `apex_prebuiltinfo_singleton` to print out a metadata file
// with information about whether source or prebuilt of an apex was used during the build.
func (p *prebuiltCommon) providePrebuiltInfo(ctx android.ModuleContext) {
	info := android.PrebuiltInfo{
		Name:        p.BaseModuleName(),
		Is_prebuilt: true,
	}
	// If Prebuilt_info information is available in the soong module definition, add it to prebuilt_info.json.
	if p.prebuiltCommonProperties.Prebuilt_info != nil {
		info.Prebuilt_info_file_path = android.PathForModuleSrc(ctx, *p.prebuiltCommonProperties.Prebuilt_info).String()
	}
	android.SetProvider(ctx, android.PrebuiltInfoProvider, info)
}

// Uses an object provided by its deps to validate that the contents of bcpf have been added to the global
// PRODUCT_APEX_BOOT_JARS
// This validation will only run on the apex which is active for this product/release_config
func validateApexClasspathFragments(ctx android.ModuleContext) {
	ctx.VisitDirectDeps(func(m android.Module) {
		if info, exists := android.OtherModuleProvider(ctx, m, java.ClasspathFragmentValidationInfoProvider); exists {
			ctx.ModuleErrorf("%s in contents of %s must also be declared in PRODUCT_APEX_BOOT_JARS", info.UnknownJars, info.ClasspathFragmentModuleName)
		}
	})
}

func (p *Prebuilt) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	// Validate contents of classpath fragments
	if !p.IsHideFromMake() {
		validateApexClasspathFragments(ctx)
	}

	p.apexKeysPath = writeApexKeys(ctx, p)
	// TODO(jungjw): Check the key validity.
	p.inputApex = android.PathForModuleSrc(ctx, p.properties.prebuiltApexSelector(ctx, ctx.Module()))
	p.installDir = android.PathForModuleInstall(ctx, "apex")
	p.installFilename = p.InstallFilename()
	if !strings.HasSuffix(p.installFilename, imageApexSuffix) {
		ctx.ModuleErrorf("filename should end in %s for prebuilt_apex", imageApexSuffix)
	}
	p.outputApex = android.PathForModuleOut(ctx, p.installFilename)
	ctx.Build(pctx, android.BuildParams{
		Rule:   android.Cp,
		Input:  p.inputApex,
		Output: p.outputApex,
	})

	if p.prebuiltCommon.checkForceDisable(ctx) {
		p.HideFromMake()
		return
	}

	deapexerInfo := p.getDeapexerInfo(ctx, p.inputApex)

	// dexpreopt any system server jars if present
	p.dexpreoptSystemServerJars(ctx, deapexerInfo)

	// provide info used for generating the boot image
	p.provideApexExportsInfo(ctx, deapexerInfo)

	p.providePrebuiltInfo(ctx)

	// Save the files that need to be made available to Make.
	p.initApexFilesForAndroidMk(ctx)

	// in case that prebuilt_apex replaces source apex (using prefer: prop)
	p.compatSymlinks = makeCompatSymlinks(p.BaseModuleName(), ctx)
	// or that prebuilt_apex overrides other apexes (using overrides: prop)
	for _, overridden := range p.prebuiltCommonProperties.Overrides {
		p.compatSymlinks = append(p.compatSymlinks, makeCompatSymlinks(overridden, ctx)...)
	}

	if p.installable() {
		p.installedFile = ctx.InstallFile(p.installDir, p.installFilename, p.inputApex, p.compatSymlinks...)
		p.provenanceMetaDataFile = provenance.GenerateArtifactProvenanceMetaData(ctx, p.inputApex, p.installedFile)
	}

	ctx.SetOutputFiles(android.Paths{p.outputApex}, "")
}

func (p *Prebuilt) ProvenanceMetaDataFile() android.OutputPath {
	return p.provenanceMetaDataFile
}

// prebuiltApexExtractorModule is a private module type that is only created by the prebuilt_apex
// module. It extracts the correct apex to use and makes it available for use by apex_set.
type prebuiltApexExtractorModule struct {
	android.ModuleBase

	properties ApexExtractorProperties

	extractedApex android.WritablePath
}

// extract registers the build actions to extract an apex from .apks file
// returns the path of the extracted apex
func extract(ctx android.ModuleContext, apexSet android.Path, prerelease *bool) android.Path {
	defaultAllowPrerelease := ctx.Config().IsEnvTrue("SOONG_ALLOW_PRERELEASE_APEXES")
	extractedApex := android.PathForModuleOut(ctx, "extracted", apexSet.Base())
	// Filter out NativeBridge archs (b/260115309)
	abis := java.SupportedAbis(ctx, true)
	ctx.Build(pctx,
		android.BuildParams{
			Rule:        extractMatchingApex,
			Description: "Extract an apex from an apex set",
			Inputs:      android.Paths{apexSet},
			Output:      extractedApex,
			Args: map[string]string{
				"abis":              strings.Join(abis, ","),
				"allow-prereleased": strconv.FormatBool(proptools.BoolDefault(prerelease, defaultAllowPrerelease)),
				"sdk-version":       ctx.Config().PlatformSdkVersion().String(),
				"skip-sdk-check":    strconv.FormatBool(ctx.Config().IsEnvTrue("SOONG_SKIP_APPSET_SDK_CHECK")),
			},
		},
	)
	return extractedApex
}

type ApexSet struct {
	prebuiltCommon

	properties ApexSetProperties
}

type ApexExtractorProperties struct {
	// the .apks file path that contains prebuilt apex files to be extracted.
	Set *string `android:"path"`

	Sanitized struct {
		None struct {
			Set *string `android:"path"`
		}
		Address struct {
			Set *string `android:"path"`
		}
		Hwaddress struct {
			Set *string `android:"path"`
		}
	}

	// apexes in this set use prerelease SDK version
	Prerelease *bool
}

func (e *ApexExtractorProperties) prebuiltSrcs(ctx android.BaseModuleContext) []string {
	var srcs []string
	if e.Set != nil {
		srcs = append(srcs, *e.Set)
	}

	sanitizers := ctx.Config().SanitizeDevice()

	if android.InList("address", sanitizers) && e.Sanitized.Address.Set != nil {
		srcs = append(srcs, *e.Sanitized.Address.Set)
	} else if android.InList("hwaddress", sanitizers) && e.Sanitized.Hwaddress.Set != nil {
		srcs = append(srcs, *e.Sanitized.Hwaddress.Set)
	} else if e.Sanitized.None.Set != nil {
		srcs = append(srcs, *e.Sanitized.None.Set)
	}

	return srcs
}

type ApexSetProperties struct {
	ApexExtractorProperties

	PrebuiltCommonProperties
}

func (a *ApexSet) hasSanitizedSource(sanitizer string) bool {
	if sanitizer == "address" {
		return a.properties.Sanitized.Address.Set != nil
	}
	if sanitizer == "hwaddress" {
		return a.properties.Sanitized.Hwaddress.Set != nil
	}

	return false
}

// prebuilt_apex imports an `.apex` file into the build graph as if it was built with apex.
func apexSetFactory() android.Module {
	module := &ApexSet{}
	module.AddProperties(&module.properties)
	module.prebuiltCommon.prebuiltCommonProperties = &module.properties.PrebuiltCommonProperties

	// init the module as a prebuilt
	// even though this module type has srcs, use `InitPrebuiltModuleWithoutSrcs`, since the existing
	// InitPrebuiltModule* are not friendly with Sources of Configurable type.
	// The actual src will be evaluated in GenerateAndroidBuildActions.
	android.InitPrebuiltModuleWithoutSrcs(module)
	android.InitAndroidMultiTargetsArchModule(module, android.DeviceSupported, android.MultilibCommon)

	return module
}

func (a *ApexSet) ComponentDepsMutator(ctx android.BottomUpMutatorContext) {
	a.prebuiltApexContentsDeps(ctx)
}

var _ ApexInfoMutator = (*ApexSet)(nil)

func (a *ApexSet) ApexInfoMutator(mctx android.TopDownMutatorContext) {
	a.apexInfoMutator(mctx)
}

func (a *ApexSet) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	// Validate contents of classpath fragments
	if !a.IsHideFromMake() {
		validateApexClasspathFragments(ctx)
	}

	a.apexKeysPath = writeApexKeys(ctx, a)
	a.installFilename = a.InstallFilename()
	if !strings.HasSuffix(a.installFilename, imageApexSuffix) && !strings.HasSuffix(a.installFilename, imageCapexSuffix) {
		ctx.ModuleErrorf("filename should end in %s or %s for apex_set", imageApexSuffix, imageCapexSuffix)
	}

	var apexSet android.Path
	if srcs := a.properties.prebuiltSrcs(ctx); len(srcs) == 1 {
		apexSet = android.PathForModuleSrc(ctx, srcs[0])
	} else {
		ctx.ModuleErrorf("Expected exactly one source apex_set file, found %v\n", srcs)
	}

	extractedApex := extract(ctx, apexSet, a.properties.Prerelease)

	a.outputApex = android.PathForModuleOut(ctx, a.installFilename)

	// Build the output APEX. If compression is not enabled, make sure the output is not compressed even if the input is compressed
	buildRule := android.Cp
	if !ctx.Config().ApexCompressionEnabled() {
		buildRule = decompressApex
	}
	ctx.Build(pctx, android.BuildParams{
		Rule:   buildRule,
		Input:  extractedApex,
		Output: a.outputApex,
	})

	if a.prebuiltCommon.checkForceDisable(ctx) {
		a.HideFromMake()
		return
	}

	deapexerInfo := a.getDeapexerInfo(ctx, extractedApex)

	// dexpreopt any system server jars if present
	a.dexpreoptSystemServerJars(ctx, deapexerInfo)

	// provide info used for generating the boot image
	a.provideApexExportsInfo(ctx, deapexerInfo)

	a.providePrebuiltInfo(ctx)

	// Save the files that need to be made available to Make.
	a.initApexFilesForAndroidMk(ctx)

	a.installDir = android.PathForModuleInstall(ctx, "apex")
	if a.installable() {
		a.installedFile = ctx.InstallFile(a.installDir, a.installFilename, a.outputApex)
	}

	// in case that apex_set replaces source apex (using prefer: prop)
	a.compatSymlinks = makeCompatSymlinks(a.BaseModuleName(), ctx)
	// or that apex_set overrides other apexes (using overrides: prop)
	for _, overridden := range a.prebuiltCommonProperties.Overrides {
		a.compatSymlinks = append(a.compatSymlinks, makeCompatSymlinks(overridden, ctx)...)
	}

	ctx.SetOutputFiles(android.Paths{a.outputApex}, "")
}

type systemExtContext struct {
	android.ModuleContext
}

func (*systemExtContext) SystemExtSpecific() bool {
	return true
}
