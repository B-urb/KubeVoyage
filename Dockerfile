FROM golang:1.21

LABEL authors="bjornurban"
EXPOSE 8080:8080

WORKDIR /kubevoyage

# Copy frontend and backend files
COPY frontend/public ./public
COPY backend/build ./bin

# Ensure the binary has executable permissions
RUN chmod +x ./bin/kubevoyage

ENTRYPOINT ["./bin/kubevoyage"]