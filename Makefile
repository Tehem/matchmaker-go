.PHONY: test test-coverage test-coverage-html

# Run all tests
test:
	go test ./... -v

# Run tests with coverage and generate coverage report
test-coverage:
	go test ./... -coverprofile=coverage.out

# Generate HTML coverage report
test-coverage-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html

# Clean up coverage files
clean:
	rm -f coverage.out coverage.html 