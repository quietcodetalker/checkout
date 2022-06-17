FROM golang as build

COPY . /src

WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux go build -o stock ./cmd/stock


FROM alpine

COPY --from=build /src/stock .

RUN apk --no-cache add ca-certificates

CMD ["/stock"]




