# p47-network-controller
custom-controller

# build image
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o server ./main.go ./controller.go