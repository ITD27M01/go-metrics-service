FROM golang:1.17

WORKDIR /app

RUN curl -L -o /usr/bin/statictest \
    https://github.com/Yandex-Practicum/go-autotests-bin/releases/latest/download/statictest; \
    chmod +x /usr/bin/statictest

RUN curl -L -o /usr/bin/devopstest \
    https://github.com/Yandex-Practicum/go-autotests-bin/releases/latest/download/devopstest; \
    chmod +x /usr/bin/devopstest

COPY go.* ./
RUN go mod download

COPY . ./

RUN make build

ENTRYPOINT ["go", "vet", "-vettool=/usr/bin/statictest", "./..."]
