// Copyright 2020 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License")
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
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/blueprint"
	"github.com/google/blueprint/proptools"
)

// PackagingSpec abstracts a request to place a built artifact at a certain path in a package. A
// package can be the traditional <partition>.img, but isn't limited to those. Other examples could
// be a new filesystem image that is a subset of system.img (e.g. for an Android-like mini OS
// running on a VM), or a zip archive for some of the host tools.
type PackagingSpec struct {
	// Path relative to the root of the package
	relPathInPackage string

	// The path to the built artifact
	srcPath Path

	// If this is not empty, then relPathInPackage should be a symlink to this target. (Then
	// srcPath is of course ignored.)
	symlinkTarget string

	// Whether relPathInPackage should be marked as executable or not
	executable bool

	effectiveLicenseFiles *Paths

	partition string

	// Whether this packaging spec represents an installation of the srcPath (i.e. this struct
	// is created via InstallFile or InstallSymlink) or a simple packaging (i.e. created via
	// PackageFile).
	skipInstall bool

	// Paths of aconfig files for the built artifact
	aconfigPaths *Paths

	// ArchType of the module which produced this packaging spec
	archType ArchType

	// List of module names that this packaging spec overrides
	overrides *[]string

	// Name of the module where this packaging spec is output of
	owner string
}

func (p *PackagingSpec) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := errors.Join(encoder.Encode(p.relPathInPackage), encoder.Encode(p.srcPath),
		encoder.Encode(p.symlinkTarget), encoder.Encode(p.executable),
		encoder.Encode(p.effectiveLicenseFiles), encoder.Encode(p.partition),
		encoder.Encode(p.skipInstall), encoder.Encode(p.aconfigPaths),
		encoder.Encode(p.archType))
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (p *PackagingSpec) GobDecode(data []byte) error {
	r := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(r)
	err := errors.Join(decoder.Decode(&p.relPathInPackage), decoder.Decode(&p.srcPath),
		decoder.Decode(&p.symlinkTarget), decoder.Decode(&p.executable),
		decoder.Decode(&p.effectiveLicenseFiles), decoder.Decode(&p.partition),
		decoder.Decode(&p.skipInstall), decoder.Decode(&p.aconfigPaths),
		decoder.Decode(&p.archType))
	if err != nil {
		return err
	}

	return nil
}

func (p *PackagingSpec) Equals(other *PackagingSpec) bool {
	if other == nil {
		return false
	}
	if p.relPathInPackage != other.relPathInPackage {
		return false
	}
	if p.srcPath != other.srcPath || p.symlinkTarget != other.symlinkTarget {
		return false
	}
	if p.executable != other.executable {
		return false
	}
	if p.partition != other.partition {
		return false
	}
	return true
}

// Get file name of installed package
func (p *PackagingSpec) FileName() string {
	if p.relPathInPackage != "" {
		return filepath.Base(p.relPathInPackage)
	}

	return ""
}

// Path relative to the root of the package
func (p *PackagingSpec) RelPathInPackage() string {
	return p.relPathInPackage
}

func (p *PackagingSpec) SetRelPathInPackage(relPathInPackage string) {
	p.relPathInPackage = relPathInPackage
}

func (p *PackagingSpec) EffectiveLicenseFiles() Paths {
	if p.effectiveLicenseFiles == nil {
		return Paths{}
	}
	return *p.effectiveLicenseFiles
}

func (p *PackagingSpec) Partition() string {
	return p.partition
}

func (p *PackagingSpec) SkipInstall() bool {
	return p.skipInstall
}

// Paths of aconfig files for the built artifact
func (p *PackagingSpec) GetAconfigPaths() Paths {
	return *p.aconfigPaths
}

type PackageModule interface {
	Module
	packagingBase() *PackagingBase

	// AddDeps adds dependencies to the `deps` modules. This should be called in DepsMutator.
	// When adding the dependencies, depTag is used as the tag. If `deps` modules are meant to
	// be copied to a zip in CopyDepsToZip, `depTag` should implement PackagingItem marker interface.
	AddDeps(ctx BottomUpMutatorContext, depTag blueprint.DependencyTag)

	// GatherPackagingSpecs gathers PackagingSpecs of transitive dependencies.
	GatherPackagingSpecs(ctx ModuleContext) map[string]PackagingSpec
	GatherPackagingSpecsWithFilter(ctx ModuleContext, filter func(PackagingSpec) bool) map[string]PackagingSpec

	// CopyDepsToZip zips the built artifacts of the dependencies into the given zip file and
	// returns zip entries in it. This is expected to be called in GenerateAndroidBuildActions,
	// followed by a build rule that unzips it and creates the final output (img, zip, tar.gz,
	// etc.) from the extracted files
	CopyDepsToZip(ctx ModuleContext, specs map[string]PackagingSpec, zipOut WritablePath) []string
}

