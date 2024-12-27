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
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/google/blueprint"
)

// ----------------------------------------------------------------------------
// Start of the definitions of exception functions and the lookup table.
//
// Functions cannot be used as a value passed in providers, because functions are not
// hashable. As a workaround, the [exceptionHandleFuncLabel] enum values are passed using providers,
// and the corresponding functions are called from [exceptionHandleFunctionsTable] map.
// ----------------------------------------------------------------------------

type exceptionHandleFunc func(ModuleContext, Module, Module) bool

type StubsAvailableModule interface {
	IsStubsModule() bool
}

// Returns true if the dependency module is a stubs module
var depIsStubsModule exceptionHandleFunc = func(_ ModuleContext, _, dep Module) bool {
	if stubsModule, ok := dep.(StubsAvailableModule); ok {
		return stubsModule.IsStubsModule()
	}
	return false
}

// Returns true if the dependency module belongs to any of the apexes.
var depIsApexModule exceptionHandleFunc = func(mctx ModuleContext, _, dep Module) bool {
	depContainersInfo, _ := getContainerModuleInfo(mctx, dep)
	return InList(ApexContainer, depContainersInfo.belongingContainers)
}

// Returns true if the module and the dependent module belongs to common apexes.
var belongsToCommonApexes exceptionHandleFunc = func(mctx ModuleContext, m, dep Module) bool {
	mContainersInfo, _ := getContainerModuleInfo(mctx, m)
	depContainersInfo, _ := getContainerModuleInfo(mctx, dep)

	return HasIntersection(mContainersInfo.ApexNames(), depContainersInfo.ApexNames())
}

// Returns true when all apexes that the module belongs to are non updatable.
// For an apex module to be allowed to depend on a non-apex partition module,
// all apexes that the module belong to must be non updatable.
var belongsToNonUpdatableApex exceptionHandleFunc = func(mctx ModuleContext, m, _ Module) bool {
	mContainersInfo, _ := getContainerModuleInfo(mctx, m)

	return !mContainersInfo.UpdatableApex()
}

// Returns true if the dependency is added via dependency tags that are not used to tag dynamic
// dependency tags.
var depIsNotDynamicDepTag exceptionHandleFunc = func(ctx ModuleContext, m, dep Module) bool {
	mInstallable, _ := m.(InstallableModule)
	depTag := ctx.OtherModuleDependencyTag(dep)
	return !InList(depTag, mInstallable.DynamicDependencyTags())
}

// Returns true if the dependency is added via dependency tags that are not used to tag static
// or dynamic dependency tags. These dependencies do not affect the module in compile time or in
// runtime, thus are not significant enough to raise an error.
var depIsNotStaticOrDynamicDepTag exceptionHandleFunc = func(ctx ModuleContext, m, dep Module) bool {
	mInstallable, _ := m.(InstallableModule)
	depTag := ctx.OtherModuleDependencyTag(dep)
	return !InList(depTag, append(mInstallable.StaticDependencyTags(), mInstallable.DynamicDependencyTags()...))
}

var globallyAllowlistedDependencies = []string{
	// Modules that provide annotations used within the platform and apexes.
	"aconfig-annotations-lib",
	"framework-annotations-lib",
	"unsupportedappusage",

	// TODO(b/363016634): Remove from the allowlist when the module is converted
	// to java_sdk_library and the java_aconfig_library modules depend on the stub.
	"aconfig_storage_reader_java",

	// framework-res provides core resources essential for building apps and system UI.
	// This module is implicitly added as a dependency for java modules even when the
	// dependency specifies sdk_version.
	"framework-res",

	// jacocoagent is implicitly added as a dependency in coverage builds, and is not installed
	// on the device.
	"jacocoagent",
}

// Returns true when the dependency is globally allowlisted for inter-container dependency
var depIsGloballyAllowlisted exceptionHandleFunc = func(_ ModuleContext, _, dep Module) bool {
	return InList(dep.Name(), globallyAllowlistedDependencies)
}

// Labels of exception functions, which are used to determine special dependencies that allow
// otherwise restricted inter-container dependencies
type exceptionHandleFuncLabel int

const (
	checkStubs exceptionHandleFuncLabel = iota
	checkApexModule
	checkInCommonApexes
	checkApexIsNonUpdatable
	checkNotDynamicDepTag
	checkNotStaticOrDynamicDepTag
	checkGlobalAllowlistedDep
)

