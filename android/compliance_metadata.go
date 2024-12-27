// Copyright 2024 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package android

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/google/blueprint"
)

var (
	// Constants of property names used in compliance metadata of modules
	ComplianceMetadataProp = struct {
		NAME                   string
		PACKAGE                string
		MODULE_TYPE            string
		OS                     string
		ARCH                   string
		IS_PRIMARY_ARCH        string
		VARIANT                string
		IS_STATIC_LIB          string
		INSTALLED_FILES        string
		BUILT_FILES            string
		STATIC_DEPS            string
		STATIC_DEP_FILES       string
		WHOLE_STATIC_DEPS      string
		WHOLE_STATIC_DEP_FILES string
		LICENSES               string

		// module_type=package
		PKG_DEFAULT_APPLICABLE_LICENSES string

		// module_type=license
		LIC_LICENSE_KINDS string
		LIC_LICENSE_TEXT  string
		LIC_PACKAGE_NAME  string

		// module_type=license_kind
		LK_CONDITIONS string
		LK_URL        string
	}{
		"name",
		"package",
		"module_type",
		"os",
		"arch",
		"is_primary_arch",
		"variant",
		"is_static_lib",
		"installed_files",
		"built_files",
		"static_deps",
		"static_dep_files",
		"whole_static_deps",
		"whole_static_dep_files",
		"licenses",

		"pkg_default_applicable_licenses",

		"lic_license_kinds",
		"lic_license_text",
		"lic_package_name",

		"lk_conditions",
		"lk_url",
	}

	// A constant list of all property names in compliance metadata
	// Order of properties here is the order of columns in the exported CSV file.
	COMPLIANCE_METADATA_PROPS = []string{
		ComplianceMetadataProp.NAME,
		ComplianceMetadataProp.PACKAGE,
		ComplianceMetadataProp.MODULE_TYPE,
		ComplianceMetadataProp.OS,
		ComplianceMetadataProp.ARCH,
		ComplianceMetadataProp.VARIANT,
		ComplianceMetadataProp.IS_STATIC_LIB,
		ComplianceMetadataProp.IS_PRIMARY_ARCH,
		// Space separated installed files
		ComplianceMetadataProp.INSTALLED_FILES,
		// Space separated built files
		ComplianceMetadataProp.BUILT_FILES,
		// Space separated module names of static dependencies
		ComplianceMetadataProp.STATIC_DEPS,
		// Space separated file paths of static dependencies
		ComplianceMetadataProp.STATIC_DEP_FILES,
		// Space separated module names of whole static dependencies
		ComplianceMetadataProp.WHOLE_STATIC_DEPS,
		// Space separated file paths of whole static dependencies
		ComplianceMetadataProp.WHOLE_STATIC_DEP_FILES,
		ComplianceMetadataProp.LICENSES,
		// module_type=package
		ComplianceMetadataProp.PKG_DEFAULT_APPLICABLE_LICENSES,
		// module_type=license
		ComplianceMetadataProp.LIC_LICENSE_KINDS,
		ComplianceMetadataProp.LIC_LICENSE_TEXT, // resolve to file paths
		ComplianceMetadataProp.LIC_PACKAGE_NAME,
		// module_type=license_kind
		ComplianceMetadataProp.LK_CONDITIONS,
		ComplianceMetadataProp.LK_URL,
	}
)

// ComplianceMetadataInfo provides all metadata of a module, e.g. name, module type, package, license,
// dependencies, built/installed files, etc. It is a wrapper on a map[string]string with some utility
// methods to get/set properties' values.
type ComplianceMetadataInfo struct {
	properties map[string]string
}

func NewComplianceMetadataInfo() *ComplianceMetadataInfo {
	return &ComplianceMetadataInfo{
		properties: map[string]string{},
	}
}

func (c *ComplianceMetadataInfo) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(c.properties)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (c *ComplianceMetadataInfo) GobDecode(data []byte) error {
	r := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&c.properties)
	if err != nil {
		return err
	}

	return nil
}

func (c *ComplianceMetadataInfo) SetStringValue(propertyName string, value string) {
	if !slices.Contains(COMPLIANCE_METADATA_PROPS, propertyName) {
		panic(fmt.Errorf("Unknown metadata property: %s.", propertyName))
	}
	c.properties[propertyName] = value
}

