## nspawn-oci

Wrapper for systemd-nspawn to run [OpenContainer runtime
bundles](https://github.com/opencontainers/runtime-spec)

### Building

```bash
go get github.com/vbatts/nspawn-oci
```

### Usage

Using [skopeo](https://github.com/projectatomic/skopeo) and
[oci-image-tool](https://github.com/opencontainers/image-spec/tree/master/cmd/oci-image-tool):

```bash
skopeo copy docker://busybox oci:busybox-oci
mkdir busybox-bundle
oci-image-tool create-runtime-bundle --ref latest busybox-oci busybox-bundle
cd busybox-bundle && sudo nspawn-oci .
```

