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

var ContainerDependencyViolationAllowlist = map[string][]string{
	"adservices-service-core": {
		"gson", // apex [com.android.adservices, com.android.extservices] -> apex [com.android.virt]
	},

	"android.car-module.impl": {
		"modules-utils-preconditions", // apex [com.android.car.framework] -> apex [com.android.adservices, com.android.appsearch, com.android.cellbroadcast, com.android.extservices, com.android.ondevicepersonalization, com.android.tethering, com.android.uwb, com.android.wifi, test_com.android.cellbroadcast, test_com.android.wifi]
	},

	"AppInstalledOnMultipleUsers": {
		"framework", // cts -> unstable
	},

	"art-aconfig-flags-java-lib": {
		"framework-api-annotations-lib", // apex [com.android.art, com.android.art.debug, com.android.art.testing, test_imgdiag_com.android.art, test_jitzygote_com.android.art] -> system
	},

	"Bluetooth": {
		"app-compat-annotations",         // apex [com.android.btservices] -> system
		"framework-bluetooth-pre-jarjar", // apex [com.android.btservices] -> system
	},

	"bluetooth-nano-protos": {
		"libprotobuf-java-nano", // apex [com.android.btservices] -> apex [com.android.wifi, test_com.android.wifi]
	},

	"bluetooth.change-ids": {
		"app-compat-annotations", // apex [com.android.btservices] -> system
	},

	"CarServiceUpdatable": {
		"modules-utils-os",                    // apex [com.android.car.framework] -> apex [com.android.permission, test_com.android.permission]
		"modules-utils-preconditions",         // apex [com.android.car.framework] -> apex [com.android.adservices, com.android.appsearch, com.android.cellbroadcast, com.android.extservices, com.android.ondevicepersonalization, com.android.tethering, com.android.uwb, com.android.wifi, test_com.android.cellbroadcast, test_com.android.wifi]
		"modules-utils-shell-command-handler", // apex [com.android.car.framework] -> apex [com.android.adservices, com.android.art, com.android.art.debug, com.android.art.testing, com.android.btservices, com.android.configinfrastructure, com.android.mediaprovider, com.android.nfcservices, com.android.permission, com.android.scheduling, com.android.tethering, com.android.uwb, com.android.wifi, test_com.android.mediaprovider, test_com.android.permission, test_com.android.wifi, test_imgdiag_com.android.art, test_jitzygote_com.android.art]
	},

	"cellbroadcastreceiver_aconfig_flags_lib": {
		"ext",       // apex [com.android.cellbroadcast, test_com.android.cellbroadcast] -> system
		"framework", // apex [com.android.cellbroadcast, test_com.android.cellbroadcast] -> system
	},

	"connectivity-net-module-utils-bpf": {
		"net-utils-device-common-struct-base", // apex [com.android.tethering] -> system
	},

	"conscrypt-aconfig-flags-lib": {
		"aconfig-annotations-lib-sdk-none", // apex [com.android.conscrypt, test_com.android.conscrypt] -> system
	},

	"cronet_aml_base_base_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
		"jsr305", // apex [com.android.tethering] -> apex [com.android.adservices, com.android.devicelock, com.android.extservices, com.android.healthfitness, com.android.media, com.android.mediaprovider, test_com.android.media, test_com.android.mediaprovider]
	},

	"cronet_aml_build_android_build_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_components_cronet_android_base_feature_overrides_java_proto": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_components_cronet_android_cronet_api_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_components_cronet_android_cronet_impl_common_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_components_cronet_android_cronet_impl_native_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
		"jsr305", // apex [com.android.tethering] -> apex [com.android.adservices, com.android.devicelock, com.android.extservices, com.android.healthfitness, com.android.media, com.android.mediaprovider, test_com.android.media, test_com.android.mediaprovider]
	},

	"cronet_aml_components_cronet_android_cronet_jni_registration_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_components_cronet_android_cronet_shared_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_components_cronet_android_cronet_stats_log_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_components_cronet_android_cronet_urlconnection_impl_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_components_cronet_android_flags_java_proto": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_components_cronet_android_request_context_config_java_proto": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_net_android_net_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
		"jsr305", // apex [com.android.tethering] -> apex [com.android.adservices, com.android.devicelock, com.android.extservices, com.android.healthfitness, com.android.media, com.android.mediaprovider, test_com.android.media, test_com.android.mediaprovider]
	},

	"cronet_aml_net_android_net_thread_stats_uid_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_third_party_jni_zero_jni_zero_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"cronet_aml_url_url_java": {
		"framework-connectivity-pre-jarjar-without-cronet", // apex [com.android.tethering] -> system
	},

	"CtsAdservicesHostTestApp": {
		"framework", // cts -> unstable
	},

	"CtsAdServicesNotInAllowListEndToEndTests": {
		"framework", // cts -> unstable
	},

	"CtsAdServicesPermissionsAppOptOutEndToEndTests": {
		"framework", // cts -> unstable
	},

	"CtsAdServicesPermissionsNoPermEndToEndTests": {
		"framework", // cts -> unstable
	},

	"CtsAdServicesPermissionsValidEndToEndTests": {
		"framework", // cts -> unstable
	},

	"CtsAlarmManagerTestCases": {
		"framework", // cts -> unstable
	},

	"CtsAndroidAppTestCases": {
		"framework", // cts -> unstable
	},

	"CtsAppExitTestCases": {
		"framework", // cts -> unstable
	},

	"CtsAppFgsStartTestCases": {
		"framework", // cts -> unstable
	},

	"CtsAppFgsTestCases": {
		"framework", // cts -> unstable
	},

	"CtsAppFunctionTestCases": {
		"framework", // cts -> unstable
	},

	"CtsAppOpsTestCases": {
		"framework", // cts -> unstable
	},

	"CtsAppSearchTestCases": {
		"framework", // cts -> unstable
	},

	"CtsAppStartTestCases": {
		"framework", // cts -> unstable
	},

	"CtsAppTestStubsApp2": {
		"framework", // cts -> unstable
	},

	"CtsAudioHostTestApp": {
		"framework", // cts -> unstable
	},

	"CtsBackgroundActivityAppAllowCrossUidFlagDefault": {
		"framework", // cts -> unstable
	},

	"CtsBatterySavingTestCases": {
		"framework", // cts -> unstable
	},

	"CtsBluetoothTestCases": {
		"framework", // cts -> unstable
	},

	"CtsBootDisplayModeApp": {
		"framework", // cts -> unstable
	},

	"CtsBroadcastTestCases": {
		"framework", // cts -> unstable
	},

	"CtsBRSTestCases": {
		"framework", // cts -> unstable
	},

	"CtsCompanionDeviceManagerCoreTestCases": {
		"framework", // cts -> unstable
	},

	"CtsCompanionDeviceManagerMultiProcessTestCases": {
		"framework", // cts -> unstable
	},

	"CtsCompanionDeviceManagerUiAutomationTestCases": {
		"framework", // cts -> unstable
	},

	"CtsContentSuggestionsTestCases": {
		"framework", // cts -> unstable
	},

	"CtsContentTestCases": {
		"framework", // cts -> unstable
	},

	"CtsCredentialManagerBackupRestoreApp": {
		"framework", // cts -> unstable
	},

	"CtsCrossProfileEnabledApp": {
		"framework", // cts -> unstable
	},

	"CtsCrossProfileEnabledNoPermsApp": {
		"framework", // cts -> unstable
	},

	"CtsCrossProfileNotEnabledApp": {
		"framework", // cts -> unstable
	},

	"CtsCrossProfileUserEnabledApp": {
		"framework", // cts -> unstable
	},

	"CtsDeviceAndProfileOwnerApp": {
		"framework", // cts -> unstable
	},

	"CtsDeviceAndProfileOwnerApp23": {
		"framework", // cts -> unstable
	},

	"CtsDeviceAndProfileOwnerApp25": {
		"framework", // cts -> unstable
	},

	"CtsDeviceAndProfileOwnerApp30": {
		"framework", // cts -> unstable
	},

	"CtsDeviceLockTestCases": {
		"framework", // cts -> unstable
	},

	"CtsDeviceOwnerApp": {
		"framework", // cts -> unstable
	},

	"CtsDevicePolicySimTestCases": {
		"framework", // cts -> unstable
	},

	"CtsDevicePolicyTestCases": {
		"framework", // cts -> unstable
	},

	"CtsDocumentContentTestCases": {
		"framework", // cts -> unstable
	},

	"CtsDreamsTestCases": {
		"framework", // cts -> unstable
	},

	"CtsDrmTestCases": {
		"framework", // cts -> unstable
	},

	"CtsEmptyTestApp_RejectedByVerifier": {
		"framework", // cts -> unstable
	},

	"CtsEphemeralTestsEphemeralApp1": {
		"framework", // cts -> unstable
	},

	"CtsFgsBootCompletedTestCases": {
		"framework", // cts -> unstable
	},

	"CtsFgsBootCompletedTestCasesApi35": {
		"framework", // cts -> unstable
	},

	"CtsFgsStartTestHelperApi34": {
		"framework", // cts -> unstable
	},

	"CtsFgsStartTestHelperCurrent": {
		"framework", // cts -> unstable
	},

	"CtsFgsTimeoutTestCases": {
		"framework", // cts -> unstable
	},

	"CtsFileDescriptorTestCases": {
		"framework", // cts -> unstable
	},

	"CtsFingerprintTestCases": {
		"framework", // cts -> unstable
	},

	"CtsHostsideCompatChangeTestsApp": {
		"framework", // cts -> unstable
	},

	"CtsHostsideNetworkPolicyTestsApp2": {
		"framework", // cts -> unstable
	},

	"CtsIdentityTestCases": {
		"framework", // cts -> unstable
	},

	"CtsIkeTestCases": {
		"framework", // cts -> unstable
	},

	"CtsInstalledLoadingProgressDeviceTests": {
		"framework", // cts -> unstable
	},

	"CtsInstantAppTests": {
		"framework", // cts -> unstable
	},

	"CtsIntentSenderApp": {
		"framework", // cts -> unstable
	},

	"CtsJobSchedulerTestCases": {
		"framework", // cts -> unstable
	},

	"CtsKeystoreTestCases": {
		"framework", // cts -> unstable
	},

	"CtsLegacyNotification27TestCases": {
		"framework", // cts -> unstable
	},

	"CtsLibcoreTestCases": {
		"framework", // cts -> unstable
	},

	"CtsLibcoreWycheproofConscryptTestCases": {
		"framework", // cts -> unstable
	},

	"CtsListeningPortsTest": {
		"framework", // cts -> unstable
	},

	"CtsLocationCoarseTestCases": {
		"framework", // cts -> unstable
	},

	"CtsLocationFineTestCases": {
		"framework", // cts -> unstable
	},

	"CtsLocationNoneTestCases": {
		"framework", // cts -> unstable
	},

	"CtsLocationPrivilegedTestCases": {
		"framework", // cts -> unstable
	},

	"CtsManagedProfileApp": {
		"framework", // cts -> unstable
	},

	"CtsMediaAudioTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaBetterTogetherTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaCodecTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaDecoderTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaDrmFrameworkTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaEncoderTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaExtractorTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaMiscTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaMuxerTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaPerformanceClassTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaPlayerTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaProjectionSDK33TestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaProjectionSDK34TestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaProjectionTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaProviderTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaProviderTranscodeTests": {
		"framework", // cts -> unstable
	},

	"CtsMediaRecorderTestCases": {
		"framework", // cts -> unstable
	},

	"CtsMediaRouterHostSideTestBluetoothPermissionsApp": {
		"framework", // cts -> unstable
	},

	"CtsMediaRouterHostSideTestMediaRoutingControlApp": {
		"framework", // cts -> unstable
	},

	"CtsMediaRouterHostSideTestModifyAudioRoutingApp": {
		"framework", // cts -> unstable
	},

	"CtsMediaV2TestCases": {
		"framework", // cts -> unstable
	},

	"CtsMimeMapTestCases": {
		"framework", // cts -> unstable
	},

	"CtsModifyQuietModeEnabledApp": {
		"framework", // cts -> unstable
	},

	"CtsMusicRecognitionTestCases": {
		"framework", // cts -> unstable
	},

	"CtsNativeMediaAAudioTestCases": {
		"framework", // cts -> unstable
	},

	"CtsNetTestCases": {
		"framework", // cts -> unstable
	},

	"CtsNetTestCasesLegacyApi22": {
		"framework", // cts -> unstable
	},

	"CtsNetTestCasesMaxTargetSdk30": {
		"framework", // cts -> unstable
	},

	"CtsNetTestCasesMaxTargetSdk31": {
		"framework", // cts -> unstable
	},

	"CtsNetTestCasesMaxTargetSdk33": {
		"framework", // cts -> unstable
	},

	"CtsNetTestCasesUpdateStatsPermission": {
		"framework", // cts -> unstable
	},

	"CtsNfcTestCases": {
		"framework", // cts -> unstable
	},

	"CtsOnDeviceIntelligenceServiceTestCases": {
		"framework", // cts -> unstable
	},

	"CtsOnDevicePersonalizationTestCases": {
		"framework", // cts -> unstable
	},

	"CtsPackageInstallerApp": {
		"framework", // cts -> unstable
	},

	"CtsPackageManagerTestCases": {
		"framework", // cts -> unstable
	},

	"CtsPackageSchemeTestsWithoutVisibility": {
		"framework", // cts -> unstable
	},

	"CtsPackageSchemeTestsWithVisibility": {
		"framework", // cts -> unstable
	},

	"CtsPackageWatchdogTestCases": {
		"framework", // cts -> unstable
	},

	"CtsPermissionsSyncTestApp": {
		"framework", // cts -> unstable
	},

	"CtsPreservedSettingsApp": {
		"framework", // cts -> unstable
	},

	"CtsProtoTestCases": {
		"framework", // cts -> unstable
	},

	"CtsProviderTestCases": {
		"framework", // cts -> unstable
	},

	"CtsProxyMediaRouterTestHelperApp": {
		"framework", // cts -> unstable
	},

	"CtsRebootReadinessTestCases": {
		"framework", // cts -> unstable
	},

	"CtsResourcesLoaderTests": {
		"framework", // cts -> unstable
	},

	"CtsResourcesTestCases": {
		"framework", // cts -> unstable
	},

	"CtsSandboxedAdIdManagerTests": {
		"framework", // cts -> unstable
	},

	"CtsSandboxedAppSetIdManagerTests": {
		"framework", // cts -> unstable
	},

	"CtsSandboxedFledgeManagerTests": {
		"framework", // cts -> unstable
	},

	"CtsSandboxedMeasurementManagerTests": {
		"framework", // cts -> unstable
	},

	"CtsSandboxedTopicsManagerTests": {
		"framework", // cts -> unstable
	},

	"CtsSdkExtensionsTestCases": {
		"framework", // cts -> unstable
	},

	"CtsSdkSandboxInprocessTests": {
		"framework", // cts -> unstable
	},

	"CtsSecureElementTestCases": {
		"framework", // cts -> unstable
	},

	"CtsSecurityTestCases": {
		"framework", // cts -> unstable
	},

	"CtsSelinuxEphemeralTestCases": {
		"framework", // cts -> unstable
	},

	"CtsSelinuxTargetSdk25TestCases": {
		"framework", // cts -> unstable
	},

	"CtsSelinuxTargetSdk27TestCases": {
		"framework", // cts -> unstable
	},

	"CtsSelinuxTargetSdk28TestCases": {
		"framework", // cts -> unstable
	},

	"CtsSelinuxTargetSdk29TestCases": {
		"framework", // cts -> unstable
	},

	"CtsSelinuxTargetSdk30TestCases": {
		"framework", // cts -> unstable
	},

	"CtsSelinuxTargetSdkCurrentTestCases": {
		"framework", // cts -> unstable
	},

	"CtsSettingsDeviceOwnerApp": {
		"framework", // cts -> unstable
	},

	"CtsSharedUserMigrationTestCases": {
		"framework", // cts -> unstable
	},

	"CtsShortFgsTestCases": {
		"framework", // cts -> unstable
	},

	"CtsSimRestrictedApisTestCases": {
		"framework", // cts -> unstable
	},

	"CtsSliceTestCases": {
		"framework", // cts -> unstable
	},

	"CtsSpeechTestCases": {
		"framework", // cts -> unstable
	},

	"CtsStatsSecurityApp": {
		"framework", // cts -> unstable
	},

	"CtsSuspendAppsTestCases": {
		"framework", // cts -> unstable
	},

	"CtsSystemUiTestCases": {
		"framework", // cts -> unstable
	},

	"CtsTareTestCases": {
		"framework", // cts -> unstable
	},

	"CtsTelephonyTestCases": {
		"framework", // cts -> unstable
	},

	"CtsTetheringTest": {
		"framework", // cts -> unstable
	},

	"CtsThreadNetworkTestCases": {
		"framework", // cts -> unstable
	},

	"CtsTvInputTestCases": {
		"framework", // cts -> unstable
	},

	"CtsTvTunerTestCases": {
		"framework", // cts -> unstable
	},

	"CtsUsageStatsTestCases": {
		"framework", // cts -> unstable
	},

	"CtsUsbManagerTestCases": {
		"framework", // cts -> unstable
	},

	"CtsUserRestrictionTestCases": {
		"framework", // cts -> unstable
	},

	"CtsUtilTestCases": {
		"framework", // cts -> unstable
	},

	"CtsUwbTestCases": {
		"framework", // cts -> unstable
	},

	"CtsVcnTestCases": {
		"framework", // cts -> unstable
	},

	"CtsVideoCodecTestCases": {
		"framework", // cts -> unstable
	},

	"CtsVideoTestCases": {
		"framework", // cts -> unstable
	},

	"CtsViewReceiveContentTestCases": {
		"framework", // cts -> unstable
	},

	"CtsVirtualDevicesAppLaunchTestCases": {
		"framework", // cts -> unstable
	},

	"CtsVirtualDevicesAudioTestCases": {
		"framework", // cts -> unstable
	},

	"CtsVirtualDevicesCameraTestCases": {
		"framework", // cts -> unstable
	},

	"CtsVirtualDevicesSensorTestCases": {
		"framework", // cts -> unstable
	},

	"CtsVirtualDevicesTestCases": {
		"framework", // cts -> unstable
	},

	"CtsWearableSensingServiceTestCases": {
		"framework", // cts -> unstable
	},

	"CtsWebViewCompatChangeApp": {
		"framework", // cts -> unstable
	},

	"CtsWidgetTestCases": {
		"framework", // cts -> unstable
	},

	"CtsWidgetTestCases29": {
		"framework", // cts -> unstable
	},

	"CtsWifiNonUpdatableTestCases": {
		"framework", // cts -> unstable
	},

	"CtsWifiTestCases": {
		"framework", // cts -> unstable
	},

	"CtsWindowManagerExternalApp": {
		"framework", // cts -> unstable
	},

	"CtsWindowManagerTestCases": {
		"framework", // cts -> unstable
	},

	"CtsZipValidateApp": {
		"framework", // cts -> unstable
	},

	"CVE-2021-0965": {
		"framework", // cts -> unstable
	},

	"device_config_reboot_flags_java_lib": {
		"ext",       // apex [com.android.configinfrastructure] -> system
		"framework", // apex [com.android.configinfrastructure] -> system
	},

	"devicelockcontroller-lib": {
		"modules-utils-expresslog", // apex [com.android.devicelock] -> apex [com.android.btservices, com.android.car.framework]
	},

	"FederatedCompute": {
		"auto_value_annotations", // apex [com.android.ondevicepersonalization] -> apex [com.android.adservices, com.android.extservices, com.android.extservices_tplus]
	},

	"framework-adservices.impl": {
		"adservices_flags_lib", // apex [com.android.adservices, com.android.extservices] -> system
	},

	"framework-bluetooth.impl": {
		"app-compat-annotations", // apex [com.android.btservices] -> system
	},

	"framework-configinfrastructure.impl": {
		"configinfra_framework_flags_java_lib", // apex [com.android.configinfrastructure] -> system
	},

	"framework-connectivity-t.impl": {
		"app-compat-annotations",            // apex [com.android.tethering] -> system
		"framework-connectivity-pre-jarjar", // apex [com.android.tethering] -> system
	},

	"framework-connectivity.impl": {
		"app-compat-annotations", // apex [com.android.tethering] -> system
	},

	"framework-ondevicepersonalization.impl": {
		"app-compat-annotations",            // apex [com.android.ondevicepersonalization] -> system
		"ondevicepersonalization_flags_lib", // apex [com.android.ondevicepersonalization] -> system
	},

	"framework-pdf-v.impl": {
		"app-compat-annotations",      // apex [com.android.mediaprovider, test_com.android.mediaprovider] -> system
		"modules-utils-preconditions", // apex [com.android.mediaprovider, test_com.android.mediaprovider] -> apex [com.android.adservices, com.android.appsearch, com.android.cellbroadcast, com.android.extservices, com.android.ondevicepersonalization, com.android.tethering, com.android.uwb, com.android.wifi, test_com.android.cellbroadcast, test_com.android.wifi]
	},

	"framework-pdf.impl": {
		"modules-utils-preconditions", // apex [com.android.mediaprovider, test_com.android.mediaprovider] -> apex [com.android.adservices, com.android.appsearch, com.android.cellbroadcast, com.android.extservices, com.android.ondevicepersonalization, com.android.tethering, com.android.uwb, com.android.wifi, test_com.android.cellbroadcast, test_com.android.wifi]
	},

	"framework-permission-s.impl": {
		"app-compat-annotations", // apex [com.android.permission, test_com.android.permission] -> system
	},

	"framework-wifi.impl": {
		"aconfig_storage_reader_java", // apex [com.android.wifi, test_com.android.wifi] -> system
		"app-compat-annotations",      // apex [com.android.wifi, test_com.android.wifi] -> system
	},

	"grpc-java-core-internal": {
		"gson",             // apex [com.android.adservices, com.android.devicelock, com.android.extservices] -> apex [com.android.virt]
		"perfmark-api-lib", // apex [com.android.adservices, com.android.devicelock, com.android.extservices] -> system
	},

	"httpclient_impl": {
		"httpclient_api", // apex [com.android.tethering] -> system
	},

	"IncrementalTestAppValidator": {
		"framework", // cts -> unstable
	},

	"libcore-aconfig-flags-lib": {
		"framework-api-annotations-lib", // apex [com.android.art, com.android.art.debug, com.android.art.testing, test_imgdiag_com.android.art, test_jitzygote_com.android.art] -> system
	},

	"loadlibrarytest_product_app": {
		"libnativeloader_vendor_shared_lib", // product -> vendor
	},

	"loadlibrarytest_testlib": {
		"libnativeloader_vendor_shared_lib", // system -> vendor
	},

	"MctsMediaBetterTogetherTestCases": {
		"framework", // cts -> unstable
	},

	"MctsMediaCodecTestCases": {
		"framework", // cts -> unstable
	},

	"MctsMediaDecoderTestCases": {
		"framework", // cts -> unstable
	},

	"MctsMediaDrmFrameworkTestCases": {
		"framework", // cts -> unstable
	},

	"MctsMediaEncoderTestCases": {
		"framework", // cts -> unstable
	},

	"MctsMediaExtractorTestCases": {
		"framework", // cts -> unstable
	},

	"MctsMediaMiscTestCases": {
		"framework", // cts -> unstable
	},

	"MctsMediaMuxerTestCases": {
		"framework", // cts -> unstable
	},

	"MctsMediaPlayerTestCases": {
		"framework", // cts -> unstable
	},

	"MctsMediaRecorderTestCases": {
		"framework", // cts -> unstable
	},

	"MctsMediaTranscodingTestCases": {
		"framework", // cts -> unstable
	},

	"MctsMediaV2TestCases": {
		"framework", // cts -> unstable
	},

	"MediaProvider": {
		"app-compat-annotations", // apex [com.android.mediaprovider, test_com.android.mediaprovider] -> system
	},

	"mediaprovider_flags_java_lib": {
		"ext",       // apex [com.android.mediaprovider, test_com.android.mediaprovider] -> system
		"framework", // apex [com.android.mediaprovider, test_com.android.mediaprovider] -> system
	},

	"MockSatelliteGatewayServiceApp": {
		"framework", // cts -> unstable
	},

	"MockSatelliteServiceApp": {
		"framework", // cts -> unstable
	},

	"net-utils-device-common-netlink": {
		"net-utils-device-common-struct-base", // apex [com.android.tethering] -> system
	},

	"net-utils-device-common-struct": {
		"net-utils-device-common-struct-base", // apex [com.android.tethering] -> system
	},

	"NfcNciApex": {
		"android.permission.flags-aconfig-java", // apex [com.android.nfcservices] -> apex [com.android.permission, test_com.android.permission]
	},

	"okhttp-norepackage": {
		"okhttp-android-util-log", // apex [com.android.adservices, com.android.devicelock, com.android.extservices] -> system
	},

	"ondevicepersonalization-plugin-lib": {
		"auto_value_annotations", // apex [com.android.ondevicepersonalization] -> apex [com.android.adservices, com.android.extservices, com.android.extservices_tplus]
	},

	"opencensus-java-api": {
		"auto_value_annotations", // apex [com.android.devicelock] -> apex [com.android.adservices, com.android.extservices, com.android.extservices_tplus]
	},

	"PermissionController-lib": {
		"safety-center-annotations", // apex [com.android.permission, test_com.android.permission] -> system
	},

	"PlatformProperties": {
		"sysprop-library-stub-platform", // apex [com.android.btservices, com.android.nfcservices, com.android.tethering, com.android.virt, com.android.wifi, test_com.android.wifi] -> system
	},

	"safety-center-config": {
		"safety-center-annotations", // apex [com.android.permission, test_com.android.permission] -> system
	},

	"safety-center-internal-data": {
		"safety-center-annotations", // apex [com.android.permission, test_com.android.permission] -> system
	},

	"safety-center-pending-intents": {
		"safety-center-annotations", // apex [com.android.permission, test_com.android.permission] -> system
	},

	"safety-center-persistence": {
		"safety-center-annotations", // apex [com.android.permission, test_com.android.permission] -> system
	},

	"safety-center-resources-lib": {
		"safety-center-annotations", // apex [com.android.permission, test_com.android.permission] -> system
	},

	"SdkSandboxManagerDisabledTests": {
		"framework", // cts -> unstable
	},

	"SdkSandboxManagerTests": {
		"framework", // cts -> unstable
	},

	"service-art.impl": {
		"auto_value_annotations", // apex [com.android.art, com.android.art.debug, com.android.art.testing, test_imgdiag_com.android.art, test_jitzygote_com.android.art] -> apex [com.android.adservices, com.android.extservices, com.android.extservices_tplus]
	},

	"service-bluetooth-pre-jarjar": {
		"framework-bluetooth-pre-jarjar", // apex [com.android.btservices] -> system
		"service-bluetooth.change-ids",   // apex [com.android.btservices] -> system
	},

	"service-connectivity": {
		"libprotobuf-java-nano", // apex [com.android.tethering] -> apex [com.android.wifi, test_com.android.wifi]
	},

	"service-connectivity-pre-jarjar": {
		"framework-connectivity-pre-jarjar", // apex [com.android.tethering] -> system
	},

	"service-connectivity-protos": {
		"libprotobuf-java-nano", // apex [com.android.tethering] -> apex [com.android.wifi, test_com.android.wifi]
	},

	"service-connectivity-tiramisu-pre-jarjar": {
		"framework-connectivity-pre-jarjar",   // apex [com.android.tethering] -> system
		"framework-connectivity-t-pre-jarjar", // apex [com.android.tethering] -> system
	},

	"service-entitlement": {
		"auto_value_annotations", // apex [com.android.wifi, test_com.android.wifi] -> apex [com.android.adservices, com.android.extservices, com.android.extservices_tplus]
	},

	"service-entitlement-api": {
		"auto_value_annotations", // apex [com.android.wifi, test_com.android.wifi] -> apex [com.android.adservices, com.android.extservices, com.android.extservices_tplus]
	},

	"service-entitlement-data": {
		"auto_value_annotations", // apex [com.android.wifi, test_com.android.wifi] -> apex [com.android.adservices, com.android.extservices, com.android.extservices_tplus]
	},

	"service-entitlement-impl": {
		"auto_value_annotations", // apex [com.android.wifi, test_com.android.wifi] -> apex [com.android.adservices, com.android.extservices, com.android.extservices_tplus]
	},

	"service-healthfitness.impl": {
		"modules-utils-preconditions", // apex [com.android.healthfitness] -> apex [com.android.adservices, com.android.appsearch, com.android.cellbroadcast, com.android.extservices, com.android.ondevicepersonalization, com.android.tethering, com.android.uwb, com.android.wifi, test_com.android.cellbroadcast, test_com.android.wifi]
	},

	"service-networksecurity-pre-jarjar": {
		"framework-connectivity-pre-jarjar", // apex [com.android.tethering] -> system
	},

	"service-permission.impl": {
		"jsr305",                    // apex [com.android.permission, test_com.android.permission] -> apex [com.android.adservices, com.android.devicelock, com.android.extservices, com.android.healthfitness, com.android.media, com.android.mediaprovider, test_com.android.media, test_com.android.mediaprovider]
		"safety-center-annotations", // apex [com.android.permission, test_com.android.permission] -> system
	},

	"service-remoteauth-pre-jarjar": {
		"framework-connectivity-pre-jarjar",   // apex [com.android.tethering] -> system
		"framework-connectivity-t-pre-jarjar", // apex [com.android.tethering] -> system
	},

	"service-thread-pre-jarjar": {
		"framework-connectivity-pre-jarjar",   // apex [com.android.tethering] -> system
		"framework-connectivity-t-pre-jarjar", // apex [com.android.tethering] -> system
	},

	"service-uwb-pre-jarjar": {
		"framework-uwb-pre-jarjar", // apex [com.android.uwb] -> system
	},

	"service-wifi": {
		"auto_value_annotations", // apex [com.android.wifi, test_com.android.wifi] -> apex [com.android.adservices, com.android.extservices, com.android.extservices_tplus]
	},

	"TelephonyDeviceTest": {
		"framework", // cts -> unstable
	},

	"tensorflowlite_java": {
		"android-support-annotations", // apex [com.android.adservices, com.android.extservices, com.android.ondevicepersonalization] -> system
	},

	"TestExternalImsServiceApp": {
		"framework", // cts -> unstable
	},

	"TestSmsRetrieverApp": {
		"framework", // cts -> unstable
	},

	"TetheringApiCurrentLib": {
		"connectivity-internal-api-util", // apex [com.android.tethering] -> system
	},

	"TetheringNext": {
		"connectivity-internal-api-util", // apex [com.android.tethering] -> system
	},

	"tetheringstatsprotos": {
		"ext",       // apex [com.android.tethering] -> system
		"framework", // apex [com.android.tethering] -> system
	},

	"uwb_aconfig_flags_lib": {
		"ext",       // apex [com.android.uwb] -> system
		"framework", // apex [com.android.uwb] -> system
	},

	"uwb_androidx_backend": {
		"android-support-annotations", // apex [com.android.tethering] -> system
	},

	"wifi-service-pre-jarjar": {
		"app-compat-annotations",    // apex [com.android.wifi, test_com.android.wifi] -> system
		"auto_value_annotations",    // apex [com.android.wifi, test_com.android.wifi] -> apex [com.android.adservices, com.android.extservices, com.android.extservices_tplus]
		"framework-wifi-pre-jarjar", // apex [com.android.wifi, test_com.android.wifi] -> system
		"jsr305",                    // apex [com.android.wifi, test_com.android.wifi] -> apex [com.android.adservices, com.android.devicelock, com.android.extservices, com.android.healthfitness, com.android.media, com.android.mediaprovider, test_com.android.media, test_com.android.mediaprovider]
	},
}
