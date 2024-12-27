## Soong Android Resource Compilation

The Android build process involves several steps to compile resources into a format that the Android app can use
efficiently in android_library, android_app and android_test modules.  See the
[resources documentation](https://developer.android.com/guide/topics/resources/providing-resources) for general
information on resources (with a focus on building with Gradle).

For all modules, AAPT2 compiles resources provided by directories listed in the resource_dirs directory (which is
implicitly set to `["res"]` if unset, but can be overridden by setting the `resource_dirs` property).

## android_library with resource processor
For an android_library with resource processor enabled (currently by setting `use_resource_processor: true`, but will be
enabled by default in the future):
- AAPT2 generates the `package-res.apk` file with a resource table that contains all resources from the current
android_library module.  `package-res.apk` files from transitive dependencies are passed to AAPT2 with the `-I` flag to
resolve references to resources from dependencies.
- AAPT2 generates an R.txt file that lists all the resources provided by the current android_library module.
- ResourceProcessorBusyBox reads the `R.txt` file for the current android_library and produces an `R.jar` with an
`R.class` in the package listed in the android_library's `AndroidManifest.xml` file that contains java fields for each
resource ID.  The resource IDs are non-final, as the final IDs will not be known until the resource table of the final
android_app or android_test module is built.
- The android_library's java and/or kotlin code is compiled with the generated `R.jar` in the classpath, along with the
`R.jar` files from all transitive android_library dependencies.

## android_app or android_test with resource processor
For an android_app or android_test with resource processor enabled (currently by setting `use_resource_processor: true`,
but will be enabled by default in the future):
- AAPT2 generates the `package-res.apk` file with a resource table that contains all resources from the current
android_app or android_test, as well as all transitive android_library modules referenced via `static_libs`.  The
current module is overlaid on dependencies so that resources from the current module replace resources from dependencies
in the case of conflicts.
- AAPT2 generates an R.txt file that lists all the resources provided by the current android_app or android_test, as
well as all transitive android_library modules referenced via `static_libs`.  The R.txt file contains the final resource
ID for each resource.
- ResourceProcessorBusyBox reads the `R.txt` file for the current android_app or android_test, as well as all transitive
android_library modules referenced via `static_libs`, and produces an `R.jar` with an `R.class` in the package listed in
the android_app or android_test's `AndroidManifest.xml` file that contains java fields for all local or transitive
resource IDs.  In addition, it creates an `R.class` in the package listed in each android_library dependency's
`AndroidManifest.xml` file that contains final resource IDs for the resources that were found in that library.
- The android_app or android_test's java and/or kotlin code is compiled with the current module's `R.jar` in the
classpath, but not the `R.jar` files from transitive android_library dependencies.  The `R.jar` file is also merged into
the program  classes that are dexed and placed in the final APK.

## android_app, android_test or android_library without resource processor
For an android_app, android_test or android_library without resource processor enabled (current the default, or
explicitly set with `use_resource_processor: false`):
- AAPT2 generates the `package-res.apk` file with a resource table that contains all resources from the current
android_app, android_test or android_library module, as well as all transitive android_library modules referenced via
`static_libs`.  The current module is overlaid on dependencies so that resources from the current module replace
resources from dependencies in the case of conflicts.
- AAPT2 generates an `R.java` file in the package listed in each the current module's `AndroidManifest.xml` file that
contains resource IDs for all resources from the current module as well as all transitive android_library modules
referenced via `static_libs`.  The same `R.java` containing all local and transitive resources is also duplicated into
every package listed in an `AndroidManifest.xml` file in any static `android_library` dependency.
- The module's java and/or kotlin code is compiled along with all the generated `R.java` files.


## Downsides of legacy resource compilation without resource processor

Compiling resources without using the resource processor results in a generated R.java source file for every transitive
package that contains every transitive resource.  For modules with large transitive dependency trees this can be tens of
thousands of resource IDs duplicated in tens to a hundred java sources.  These java sources all have to be compiled in
every successive module in the dependency tree, and then the final R8 step has to drop hundreds of thousands of
unreferenced fields.  This results in significant build time and disk usage increases over building with resource
processor.

## Converting to compilation with resource processor

### Reference resources using the package name of the module that includes them.
Converting an android_library module to build with resource processor requires fixing any references to resources
provided by android_library dependencies to reference the R classes using the package name found in the
`AndroidManifest.xml` file of the dependency.  For example, when referencing an androidx resource:
```java
View.inflate(mContext, R.layout.preference, null));
```
must be replaced with:
```java
View.inflate(mContext, androidx.preference.R.layout.preference, null));
```

### Use unique package names for each module in `AndroidManifest.xml`

Each module will produce an `R.jar` containing an `R.class` in the package specified in it's `AndroidManifest.xml`.
If multiple modules use the same package name they will produce conflicting `R.class` files, which can cause some
resource IDs to appear to be missing.

If existing code has multiple modules that contribute resources to the same package, one option is to move all the
resources into a single resources-only `android_library` module with no code, and then depend on that from all the other
modules.