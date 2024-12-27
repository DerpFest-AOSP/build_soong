// Copyright 2019 Google Inc. All rights reserved.
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

// ImageInterface is implemented by modules that need to be split by the imageTransitionMutator.
type ImageInterface interface {
	// ImageMutatorBegin is called before any other method in the ImageInterface.
	ImageMutatorBegin(ctx BaseModuleContext)

	// VendorVariantNeeded should return true if the module needs a vendor variant (installed on the vendor image).
	VendorVariantNeeded(ctx BaseModuleContext) bool

	// ProductVariantNeeded should return true if the module needs a product variant (installed on the product image).
	ProductVariantNeeded(ctx BaseModuleContext) bool

	// CoreVariantNeeded should return true if the module needs a core variant (installed on the system image).
	CoreVariantNeeded(ctx BaseModuleContext) bool

	// RamdiskVariantNeeded should return true if the module needs a ramdisk variant (installed on the
	// ramdisk partition).
	RamdiskVariantNeeded(ctx BaseModuleContext) bool

	// VendorRamdiskVariantNeeded should return true if the module needs a vendor ramdisk variant (installed on the
	// vendor ramdisk partition).
	VendorRamdiskVariantNeeded(ctx BaseModuleContext) bool

	// DebugRamdiskVariantNeeded should return true if the module needs a debug ramdisk variant (installed on the
	// debug ramdisk partition: $(PRODUCT_OUT)/debug_ramdisk).
	DebugRamdiskVariantNeeded(ctx BaseModuleContext) bool

	// RecoveryVariantNeeded should return true if the module needs a recovery variant (installed on the
	// recovery partition).
	RecoveryVariantNeeded(ctx BaseModuleContext) bool

	// ExtraImageVariations should return a list of the additional variations needed for the module.  After the
	// variants are created the SetImageVariation method will be called on each newly created variant with the
	// its variation.
	ExtraImageVariations(ctx BaseModuleContext) []string

	// SetImageVariation is called for each newly created image variant. The receiver is the original
	// module, "variation" is the name of the newly created variant. "variation" is set on the receiver.
	SetImageVariation(ctx BaseModuleContext, variation string)
}

const (
	// VendorVariation is the variant name used for /vendor code that does not
	// compile against the VNDK.
	VendorVariation string = "vendor"

	// ProductVariation is the variant name used for /product code that does not
	// compile against the VNDK.
	ProductVariation string = "product"

	// CoreVariation is the variant used for framework-private libraries, or
	// SDK libraries. (which framework-private libraries can use), which
	// will be installed to the system image.
	CoreVariation string = ""

	// RecoveryVariation means a module to be installed to recovery image.
	RecoveryVariation string = "recovery"

	// RamdiskVariation means a module to be installed to ramdisk image.
	RamdiskVariation string = "ramdisk"

	// VendorRamdiskVariation means a module to be installed to vendor ramdisk image.
	VendorRamdiskVariation string = "vendor_ramdisk"

	// DebugRamdiskVariation means a module to be installed to debug ramdisk image.
	DebugRamdiskVariation string = "debug_ramdisk"
)

// imageTransitionMutator creates variants for modules that implement the ImageInterface that
// allow them to build differently for each partition (recovery, core, vendor, etc.).
type imageTransitionMutator struct{}

func (imageTransitionMutator) Split(ctx BaseModuleContext) []string {
	var variations []string

	if m, ok := ctx.Module().(ImageInterface); ctx.Os() == Android && ok {
		m.ImageMutatorBegin(ctx)

		if m.CoreVariantNeeded(ctx) {
			variations = append(variations, CoreVariation)
		}
		if m.RamdiskVariantNeeded(ctx) {
			variations = append(variations, RamdiskVariation)
		}
		if m.VendorRamdiskVariantNeeded(ctx) {
			variations = append(variations, VendorRamdiskVariation)
		}
		if m.DebugRamdiskVariantNeeded(ctx) {
			variations = append(variations, DebugRamdiskVariation)
		}
		if m.RecoveryVariantNeeded(ctx) {
			variations = append(variations, RecoveryVariation)
		}
		if m.VendorVariantNeeded(ctx) {
			variations = append(variations, VendorVariation)
		}
		if m.ProductVariantNeeded(ctx) {
			variations = append(variations, ProductVariation)
		}

		extraVariations := m.ExtraImageVariations(ctx)
		variations = append(variations, extraVariations...)
	}

	if len(variations) == 0 {
		variations = append(variations, "")
	}

	return variations
}

func (imageTransitionMutator) OutgoingTransition(ctx OutgoingTransitionContext, sourceVariation string) string {
	return sourceVariation
}

func (imageTransitionMutator) IncomingTransition(ctx IncomingTransitionContext, incomingVariation string) string {
	if _, ok := ctx.Module().(ImageInterface); ctx.Os() != Android || !ok {
		return CoreVariation
	}
	return incomingVariation
}

func (imageTransitionMutator) Mutate(ctx BottomUpMutatorContext, variation string) {
	ctx.Module().base().setImageVariation(variation)
	if m, ok := ctx.Module().(ImageInterface); ok {
		m.SetImageVariation(ctx, variation)
	}
}