// PackagingBase provides basic functionality for packaging dependencies. A module is expected to
// include this struct and call InitPackageModule.
type PackagingBase struct {
	properties PackagingProperties

	// Allows this module to skip missing dependencies. In most cases, this is not required, but
	// for rare cases like when there's a dependency to a module which exists in certain repo
	// checkouts, this is needed.
	IgnoreMissingDependencies bool

	// If this is set to true by a module type inheriting PackagingBase, the deps property
	// collects the first target only even with compile_multilib: true.
	DepsCollectFirstTargetOnly bool
}

type depsProperty struct {
	// Modules to include in this package
	Deps proptools.Configurable[[]string] `android:"arch_variant"`
}

type packagingMultilibProperties struct {
	First    depsProperty `android:"arch_variant"`
	Common   depsProperty `android:"arch_variant"`
	Lib32    depsProperty `android:"arch_variant"`
	Lib64    depsProperty `android:"arch_variant"`
	Both     depsProperty `android:"arch_variant"`
	Prefer32 depsProperty `android:"arch_variant"`
}

type packagingArchProperties struct {
	Arm64  depsProperty
	Arm    depsProperty
	X86_64 depsProperty
	X86    depsProperty
}

type PackagingProperties struct {
	Deps     proptools.Configurable[[]string] `android:"arch_variant"`
	Multilib packagingMultilibProperties      `android:"arch_variant"`
	Arch     packagingArchProperties
}

func InitPackageModule(p PackageModule) {
	base := p.packagingBase()
	p.AddProperties(&base.properties)
}

func (p *PackagingBase) packagingBase() *PackagingBase {
	return p
}

// From deps and multilib.*.deps, select the dependencies that are for the given arch deps is for
// the current archicture when this module is not configured for multi target. When configured for
// multi target, deps is selected for each of the targets and is NOT selected for the current
// architecture which would be Common.
func (p *PackagingBase) getDepsForArch(ctx BaseModuleContext, arch ArchType) []string {
	get := func(prop proptools.Configurable[[]string]) []string {
		return prop.GetOrDefault(ctx, nil)
	}

	var ret []string
	if arch == ctx.Target().Arch.ArchType && len(ctx.MultiTargets()) == 0 {
		ret = append(ret, get(p.properties.Deps)...)
	} else if arch.Multilib == "lib32" {
		ret = append(ret, get(p.properties.Multilib.Lib32.Deps)...)
		// multilib.prefer32.deps are added for lib32 only when they support 32-bit arch
		for _, dep := range get(p.properties.Multilib.Prefer32.Deps) {
			if checkIfOtherModuleSupportsLib32(ctx, dep) {
				ret = append(ret, dep)
			}
		}
	} else if arch.Multilib == "lib64" {
		ret = append(ret, get(p.properties.Multilib.Lib64.Deps)...)
		// multilib.prefer32.deps are added for lib64 only when they don't support 32-bit arch
		for _, dep := range get(p.properties.Multilib.Prefer32.Deps) {
			if !checkIfOtherModuleSupportsLib32(ctx, dep) {
				ret = append(ret, dep)
			}
		}
	} else if arch == Common {
		ret = append(ret, get(p.properties.Multilib.Common.Deps)...)
	}

	if p.DepsCollectFirstTargetOnly {
		if len(get(p.properties.Multilib.First.Deps)) > 0 {
			ctx.PropertyErrorf("multilib.first.deps", "not supported. use \"deps\" instead")
		}
		for i, t := range ctx.MultiTargets() {
			if t.Arch.ArchType == arch {
				ret = append(ret, get(p.properties.Multilib.Both.Deps)...)
				if i == 0 {
					ret = append(ret, get(p.properties.Deps)...)
				}
			}
		}
	} else {
		if len(get(p.properties.Multilib.Both.Deps)) > 0 {
			ctx.PropertyErrorf("multilib.both.deps", "not supported. use \"deps\" instead")
		}
		for i, t := range ctx.MultiTargets() {
			if t.Arch.ArchType == arch {
				ret = append(ret, get(p.properties.Deps)...)
				if i == 0 {
					ret = append(ret, get(p.properties.Multilib.First.Deps)...)
				}
			}
		}
	}

	if ctx.Arch().ArchType == Common {
		switch arch {
		case Arm64:
			ret = append(ret, get(p.properties.Arch.Arm64.Deps)...)
		case Arm:
			ret = append(ret, get(p.properties.Arch.Arm.Deps)...)
		case X86_64:
			ret = append(ret, get(p.properties.Arch.X86_64.Deps)...)
		case X86:
			ret = append(ret, get(p.properties.Arch.X86.Deps)...)
		}
	}

	return FirstUniqueStrings(ret)
}

