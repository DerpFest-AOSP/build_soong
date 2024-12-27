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

package java

import (
	"android/soong/android"
	"fmt"
	"testing"
)

var checkContainerMatch = func(t *testing.T, name string, container string, expected bool, actual bool) {
	errorMessage := fmt.Sprintf("module %s container %s value differ", name, container)
	android.AssertBoolEquals(t, errorMessage, expected, actual)
}

func TestJavaContainersModuleProperties(t *testing.T) {
	result := android.GroupFixturePreparers(
		prepareForJavaTest,
	).RunTestWithBp(t, `
		java_library {
			name: "foo",
			srcs: ["A.java"],
		}
		java_library {
			name: "foo_vendor",
			srcs: ["A.java"],
			vendor: true,
			sdk_version: "current",
		}
		java_library {
			name: "foo_soc_specific",
			srcs: ["A.java"],
			soc_specific: true,
			sdk_version: "current",
		}
		java_library {
			name: "foo_product_specific",
			srcs: ["A.java"],
			product_specific: true,
			sdk_version: "current",
		}
		java_test {
			name: "foo_cts_test",
			srcs: ["A.java"],
			test_suites: [
				"cts",
			],
		}
		java_test {
			name: "foo_non_cts_test",
			srcs: ["A.java"],
			test_suites: [
				"general-tests",
			],
		}
		java_library {
			name: "bar",
			static_libs: [
				"framework-minus-apex",
			],
		}
		java_library {
			name: "baz",
			static_libs: [
				"bar",
			],
		}
	`)

	testcases := []struct {
		moduleName         string
		isSystemContainer  bool
		isVendorContainer  bool
		isProductContainer bool
		isCts              bool
		isUnstable         bool
	}{
		{
			moduleName:         "foo",
			isSystemContainer:  true,
			isVendorContainer:  false,
			isProductContainer: false,
			isCts:              false,
			isUnstable:         false,
		},
		{
			moduleName:         "foo_vendor",
			isSystemContainer:  false,
			isVendorContainer:  true,
			isProductContainer: false,
			isCts:              false,
			isUnstable:         false,
		},
		{
			moduleName:         "foo_soc_specific",
			isSystemContainer:  false,
			isVendorContainer:  true,
			isProductContainer: false,
			isCts:              false,
			isUnstable:         false,
		},
		{
			moduleName:         "foo_product_specific",
			isSystemContainer:  false,
			isVendorContainer:  false,
			isProductContainer: true,
			isCts:              false,
			isUnstable:         false,
		},
		{
			moduleName:         "foo_cts_test",
			isSystemContainer:  false,
			isVendorContainer:  false,
			isProductContainer: false,
			isCts:              true,
			isUnstable:         false,
		},
		{
			moduleName:         "foo_non_cts_test",
			isSystemContainer:  false,
			isVendorContainer:  false,
			isProductContainer: false,
			isCts:              false,
			isUnstable:         false,
		},
		{
			moduleName:         "bar",
			isSystemContainer:  true,
			isVendorContainer:  false,
			isProductContainer: false,
			isCts:              false,
			isUnstable:         true,
		},
		{
			moduleName:         "baz",
			isSystemContainer:  true,
			isVendorContainer:  false,
			isProductContainer: false,
			isCts:              false,
			isUnstable:         true,
		},
	}

	for _, c := range testcases {
		m := result.ModuleForTests(c.moduleName, "android_common")
		containers, _ := android.OtherModuleProvider(result.TestContext.OtherModuleProviderAdaptor(), m.Module(), android.ContainersInfoProvider)
		belongingContainers := containers.BelongingContainers()
		checkContainerMatch(t, c.moduleName, "system", c.isSystemContainer, android.InList(android.SystemContainer, belongingContainers))
		checkContainerMatch(t, c.moduleName, "vendor", c.isVendorContainer, android.InList(android.VendorContainer, belongingContainers))
		checkContainerMatch(t, c.moduleName, "product", c.isProductContainer, android.InList(android.ProductContainer, belongingContainers))
		checkContainerMatch(t, c.moduleName, "cts", c.isCts, android.InList(android.CtsContainer, belongingContainers))
		checkContainerMatch(t, c.moduleName, "unstable", c.isUnstable, android.InList(android.UnstableContainer, belongingContainers))
	}
}
