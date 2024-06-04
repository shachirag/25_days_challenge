# Build Stage
FROM golang:1.21 AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o main server.go

# Run Stage
FROM gcr.io/distroless/static-debian11
COPY --from=build /app/main /
# COPY ./firebase-sa-creds.json .
# COPY ./public /public

EXPOSE 8080
CMD ["/main"]
