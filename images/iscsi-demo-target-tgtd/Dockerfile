FROM debian:sid

MAINTAINER "Fabian Deutsch" \<fabiand@redhat.com\>
ENV container docker

RUN apt-get update && apt-get install -y \
  tgt bash curl bzip2 qemu-utils

RUN mkdir -p /volume

# Add alpine image
RUN curl \
      https://nl.alpinelinux.org/alpine/v3.5/releases/x86_64/alpine-virt-3.5.1-x86_64.iso \
      > /volume/alpine.iso

# Add cirros
RUN curl \
      http://download.cirros-cloud.net/0.3.5/cirros-0.3.5-x86_64-disk.img \
      > /volume/cirros.img \
      && qemu-img convert -O raw /volume/cirros.img /volume/cirros.rawa \
      && rm volume/cirros.img

ADD run-tgt.sh /

EXPOSE 3260

CMD ["/run-tgt.sh"]
