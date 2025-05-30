FROM golang:1.23.4

# Install basic development tools
RUN apt update && apt install -y git

# Set the working directory
WORKDIR /app

# Initialize a new module
RUN go mod init go-sample

# Install Go tools with specific versions
RUN go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest

# Clean up the temporary module
RUN rm go.mod

# Ensure go.mod is present when mounting
COPY go.mod* go.sum* ./
RUN go mod download

# Add a default command
CMD ["go", "run", "cmd/api/main.go"] 