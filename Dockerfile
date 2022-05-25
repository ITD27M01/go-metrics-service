FROM golang:1.17

WORKDIR /app
COPY . /app/

RUN make update; make build

RUN curl -L -o /usr/bin/statictest \
    https://github.com/Yandex-Practicum/go-autotests-bin/releases/latest/download/statictest; \
    chmod +x /usr/bin/statictest

RUN curl -L -o /usr/bin/devopstest \
    https://github.com/Yandex-Practicum/go-autotests-bin/releases/latest/download/devopstest; \
    chmod +x /usr/bin/devopstest

ENTRYPOINT ["go", "vet", "-vettool=/usr/bin/statictest", "./..."]
