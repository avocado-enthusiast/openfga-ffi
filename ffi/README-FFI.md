# Info

This uses unix sockets instead of TCP to allow for utilizing openfga in-process. It also sets the configuration to disable optional services.

# Building

To build, you can run the build-*.sh scripts.

Both have only been tested on mac.

The *.so files will be dropped in `ffi/`.

# Updates

To update from a new release of openfga, pull in the latest changes from that project, checkout the tag, create a new branch (`v1.9.6-ffi`), apply the changes from the last release `v1.9.5-ffi` with the FFI scripts and codes, and resolve any conflicts. Then review the changes and test.

Since the changes to the original source are minimal, there should be few conflicts, and any updates from the main project should be quick to manually examine.

# Release

The build binaries under `ffi/` can be uploaded on a tagged release (e.g., `v1.9.5-ffi`).