// Map of [exceptionHandleFuncLabel] to the [exceptionHandleFunc]
var exceptionHandleFunctionsTable = map[exceptionHandleFuncLabel]exceptionHandleFunc{
	checkStubs:                    depIsStubsModule,
	checkApexModule:               depIsApexModule,
	checkInCommonApexes:           belongsToCommonApexes,
	checkApexIsNonUpdatable:       belongsToNonUpdatableApex,
	checkNotDynamicDepTag:         depIsNotDynamicDepTag,
	checkNotStaticOrDynamicDepTag: depIsNotStaticOrDynamicDepTag,
	checkGlobalAllowlistedDep:     depIsGloballyAllowlisted,
}

// ----------------------------------------------------------------------------
// Start of the definitions of container determination functions.
//
// Similar to the above section, below defines the functions used to determine
// the container of each modules.
// ----------------------------------------------------------------------------

type containerBoundaryFunc func(mctx ModuleContext) bool

var vendorContainerBoundaryFunc containerBoundaryFunc = func(mctx ModuleContext) bool {
	m, ok := mctx.Module().(ImageInterface)
	return mctx.Module().InstallInVendor() || (ok && m.VendorVariantNeeded(mctx))
}

var systemContainerBoundaryFunc containerBoundaryFunc = func(mctx ModuleContext) bool {
	module := mctx.Module()

	return !module.InstallInTestcases() &&
		!module.InstallInData() &&
		!module.InstallInRamdisk() &&
		!module.InstallInVendorRamdisk() &&
		!module.InstallInDebugRamdisk() &&
		!module.InstallInRecovery() &&
		!module.InstallInVendor() &&
		!module.InstallInOdm() &&
		!module.InstallInProduct() &&
		determineModuleKind(module.base(), mctx.blueprintBaseModuleContext()) == platformModule
}

var productContainerBoundaryFunc containerBoundaryFunc = func(mctx ModuleContext) bool {
	m, ok := mctx.Module().(ImageInterface)
	return mctx.Module().InstallInProduct() || (ok && m.ProductVariantNeeded(mctx))
}

var apexContainerBoundaryFunc containerBoundaryFunc = func(mctx ModuleContext) bool {
	_, ok := ModuleProvider(mctx, AllApexInfoProvider)
	return ok
}

var ctsContainerBoundaryFunc containerBoundaryFunc = func(mctx ModuleContext) bool {
	props := mctx.Module().GetProperties()
	for _, prop := range props {
		val := reflect.ValueOf(prop).Elem()
		if val.Kind() == reflect.Struct {
			testSuites := val.FieldByName("Test_suites")
			if testSuites.IsValid() && testSuites.Kind() == reflect.Slice && slices.Contains(testSuites.Interface().([]string), "cts") {
				return true
			}
		}
	}
	return false
}

type unstableInfo struct {
	// Determines if the module contains the private APIs of the platform.
	ContainsPlatformPrivateApis bool
}

var unstableInfoProvider = blueprint.NewProvider[unstableInfo]()

func determineUnstableModule(mctx ModuleContext) bool {
	module := mctx.Module()
	unstableModule := module.Name() == "framework-minus-apex"
	if installable, ok := module.(InstallableModule); ok {
		for _, staticDepTag := range installable.StaticDependencyTags() {
			mctx.VisitDirectDepsWithTag(staticDepTag, func(dep Module) {
				if unstableInfo, ok := OtherModuleProvider(mctx, dep, unstableInfoProvider); ok {
					unstableModule = unstableModule || unstableInfo.ContainsPlatformPrivateApis
				}
			})
		}
	}
	return unstableModule
}

var unstableContainerBoundaryFunc containerBoundaryFunc = func(mctx ModuleContext) bool {
	return determineUnstableModule(mctx)
}

// Map of [*container] to the [containerBoundaryFunc]
var containerBoundaryFunctionsTable = map[*container]containerBoundaryFunc{
	VendorContainer:   vendorContainerBoundaryFunc,
	SystemContainer:   systemContainerBoundaryFunc,
	ProductContainer:  productContainerBoundaryFunc,
	ApexContainer:     apexContainerBoundaryFunc,
	CtsContainer:      ctsContainerBoundaryFunc,
	UnstableContainer: unstableContainerBoundaryFunc,
}

// ----------------------------------------------------------------------------
// End of the definitions of container determination functions.
// ----------------------------------------------------------------------------

type InstallableModule interface {
	StaticDependencyTags() []blueprint.DependencyTag
	DynamicDependencyTags() []blueprint.DependencyTag
}

