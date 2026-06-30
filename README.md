## go-mpv
[![Build Status](https://github.com/gen2brain/go-mpv/actions/workflows/build.yml/badge.svg)](https://github.com/gen2brain/go-mpv/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/gen2brain/go-mpv.svg)](https://pkg.go.dev/github.com/gen2brain/go-mpv)

> Go bindings for [libmpv](https://mpv.io/).


### Build tags

* `nocgo` - use [purego](https://github.com/ebitengine/purego) implementation (can also be used with `CGO_ENABLED=0`)
* `pkgconfig` - use pkg-config
* `static` - use static library (used with `pkgconfig`)


### License

The bindings in this repository are licensed under the [MIT](LICENSE) license.

They depend on [libmpv](https://mpv.io/), which is licensed under LGPLv2.1+ (or
GPLv2+ when built with GPL-only components). A copy of the LGPL is included as
[LICENSE.LGPL](LICENSE.LGPL); when you distribute a binary that links libmpv
(statically or via `dlopen`), its terms apply to that library.