func getSupportedTargets(ctx BaseModuleContext) []Target {
	var ret []Target
	// The current and the common OS targets are always supported
	ret = append(ret, ctx.Target())
	if ctx.Arch().ArchType != Common {
		ret = append(ret, Target{Os: ctx.Os(), Arch: Arch{ArchType: Common}})
	}
	// If this module is configured for multi targets, those should be supported as well
	ret = append(ret, ctx.MultiTargets()...)
	return ret
}

// getLib32Target returns the 32-bit target from the list of targets this module supports. If this
// module doesn't support 32-bit target, nil is returned.
func getLib32Target(ctx BaseModuleContext) *Target {
	for _, t := range getSupportedTargets(ctx) {
		if t.Arch.ArchType.Multilib == "lib32" {
			return &t
		}
	}
	return nil
}

// checkIfOtherModuleSUpportsLib32 returns true if 32-bit variant of dep exists.
func checkIfOtherModuleSupportsLib32(ctx BaseModuleContext, dep string) bool {
	t := getLib32Target(ctx)
	if t == nil {
		// This packaging module doesn't support 32bit. No point of checking if dep supports 32-bit
		// or not.
		return false
	}
	return ctx.OtherModuleFarDependencyVariantExists(t.Variations(), dep)
}

// PackagingItem is a marker interface for dependency tags.
// Direct dependencies with a tag implementing PackagingItem are packaged in CopyDepsToZip().
type PackagingItem interface {
	// IsPackagingItem returns true if the dep is to be packaged
	IsPackagingItem() bool
}

// DepTag provides default implementation of PackagingItem interface.
// PackagingBase-derived modules can define their own dependency tag by embedding this, which
// can be passed to AddDeps() or AddDependencies().
type PackagingItemAlwaysDepTag struct {
}

// IsPackagingItem returns true if the dep is to be packaged
func (PackagingItemAlwaysDepTag) IsPackagingItem() bool {
	return true
}

// See PackageModule.AddDeps
func (p *PackagingBase) AddDeps(ctx BottomUpMutatorContext, depTag blueprint.DependencyTag) {
	for _, t := range getSupportedTargets(ctx) {
		for _, dep := range p.getDepsForArch(ctx, t.Arch.ArchType) {
			if p.IgnoreMissingDependencies && !ctx.OtherModuleExists(dep) {
				continue
			}
			ctx.AddFarVariationDependencies(t.Variations(), depTag, dep)
		}
	}
}

func (p *PackagingBase) GatherPackagingSpecsWithFilter(ctx ModuleContext, filter func(PackagingSpec) bool) map[string]PackagingSpec {
	// all packaging specs gathered from the dep.
	var all []PackagingSpec
	// list of module names overridden
	var overridden []string

	var arches []ArchType
	for _, target := range getSupportedTargets(ctx) {
		arches = append(arches, target.Arch.ArchType)
	}

	// filter out packaging specs for unsupported architecture
	filterArch := func(ps PackagingSpec) bool {
		for _, arch := range arches {
			if arch == ps.archType {
				return true
			}
		}
		return false
	}

	ctx.VisitDirectDeps(func(child Module) {
		if pi, ok := ctx.OtherModuleDependencyTag(child).(PackagingItem); !ok || !pi.IsPackagingItem() {
			return
		}
		for _, ps := range OtherModuleProviderOrDefault(
			ctx, child, InstallFilesProvider).TransitivePackagingSpecs.ToList() {
			if !filterArch(ps) {
				continue
			}

			if filter != nil {
				if !filter(ps) {
					continue
				}
			}
			all = append(all, ps)
			if ps.overrides != nil {
				overridden = append(overridden, *ps.overrides...)
			}
		}
	})

	// all minus packaging specs that are overridden
	var filtered []PackagingSpec
	for _, ps := range all {
		if ps.owner != "" && InList(ps.owner, overridden) {
			continue
		}
		filtered = append(filtered, ps)
	}

	m := make(map[string]PackagingSpec)
	for _, ps := range filtered {
		dstPath := ps.relPathInPackage
		if existingPs, ok := m[dstPath]; ok {
			if !existingPs.Equals(&ps) {
				ctx.ModuleErrorf("packaging conflict at %v:\n%v\n%v", dstPath, existingPs, ps)
			}
			continue
		}
		m[dstPath] = ps
	}
	return m
}

