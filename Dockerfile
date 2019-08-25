FROM golang:1.11.4 as build

RUN mkdir /twcc-self-checker
WORKDIR /twcc-self-checker

# COPY go.mod and go.sum files to the workspace
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
#RUN go mod download

COPY . .
RUN make build-in-docker

FROM alpine:latest
COPY --from=build /twcc-self-checker/bin/twcc-self-checker /
CMD ["/twcc-self-checker","-v","1"]