FROM golang
WORKDIR /app
COPY . /app
RUN go test ./gc -v
RUN scripts/build.sh
