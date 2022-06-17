FROM golang as build

COPY . /src

WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux go build -o notifications ./cmd/notifications


FROM alpine

COPY --from=build /src/notifications .

RUN apk --no-cache add ca-certificates

CMD ["/notifications"]




