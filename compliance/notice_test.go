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

package compliance

import (
	"testing"

	"android/soong/android"
)

var prepareForNoticeXmlTest = android.GroupFixturePreparers(
	android.PrepareForTestWithArchMutator,
	PrepareForTestWithNoticeXml,
)

func TestPrebuiltEtcOutputFile(t *testing.T) {
	result := prepareForNoticeXmlTest.RunTestWithBp(t, `
		notice_xml {
			name: "notice_xml_system",
			partition_name: "system",
		}
	`)

	m := result.Module("notice_xml_system", "android_arm64_armv8-a").(*NoticeXmlModule)
	android.AssertStringEquals(t, "output file", "NOTICE.xml.gz", m.outputFile.Base())
}