// See PackageModule.GatherPackagingSpecs
func (p *PackagingBase) GatherPackagingSpecs(ctx ModuleContext) map[string]PackagingSpec {
	return p.GatherPackagingSpecsWithFilter(ctx, nil)
}

// CopySpecsToDir is a helper that will add commands to the rule builder to copy the PackagingSpec
// entries into the specified directory.
func (p *PackagingBase) CopySpecsToDir(ctx ModuleContext, builder *RuleBuilder, specs map[string]PackagingSpec, dir WritablePath) (entries []string) {
	dirsToSpecs := make(map[WritablePath]map[string]PackagingSpec)
	dirsToSpecs[dir] = specs
	return p.CopySpecsToDirs(ctx, builder, dirsToSpecs)
}

// CopySpecsToDirs is a helper that will add commands to the rule builder to copy the PackagingSpec
// entries into corresponding directories.
func (p *PackagingBase) CopySpecsToDirs(ctx ModuleContext, builder *RuleBuilder, dirsToSpecs map[WritablePath]map[string]PackagingSpec) (entries []string) {
	empty := true
	for _, specs := range dirsToSpecs {
		if len(specs) > 0 {
			empty = false
			break
		}
	}
	if empty {
		return entries
	}

	seenDir := make(map[string]bool)
	preparerPath := PathForModuleOut(ctx, "preparer.sh")
	cmd := builder.Command().Tool(preparerPath)
	var sb strings.Builder
	sb.WriteString("set -e\n")

	dirs := make([]WritablePath, 0, len(dirsToSpecs))
	for dir, _ := range dirsToSpecs {
		dirs = append(dirs, dir)
	}
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].String() < dirs[j].String()
	})

	for _, dir := range dirs {
		specs := dirsToSpecs[dir]
		for _, k := range SortedKeys(specs) {
			ps := specs[k]
			destPath := filepath.Join(dir.String(), ps.relPathInPackage)
			destDir := filepath.Dir(destPath)
			entries = append(entries, ps.relPathInPackage)
			if _, ok := seenDir[destDir]; !ok {
				seenDir[destDir] = true
				sb.WriteString(fmt.Sprintf("mkdir -p %s\n", destDir))
			}
			if ps.symlinkTarget == "" {
				cmd.Implicit(ps.srcPath)
				sb.WriteString(fmt.Sprintf("cp %s %s\n", ps.srcPath, destPath))
			} else {
				sb.WriteString(fmt.Sprintf("ln -sf %s %s\n", ps.symlinkTarget, destPath))
			}
			if ps.executable {
				sb.WriteString(fmt.Sprintf("chmod a+x %s\n", destPath))
			}
		}
	}

	WriteExecutableFileRuleVerbatim(ctx, preparerPath, sb.String())

	return entries
}

// See PackageModule.CopyDepsToZip
func (p *PackagingBase) CopyDepsToZip(ctx ModuleContext, specs map[string]PackagingSpec, zipOut WritablePath) (entries []string) {
	builder := NewRuleBuilder(pctx, ctx)

	dir := PathForModuleOut(ctx, ".zip")
	builder.Command().Text("rm").Flag("-rf").Text(dir.String())
	builder.Command().Text("mkdir").Flag("-p").Text(dir.String())
	entries = p.CopySpecsToDir(ctx, builder, specs, dir)

	builder.Command().
		BuiltTool("soong_zip").
		FlagWithOutput("-o ", zipOut).
		FlagWithArg("-C ", dir.String()).
		Flag("-L 0"). // no compression because this will be unzipped soon
		FlagWithArg("-D ", dir.String())
	builder.Command().Text("rm").Flag("-rf").Text(dir.String())

	builder.Build("zip_deps", fmt.Sprintf("Zipping deps for %s", ctx.ModuleName()))
	return entries
}
