# Development

## Getting Started

Firstly, you may want to [Ramp up](#ramp-up) on Kubernetes and Custom Resource Definitions (CRDs) as Usernaut implements several Kubernetes resource controllers configured by Usernaut CRDs.

## Ramp Up

Before diving into the development of Usernaut, it's essential to understand the underlying technologies and frameworks that power it. Usernaut is built on top of Kubernetes and utilizes Custom Resource Definitions (CRDs) to extend Kubernetes capabilities.
Usernaut is built using the Operator SDK, which simplifies the process of building Kubernetes Operators. It provides a framework for managing the lifecycle of applications and services on Kubernetes clusters.
To get started with developing for Usernaut, you should familiarize yourself with the following concepts and tools:

- [Kubernetes Basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/)
- [Custom Resource Definitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
- [Operator SDK](https://sdk.operatorframework.io/docs/)
- [Operator SDK Go Documentation](https://sdk.operatorframework.io/docs/building-operators/golang/)

### Prerequisites

- Go version v1.24.2+
- Operator SDK version v1.39.2+
- Git
- Docker/Podman
- Access to a Kubernetes cluster (e.g., Minikube, Kind, or a cloud provider)
  - Minikube: <https://minikube.sigs.k8s.io/docs/start>
  - KinD: <https://kind.sigs.k8s.io/docs/user/quick-start/#installation>

### Development Environment Setup

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/redhat-data-and-ai/usernaut.git
   ```

1. **Navigate to the Project Directory**:

   ```bash
   cd usernaut
   ```

1. **Install Dependencies**:

   Ensure you have the necessary dependencies installed. You can use Go modules to manage dependencies:

   ```bash
   go mod tidy
   ```

1. **Install the CRDs**:

   Usernaut uses Custom Resource Definitions (CRDs) to define the resources it manages. You can install the CRDs using the following command:

   ```bash
   make install
   ```

   This will apply the CRD manifests to your Kubernetes cluster.

1. **Build the Operator**:
   You can build the Usernaut operator binary using the following command:

   ```bash
   make build
   ```

1. **Setup the AppConfig**:
   Before running the operator, you need to setup the application configuration. This includes setting up the necessary key secrets, LDAP configurations, and any other required settings. You can find the default configuration file in [`appconfig`](./appconfig/default.yaml).

   **Note**:

   1. Create your own appconfig file with name `local.yaml` in the `appconfig` directory as it won't be committed to the repository.
   1. Update the `local.yaml` file with your specific configurations, such as LDAP server details, SaaS platform credentials, and any other necessary settings.

1. **Run the Operator Locally**:
   To run the operator locally, you can use the following command:

   ```bash
   make run
   ```

   This will start the operator in your local environment, allowing you to test changes in real-time.

1. **Deploy the Operator to a Kubernetes Cluster**:
   To deploy the operator to a Kubernetes cluster, you can use the following command:

   ```bash
   make deploy
   ```

   This will apply the necessary manifests to your cluster and start the operator.

1. **Test the Operator**:
   You can test the operator by creating custom resources (CRs) that the operator manages. For example, you can create a `Group` or `User` resource to see how the operator responds to changes.

   ```bash
   kubectl apply -f config/samples/v1alpha1_group.yaml
   ```

   If the operator is deployed as a kubernetes deployment, monitor the operator logs to see how it processes the CRs:

   ```bash
   kubectl logs -f deployment/usernaut-controller-manager
   ```

## Development Guidelines

To maintain a high standard of code quality and consistency, we follow these development guidelines:

- **Code Style**: Follow the Go code style guidelines. Use `gofmt` to format your code.
- **Testing**: Write unit tests for new features and ensure existing tests pass. Use `go test` to run tests.
- **Documentation**: Document your code and provide clear comments for complex logic. Update the README and other documentation files as needed.
- **Commit Messages**: Use clear and descriptive commit messages. Follow the format: `type(scope): subject`, where `type` is one of `feat`, `fix`, `docs`, `style`, `refactor`, `test`, or `chore`.
- **Pull Requests**: Ensure your pull requests are small and focused on a single change. Provide a clear description of the changes and why they are needed.

## Testing

To ensure the reliability and stability of Usernaut, we have a comprehensive testing strategy that includes unit tests, integration tests, and end-to-end tests. Here are the steps to run the tests:

1. **Run Unit Tests**: Use the following command to run unit tests:

   ```bash
   make test
   ```

<!-- 2. **Run Integration Tests**: To run integration tests, you may need to set up a test environment. Use the following command:

   ```bash
   make integration-test
   ```

3. **Run End-to-End Tests**: For end-to-end tests, ensure you have a running Kubernetes cluster and use the following command:

   ```bash
   make e2e-test
   ``` -->

## Debugging

When debugging Usernaut, you can use the following techniques:

- **Log Statements**: Add log statements in your code to trace the flow of execution and identify issues. Use the `log` package for logging.
- **Debugging Tools**: Use debugging tools like `dlv` to step through your code and inspect variables.

- **Kubernetes Debugging**: Use Kubernetes commands to inspect the state of resources. For example, you can use `kubectl describe` to get detailed information about a resource:

  ```bash
  kubectl describe group <group-name>
  ```

- **Error Handling**: Ensure proper error handling in your code. Use `errors.New` to provide context for errors and make them easier to debug.

## Best Practices

To ensure the maintainability and scalability of Usernaut, we follow these best practices:

- **Modular Code Structure**: Organize your code into packages based on functionality. This makes it easier to navigate and maintain the codebase.
- **Consistent Naming Conventions**: Use consistent naming conventions for variables, functions, and types. This improves code readability and understanding.
- **Avoid Global State**: Minimize the use of global variables and state. Instead, pass dependencies explicitly to functions and methods.
- **Use Context**: Use the `context` package to manage request-scoped values and cancellation signals. This is especially important for long-running operations in Kubernetes controllers.
- **Follow Kubernetes Best Practices**: Adhere to Kubernetes best practices for resource management, such as using labels and annotations for resources, and ensuring proper RBAC configurations.

## Troubleshooting Common Issues

If you encounter issues while developing or running Usernaut, here are some common troubleshooting steps:

- **CRD Not Found**: If you see errors related to CRDs not being found, ensure that the CRDs are applied to your cluster. You can apply them using:

  ```bash
  make install
  ```

- **Operator Not Responding**: If the operator is not responding to changes in CRs, check the operator logs for any errors. You can view the logs using:

  ```bash
  kubectl logs -f deployment/usernaut-controller-manager
  ```

- **Resource Conflicts**: If you encounter resource conflicts, ensure that the resources you are trying to create or update do not already exist in the cluster. Use `kubectl get` to check the current state of resources.
- **Network Issues**: If you have network-related issues, ensure that your Kubernetes cluster is properly configured and that the operator has access to the necessary APIs of the SaaS platforms it manages.
- **Dependency Issues**: If you encounter issues with dependencies, ensure that your Go modules are up to date. You can run:

  ```bash
  go mod tidy
  ```

- **Check Logs**: Always check the logs of the operator for any errors or warnings. You can use `kubectl logs` to view the logs of the operator pod.
- **Validate CRDs**: Ensure that the Custom Resource Definitions (CRDs) are correctly applied to your cluster. You can check the status of CRDs using `kubectl get crd`.
- **Rebuild and Redeploy**: If you make changes to the operator code, ensure you rebuild the binary and redeploy the operator to see the changes take effect.
- **Consult Documentation**: Refer to the [Operator SDK documentation](https://sdk.operatorframework.io/docs/) for guidance on common issues and best practices.
