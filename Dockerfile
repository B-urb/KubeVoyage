FROM golang:1.22
# Define an argument for the architecture, which will be passed from the build command
ARG TARGETARCH
LABEL authors="bjornurban"
EXPOSE 8080:8080

WORKDIR /kubevoyage
# Copy the correct binary based on the architecture argument
COPY backend/build/kubevoyage-${TARGETARCH} ./bin/kubevoyage
# Copy frontend and backend files
COPY frontend/public ./public
COPY backend/build ./bin

# Ensure the binary has executable permissions
RUN chmod +x ./bin/kubevoyage

ENTRYPOINT ["./bin/kubevoyage"]