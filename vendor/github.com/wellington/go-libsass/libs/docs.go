// libs is a direct mapping to libsass C API. This package should not be
// necessary to do any high level operations. It is instead intended
// that go-libsass handles everything needed in your project.
//
// In case where low level API calls are necessary, use this package
// to do so. For more details on what these methods do see:
// https://github.com/sass/libsass/wiki/API-Documentation
//
// For the most part, consider this a hand curated SWIG of libsass C API.
package libs
