# Use the official Golang image as the base image.
# Specify the version of Go to use.
FROM golang:1.22-alpine as builder

# Set the working directory inside the container.
WORKDIR /app

# Copy the source code into the container.
COPY . .

RUN pwd && ls -la
# Build the Go application.  This creates an optimized executable.
RUN go build -o main

# Use a minimal base image for the final image.  This makes the image smaller.
FROM alpine:latest

# Copy the executable from the builder stage.
COPY --from=builder /app/main /app/main

# Expose the port that the application listens on.
EXPOSE 8080

# Set the entrypoint for the container.
CMD ["/app/main"]
