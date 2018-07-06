#!/bin/bash
set -eux

release=28
label=vrutkovs/telek8s
cmd=/bin/telek8s

# run the build
dnf install -y golang godep
cd /go/src/github.com/$label
dep ensure
go build

# Install podman and buildah
dnf install -y podman buildah

# make tmpfs for /var/lib/container
mkdir -p /var/lib/containers/storage
mount -t tmpfs -o size=20G tmpfs /var/lib/containers/storage

# build a minimal image
newcontainer=$(buildah from scratch)
scratchmnt=$(buildah mount $newcontainer)

# install the packages
dnf install --installroot $scratchmnt --release $release coreutils --setopt install_weak_deps=false -y
dnf clean all -y --installroot $scratchmnt --releasever $release

cp telek8s $scratchmnt/usr/local/bin

buildah config --cmd /usr/local/bin/telek8s $newcontainer

# set some config info
buildah config --label name=$label $newcontainer

# commit the image
buildah unmount $newcontainer
buildah commit $newcontainer $label

podman push localhost/$label docker.io/$label
