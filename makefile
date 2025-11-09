BINARY=velcro
BUILD_DIR=build

.PHONY: build run

# builds the binary
build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) .

# builds & runs in one command
dev: build
	$(BUILD_DIR)/$(BINARY) $(filter-out $@,$(MAKECMDGOALS))

# runs the binary 
run:
	$(BUILD_DIR)/$(BINARY) $(filter-out $@,$(MAKECMDGOALS))

# dummy target to prevent make from erroring
%:
	@: