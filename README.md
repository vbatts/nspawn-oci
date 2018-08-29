## nspawn-oci

Wrapper for systemd-nspawn to run [OpenContainer runtime
bundles](https://github.com/opencontainers/runtime-spec)

### Building

This project uses the Go programming language and is tested with the [Go
compiler](https://golang.org/dl/). (Results with gccgo may vary)

```shell
go get -u -v github.com/vbatts/nspawn-oci
```

### Usage

Using [skopeo](https://github.com/projectatomic/skopeo) and
[oci-image-tool](https://github.com/opencontainers/image-spec/tree/master/cmd/oci-image-tool):

```shell
skopeo copy docker://busybox oci:busybox-oci
oci-image-tool create --ref platform.os=linux ./busybox-oci/ ./busybox-bundle/
cd busybox-bundle && sudo nspawn-oci .
```

