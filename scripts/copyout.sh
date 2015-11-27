#/bin/sh
# replacement for make copyout, so container doesn't need
# make tools

chown $EUID:$EGID $GOPATH/bin/wt
cp $GOPATH/bin/wt /tmp
#chown -R $(EUID):$(EGID) /build/libsass
mkdir -p /tmp/lib64
cp /usr/lib/libstdc++.so.6 /tmp/lib64
cp /usr/lib/libgcc_s.so.1 /tmp/lib64
chown -R $EUID:$EGID /tmp
