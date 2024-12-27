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

package libbpf_prog

import (
	"os"
	"testing"

	"android/soong/android"
	"android/soong/cc"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

var prepareForLibbpfProgTest = android.GroupFixturePreparers(
	cc.PrepareForTestWithCcDefaultModules,
	android.FixtureMergeMockFs(
		map[string][]byte{
			"bpf.c":              nil,
			"bpf_invalid_name.c": nil,
			"BpfTest.cpp":        nil,
		},
	),
	PrepareForTestWithLibbpfProg,
)

func TestLibbpfProgDataDependency(t *testing.T) {
	bp := `
		libbpf_prog {
			name: "bpf.o",
			srcs: ["bpf.c"],
		}

		cc_test {
			name: "vts_test_binary_bpf_module",
			srcs: ["BpfTest.cpp"],
			data: [":bpf.o"],
			gtest: false,
		}
	`

	prepareForLibbpfProgTest.RunTestWithBp(t, bp)
}

func TestLibbpfProgSourceName(t *testing.T) {
	bp := `
		libbpf_prog {
			name: "bpf_invalid_name.o",
			srcs: ["bpf_invalid_name.c"],
		}
	`
	prepareForLibbpfProgTest.ExtendWithErrorHandler(android.FixtureExpectsOneErrorPattern(
		`invalid character '_' in source name`)).
		RunTestWithBp(t, bp)
}