type restriction struct {
	// container of the dependency
	dependency *container

	// Error message to be emitted to the user when the dependency meets this restriction
	errorMessage string

	// List of labels of allowed exception functions that allows bypassing this restriction.
	// If any of the functions mapped to each labels returns true, this dependency would be
	// considered allowed and an error will not be thrown.
	allowedExceptions []exceptionHandleFuncLabel
}
type container struct {
	// The name of the container i.e. partition, api domain
	name string

	// Map of dependency restricted containers.
	restricted []restriction
}

var (
	VendorContainer = &container{
		name:       VendorVariation,
		restricted: nil,
	}

	SystemContainer = &container{
		name: "system",
		restricted: []restriction{
			{
				dependency: VendorContainer,
				errorMessage: "Module belonging to the system partition other than HALs is " +
					"not allowed to depend on the vendor partition module, in order to support " +
					"independent development/update cycles and to support the Generic System " +
					"Image. Try depending on HALs, VNDK or AIDL instead.",
				allowedExceptions: []exceptionHandleFuncLabel{
					checkStubs,
					checkNotDynamicDepTag,
					checkGlobalAllowlistedDep,
				},
			},
		},
	}

	ProductContainer = &container{
		name: ProductVariation,
		restricted: []restriction{
			{
				dependency: VendorContainer,
				errorMessage: "Module belonging to the product partition is not allowed to " +
					"depend on the vendor partition module, as this may lead to security " +
					"vulnerabilities. Try depending on the HALs or utilize AIDL instead.",
				allowedExceptions: []exceptionHandleFuncLabel{
					checkStubs,
					checkNotDynamicDepTag,
					checkGlobalAllowlistedDep,
				},
			},
		},
	}

	ApexContainer = initializeApexContainer()

	CtsContainer = &container{
		name: "cts",
		restricted: []restriction{
			{
				dependency: UnstableContainer,
				errorMessage: "CTS module should not depend on the modules that contain the " +
					"platform implementation details, including \"framework\". Depending on these " +
					"modules may lead to disclosure of implementation details and regression " +
					"due to API changes across platform versions. Try depending on the stubs instead " +
					"and ensure that the module sets an appropriate 'sdk_version'.",
				allowedExceptions: []exceptionHandleFuncLabel{
					checkStubs,
					checkNotStaticOrDynamicDepTag,
					checkGlobalAllowlistedDep,
				},
			},
		},
	}

	// Container signifying that the module contains unstable platform private APIs
	UnstableContainer = &container{
		name:       "unstable",
		restricted: nil,
	}

	allContainers = []*container{
		VendorContainer,
		SystemContainer,
		ProductContainer,
		ApexContainer,
		CtsContainer,
		UnstableContainer,
	}
)

func initializeApexContainer() *container {
	apexContainer := &container{
		name: "apex",
		restricted: []restriction{
			{
				dependency: SystemContainer,
				errorMessage: "Module belonging to Apex(es) is not allowed to depend on the " +
					"modules belonging to the system partition. Either statically depend on the " +
					"module or convert the depending module to java_sdk_library and depend on " +
					"the stubs.",
				allowedExceptions: []exceptionHandleFuncLabel{
					checkStubs,
					checkApexModule,
					checkInCommonApexes,
					checkApexIsNonUpdatable,
					checkNotStaticOrDynamicDepTag,
					checkGlobalAllowlistedDep,
				},
			},
		},
	}

	apexContainer.restricted = append(apexContainer.restricted, restriction{
		dependency: apexContainer,
		errorMessage: "Module belonging to Apex(es) is not allowed to depend on the " +
			"modules belonging to other Apex(es). Either include the depending " +
			"module in the Apex or convert the depending module to java_sdk_library " +
			"and depend on its stubs.",
		allowedExceptions: []exceptionHandleFuncLabel{
			checkStubs,
			checkInCommonApexes,
			checkNotStaticOrDynamicDepTag,
			checkGlobalAllowlistedDep,
		},
	})

	return apexContainer
}

type ContainersInfo struct {
	belongingContainers []*container

	belongingApexes []ApexInfo
}

func (c *ContainersInfo) BelongingContainers() []*container {
	return c.belongingContainers
}

func (c *ContainersInfo) ApexNames() (ret []string) {
	for _, apex := range c.belongingApexes {
		ret = append(ret, apex.InApexModules...)
	}
	slices.Sort(ret)
	return ret
}

