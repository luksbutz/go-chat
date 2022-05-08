BINARY_NAME=goChat

## build: Build binary
build:
	@echo "Building..."
	go build -o ${BINARY_NAME} ./cmd/web
	@echo "Binary built!"

## run: builds and runs the application
run: build
	@echo "Starting..."
	./${BINARY_NAME} &
	@echo "Started!"

## clean: runs go clean and deletes binary
clean:
	@echo "Cleaning..."
	@go clean
	@rm ${BINARY_NAME}
	@echo "Cleaned!"

## start: an alias to run
start: run

## stop: stops the running application
stop:
	@echo "Stopping..."
	@-pkill -SIGTERM -f "./${BINARY_NAME}"
	@echo "Stopped!"

## restart: stops and starts the running application
restart: stop start