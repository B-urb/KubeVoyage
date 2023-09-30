FROM golang:1.21

LABEL authors="bjornurban"
EXPOSE 8080:8080

RUN mkdir /kubevoyage
RUN mkdir /kubevoyage/bin
COPY public /kubevoyage/public
COPY build /kubevoyage/bin
ENTRYPOINT ["./bin/kubevoyage"]