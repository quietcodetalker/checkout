FROM golang as build

COPY . /src

WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux go build -o billing ./cmd/billing


FROM alpine

COPY --from=build /src/billing .

RUN apk --no-cache add ca-certificates

CMD ["/billing"]




