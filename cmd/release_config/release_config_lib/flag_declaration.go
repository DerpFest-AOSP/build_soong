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

package release_config_lib

import (
	rc_proto "android/soong/cmd/release_config/release_config_proto"
)

var (
	// Allowlist: these flags may have duplicate (identical) declarations
	// without generating an error.  This will be removed once all such
	// declarations have been fixed.
	DuplicateDeclarationAllowlist = map[string]bool{}
)

func FlagDeclarationFactory(protoPath string) (fd *rc_proto.FlagDeclaration) {
	fd = &rc_proto.FlagDeclaration{}
	if protoPath != "" {
		LoadMessage(protoPath, fd)
	}
	// If the input didn't specify a value, create one (== UnspecifiedValue).
	if fd.Value == nil {
		fd.Value = &rc_proto.Value{Val: &rc_proto.Value_UnspecifiedValue{false}}
	}
	return fd
}
