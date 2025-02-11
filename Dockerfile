FROM alpine/git:v2.47.1

ENTRYPOINT ["/run.sh"]

COPY ./build/linux/jx-registry /usr/bin/jx-registry
COPY run.sh /run.sh