func (c *ComplianceMetadataInfo) SetListValue(propertyName string, value []string) {
	c.SetStringValue(propertyName, strings.TrimSpace(strings.Join(value, " ")))
}

func (c *ComplianceMetadataInfo) getStringValue(propertyName string) string {
	if !slices.Contains(COMPLIANCE_METADATA_PROPS, propertyName) {
		panic(fmt.Errorf("Unknown metadata property: %s.", propertyName))
	}
	return c.properties[propertyName]
}

func (c *ComplianceMetadataInfo) getAllValues() map[string]string {
	return c.properties
}

var (
	ComplianceMetadataProvider = blueprint.NewProvider[*ComplianceMetadataInfo]()
)

// buildComplianceMetadataProvider starts with the ModuleContext.ComplianceMetadataInfo() and fills in more common metadata
// for different module types without accessing their private fields but through android.Module interface
// and public/private fields of package android. The final metadata is stored to a module's ComplianceMetadataProvider.
func buildComplianceMetadataProvider(ctx *moduleContext, m *ModuleBase) {
	complianceMetadataInfo := ctx.ComplianceMetadataInfo()
	complianceMetadataInfo.SetStringValue(ComplianceMetadataProp.NAME, m.Name())
	complianceMetadataInfo.SetStringValue(ComplianceMetadataProp.PACKAGE, ctx.ModuleDir())
	complianceMetadataInfo.SetStringValue(ComplianceMetadataProp.MODULE_TYPE, ctx.ModuleType())

	switch ctx.ModuleType() {
	case "license":
		licenseModule := m.module.(*licenseModule)
		complianceMetadataInfo.SetListValue(ComplianceMetadataProp.LIC_LICENSE_KINDS, licenseModule.properties.License_kinds)
		complianceMetadataInfo.SetListValue(ComplianceMetadataProp.LIC_LICENSE_TEXT, PathsForModuleSrc(ctx, licenseModule.properties.License_text).Strings())
		complianceMetadataInfo.SetStringValue(ComplianceMetadataProp.LIC_PACKAGE_NAME, String(licenseModule.properties.Package_name))
	case "license_kind":
		licenseKindModule := m.module.(*licenseKindModule)
		complianceMetadataInfo.SetListValue(ComplianceMetadataProp.LK_CONDITIONS, licenseKindModule.properties.Conditions)
		complianceMetadataInfo.SetStringValue(ComplianceMetadataProp.LK_URL, licenseKindModule.properties.Url)
	default:
		complianceMetadataInfo.SetStringValue(ComplianceMetadataProp.OS, ctx.Os().String())
		complianceMetadataInfo.SetStringValue(ComplianceMetadataProp.ARCH, ctx.Arch().String())
		complianceMetadataInfo.SetStringValue(ComplianceMetadataProp.IS_PRIMARY_ARCH, strconv.FormatBool(ctx.PrimaryArch()))
		complianceMetadataInfo.SetStringValue(ComplianceMetadataProp.VARIANT, ctx.ModuleSubDir())
		if m.primaryLicensesProperty != nil && m.primaryLicensesProperty.getName() == "licenses" {
			complianceMetadataInfo.SetListValue(ComplianceMetadataProp.LICENSES, m.primaryLicensesProperty.getStrings())
		}

		var installed InstallPaths
		installed = append(installed, ctx.installFiles...)
		installed = append(installed, ctx.katiInstalls.InstallPaths()...)
		installed = append(installed, ctx.katiSymlinks.InstallPaths()...)
		installed = append(installed, ctx.katiInitRcInstalls.InstallPaths()...)
		installed = append(installed, ctx.katiVintfInstalls.InstallPaths()...)
		complianceMetadataInfo.SetListValue(ComplianceMetadataProp.INSTALLED_FILES, FirstUniqueStrings(installed.Strings()))
	}
	ctx.setProvider(ComplianceMetadataProvider, complianceMetadataInfo)
}

func init() {
	RegisterComplianceMetadataSingleton(InitRegistrationContext)
}

func RegisterComplianceMetadataSingleton(ctx RegistrationContext) {
	ctx.RegisterParallelSingletonType("compliance_metadata_singleton", complianceMetadataSingletonFactory)
}

