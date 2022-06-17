FROM golang as build

COPY . /src

WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux go build -o orders ./cmd/orders


FROM alpine

COPY --from=build /src/orders .

RUN apk --no-cache add ca-certificates

CMD ["/orders"]