// Returns true if any of the apex the module belongs to is updatable.
func (c *ContainersInfo) UpdatableApex() bool {
	for _, apex := range c.belongingApexes {
		if apex.Updatable {
			return true
		}
	}
	return false
}

var ContainersInfoProvider = blueprint.NewProvider[ContainersInfo]()

func satisfyAllowedExceptions(ctx ModuleContext, allowedExceptionLabels []exceptionHandleFuncLabel, m, dep Module) bool {
	for _, label := range allowedExceptionLabels {
		if exceptionHandleFunctionsTable[label](ctx, m, dep) {
			return true
		}
	}
	return false
}

func (c *ContainersInfo) GetViolations(mctx ModuleContext, m, dep Module, depInfo ContainersInfo) []string {
	var violations []string

	// Any containers that the module belongs to but the dependency does not belong to must be examined.
	_, containersUniqueToModule, _ := ListSetDifference(c.belongingContainers, depInfo.belongingContainers)

	// Apex container should be examined even if both the module and the dependency belong to
	// the apex container to check that the two modules belong to the same apex.
	if InList(ApexContainer, c.belongingContainers) && !InList(ApexContainer, containersUniqueToModule) {
		containersUniqueToModule = append(containersUniqueToModule, ApexContainer)
	}

	for _, containerUniqueToModule := range containersUniqueToModule {
		for _, restriction := range containerUniqueToModule.restricted {
			if InList(restriction.dependency, depInfo.belongingContainers) {
				if !satisfyAllowedExceptions(mctx, restriction.allowedExceptions, m, dep) {
					violations = append(violations, restriction.errorMessage)
				}
			}
		}
	}

	return violations
}

func generateContainerInfo(ctx ModuleContext) ContainersInfo {
	var containers []*container

	for _, cnt := range allContainers {
		if containerBoundaryFunctionsTable[cnt](ctx) {
			containers = append(containers, cnt)
		}
	}

	var belongingApexes []ApexInfo
	if apexInfo, ok := ModuleProvider(ctx, AllApexInfoProvider); ok {
		belongingApexes = apexInfo.ApexInfos
	}

	return ContainersInfo{
		belongingContainers: containers,
		belongingApexes:     belongingApexes,
	}
}

func getContainerModuleInfo(ctx ModuleContext, module Module) (ContainersInfo, bool) {
	if ctx.Module() == module {
		return ctx.getContainersInfo(), true
	}

	return OtherModuleProvider(ctx, module, ContainersInfoProvider)
}

func setContainerInfo(ctx ModuleContext) {
	// Required to determine the unstable container. This provider is set here instead of the
	// unstableContainerBoundaryFunc in order to prevent setting the provider multiple times.
	SetProvider(ctx, unstableInfoProvider, unstableInfo{
		ContainsPlatformPrivateApis: determineUnstableModule(ctx),
	})

	if _, ok := ctx.Module().(InstallableModule); ok {
		containersInfo := generateContainerInfo(ctx)
		ctx.setContainersInfo(containersInfo)
		SetProvider(ctx, ContainersInfoProvider, containersInfo)
	}
}

func checkContainerViolations(ctx ModuleContext) {
	if _, ok := ctx.Module().(InstallableModule); ok {
		containersInfo, _ := getContainerModuleInfo(ctx, ctx.Module())
		ctx.VisitDirectDepsIgnoreBlueprint(func(dep Module) {
			if !dep.Enabled(ctx) {
				return
			}

			// Pre-existing violating dependencies are tracked in containerDependencyViolationAllowlist.
			// If this dependency is allowlisted, do not check for violation.
			// If not, check if this dependency matches any restricted dependency and
			// satisfies any exception functions, which allows bypassing the
			// restriction. If all of the exceptions are not satisfied, throw an error.
			if depContainersInfo, ok := getContainerModuleInfo(ctx, dep); ok {
				if allowedViolations, ok := ContainerDependencyViolationAllowlist[ctx.ModuleName()]; ok && InList(dep.Name(), allowedViolations) {
					return
				} else {
					violations := containersInfo.GetViolations(ctx, ctx.Module(), dep, depContainersInfo)
					if len(violations) > 0 {
						errorMessage := fmt.Sprintf("%s cannot depend on %s. ", ctx.ModuleName(), dep.Name())
						errorMessage += strings.Join(violations, " ")
						ctx.ModuleErrorf(errorMessage)
					}
				}
			}
		})
	}
}
