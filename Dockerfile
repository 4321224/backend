FROM golang:1.16-alpine

# Set working directory di dalam container
WORKDIR /app

# Salin go mod dan sum files
COPY go.mod go.sum ./

# Download semua dependensi
RUN go mod download

# Salin source code dari current directory ke Working Directory di dalam container
COPY . .

 # Build aplikasi Go untuk membuat binary executable.
 RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

 ################ Tahap kedua ###############
 FROM alpine:latest  
 RUN apk --no-cache add ca-certificates

 WORKDIR /root/

 # Copy binary dari tahap pertama 
 COPY --from=0 /app/main .
 
 CMD ["./main"]  
