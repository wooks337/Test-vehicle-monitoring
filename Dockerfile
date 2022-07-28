# Start fromgolang:1.18.4-alpine3.15 base image
FROM golang:1.18.4-alpine3.15 as service_builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ./

# Add Maintainer Info
LABEL maintainer="Marty"

RUN go build -v -o /Test-vehicle-monitoring

#COPY --from=service_builder /Test-vehicle-monitoring /Test-vehicle-monitoring

# Expose port number
EXPOSE 3000

CMD ["/Test-vehicle-monitoring"]