var (
	// sqlite3 command line tool
	sqlite3 = pctx.HostBinToolVariable("sqlite3", "sqlite3")

	// Command to import .csv files to sqlite3 database
	importCsv = pctx.AndroidStaticRule("importCsv",
		blueprint.RuleParams{
			Command: `rm -rf $out && ` +
				`${sqlite3} $out ".import --csv $in modules" && ` +
				`${sqlite3} $out ".import --csv ${make_metadata} make_metadata" && ` +
				`${sqlite3} $out ".import --csv ${make_modules} make_modules"`,
			CommandDeps: []string{"${sqlite3}"},
		}, "make_metadata", "make_modules")
)

func complianceMetadataSingletonFactory() Singleton {
	return &complianceMetadataSingleton{}
}

type complianceMetadataSingleton struct {
}

func writerToCsv(csvWriter *csv.Writer, row []string) {
	err := csvWriter.Write(row)
	if err != nil {
		panic(err)
	}
}

// Collect compliance metadata from all Soong modules, write to a CSV file and
// import compliance metadata from Make and Soong to a sqlite3 database.
func (c *complianceMetadataSingleton) GenerateBuildActions(ctx SingletonContext) {
	if !ctx.Config().HasDeviceProduct() {
		return
	}
	var buffer bytes.Buffer
	csvWriter := csv.NewWriter(&buffer)

	// Collect compliance metadata of modules in Soong and write to out/soong/compliance-metadata/<product>/soong-modules.csv file.
	columnNames := []string{"id"}
	columnNames = append(columnNames, COMPLIANCE_METADATA_PROPS...)
	writerToCsv(csvWriter, columnNames)

	rowId := -1
	ctx.VisitAllModules(func(module Module) {
		if !module.Enabled(ctx) {
			return
		}
		moduleType := ctx.ModuleType(module)
		if moduleType == "package" {
			metadataMap := map[string]string{
				ComplianceMetadataProp.NAME:                            ctx.ModuleName(module),
				ComplianceMetadataProp.MODULE_TYPE:                     ctx.ModuleType(module),
				ComplianceMetadataProp.PKG_DEFAULT_APPLICABLE_LICENSES: strings.Join(module.base().primaryLicensesProperty.getStrings(), " "),
			}
			rowId = rowId + 1
			metadata := []string{strconv.Itoa(rowId)}
			for _, propertyName := range COMPLIANCE_METADATA_PROPS {
				metadata = append(metadata, metadataMap[propertyName])
			}
			writerToCsv(csvWriter, metadata)
			return
		}
		if provider, ok := ctx.otherModuleProvider(module, ComplianceMetadataProvider); ok {
			metadataInfo := provider.(*ComplianceMetadataInfo)
			rowId = rowId + 1
			metadata := []string{strconv.Itoa(rowId)}
			for _, propertyName := range COMPLIANCE_METADATA_PROPS {
				metadata = append(metadata, metadataInfo.getStringValue(propertyName))
			}
			writerToCsv(csvWriter, metadata)
			return
		}
	})
	csvWriter.Flush()

	deviceProduct := ctx.Config().DeviceProduct()
	modulesCsv := PathForOutput(ctx, "compliance-metadata", deviceProduct, "soong-modules.csv")
	WriteFileRuleVerbatim(ctx, modulesCsv, buffer.String())

	// Metadata generated in Make
	makeMetadataCsv := PathForOutput(ctx, "compliance-metadata", deviceProduct, "make-metadata.csv")
	makeModulesCsv := PathForOutput(ctx, "compliance-metadata", deviceProduct, "make-modules.csv")

	// Import metadata from Make and Soong to sqlite3 database
	complianceMetadataDb := PathForOutput(ctx, "compliance-metadata", deviceProduct, "compliance-metadata.db")
	ctx.Build(pctx, BuildParams{
		Rule:  importCsv,
		Input: modulesCsv,
		Implicits: []Path{
			makeMetadataCsv,
			makeModulesCsv,
		},
		Output: complianceMetadataDb,
		Args: map[string]string{
			"make_metadata": makeMetadataCsv.String(),
			"make_modules":  makeModulesCsv.String(),
		},
	})

	// Phony rule "compliance-metadata.db". "m compliance-metadata.db" to create the compliance metadata database.
	ctx.Build(pctx, BuildParams{
		Rule:   blueprint.Phony,
		Inputs: []Path{complianceMetadataDb},
		Output: PathForPhony(ctx, "compliance-metadata.db"),
	})

}
