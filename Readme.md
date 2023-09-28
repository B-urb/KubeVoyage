# KubeVoyage: Kubernetes Authentication Proxy

Embarking on a secure journey in Kubernetes.

`KubeVoyage` is a Kubernetes authentication proxy designed to streamline user access to various sites. Built with a Svelte frontend, a Go backend, and an SQL database, it offers a robust solution for managing user access in a Kubernetes environment.

![KubeVoyage Logo](path_to_logo.png)  <!-- If you have a logo, replace 'path_to_logo.png' with its path -->

## Features

- **User Management**: Admins can grant or deny access to users.
- **Two Roles**: Users can either be admins with full access or regular users with specific site access.
- **SSO Integration**: Single Sign-On with platforms like Google, GitHub, and Microsoft.
- **Helm Deployment**: Easily deploy on Kubernetes using the provided Helm chart.

## Getting Started

### Prerequisites

- Go (version 1.x+)
- Node.js and npm
- Kubernetes cluster (for deployment)
- Helm (for deployment)

### Installation

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/yourusername/kubevoyage.git
   cd kubevoyage
   ```

2. **Backend Setup**:

   Navigate to the backend directory and fetch the required Go modules:

   ```bash
   cd backend
   go mod download
   ```

3. **Frontend Setup**:

   Navigate to the frontend directory and install the required npm packages:

   ```bash
   cd frontend
   npm install
   ```

### Running Locally

1. **Backend**:

   From the backend directory:

   ```bash
   go run .
   ```

2. **Frontend**:

   From the frontend directory:

   ```bash
   npm run dev
   ```

Visit `http://localhost:8080` in your browser.

### Deployment

Use the provided Helm chart to deploy `KubeVoyage` to your Kubernetes cluster:

```bash
helm install kubevoyage ./path_to_helm_chart
```

## Testing

To run tests for the backend:

```bash
cd backend
go test ./...
```

## Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you'd like to change.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

---

This README provides a basic overview of the project. Depending on the specifics of your project, you might want to add more sections, such as "API Documentation", "Known Issues", "Future Enhancements", etc. Remember, a well-documented README can significantly improve the user experience for those trying to understand or contribute to your project.