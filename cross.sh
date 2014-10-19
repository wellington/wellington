for GOOS in darwin linux; do
  for GOARCH in 386 amd64; do
    go build -v -o myapp-$GOOS-$GOARCH
  done
done
