#!/bin/bash
set -eux

# Do `podman login` and `oc login` previosly!

release=28
baselabel=vrutkovs/base-28
label=vrutkovs/telek8s
cmd=/bin/telek8s
deps="glibc ca-certificates"

# pre - install tools
dnf install -y podman buildah origin-clients

# prepare tmpfs storage and build base container
if ! ls /var/lib/containers/storage; then
    mkdir -p /var/lib/containers/storage
    mount -t tmpfs -o size=20G tmpfs /var/lib/containers/storage
fi

# build base container
if ! podman images localhost/$baselabel; then
    cachecontainer=$(buildah from scratch)
    scratchmnt=$(buildah mount $cachecontainer)
    dnf install --installroot $scratchmnt --release $release $deps --setopt install_weak_deps=false -y
    dnf clean all -y --installroot $scratchmnt --releasever $release
    buildah unmount $cachecontainer
    buildah commit $cachecontainer $baselabel
fi

# run the build
dnf install -y golang godep
cd /go/src/github.com/$label
dep ensure
go build

# put the binary in the container and build a new one
newcontainer=$(buildah from ${baselabel})
scratchmnt=$(buildah mount $newcontainer)
cp telek8s $scratchmnt/usr/local/bin
buildah config --cmd /usr/local/bin/telek8s $newcontainer
buildah config --label name=$label $newcontainer
buildah unmount $newcontainer
buildah commit $newcontainer $label

podman push localhost/$label docker.io/$label
oc import-image telek8s -n telek8s
