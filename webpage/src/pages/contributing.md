# Contributing

Neko is an open-source project, and we welcome contributions from the community. Whether you're a developer, designer, or user, you can help improve Neko by reporting bugs, suggesting new features, or submitting code changes.

## Reporting issues

If you encounter a bug or have a feature request, please open a new issue on the [GitHub repository](https://github.com/m1k1o/neko/issues). Before opening an issue, please check if a similar issue has already been reported.

When reporting an issue, please provide as much information as possible, including:

- A detailed description of the problem
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Screenshots or error messages (if applicable)
- Your operating system and browser version

## Contributing code

If you're a developer and want to contribute code to Neko, follow these steps:

1. **Fork the [project](https://github.com/m1k1o/neko)**: Create a personal copy of the repository by [forking it](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/fork-a-repo) to your GitHub account.

2. **Navigate to [.docker/README.md](https://github.com/m1k1o/neko/tree/master/.docker)**: Follow the instructions in the `.docker/README.md` file for setting up the Docker environment required for development.

3. **Edit files in your branch**: Make your changes in a new branch created from the `master` branch. Ensure your changes are as well documented and tested.

4. **Submit a [pull request](https://github.com/m1k1o/neko/pulls)**: Once your changes are ready, submit a pull request with a detailed explanation of the improvements and any relevant information for the reviewers.

## Server build dependencies

To compile the Golang code locally, you need to install the following dependencies:

```shell
apt-get install -y --no-install-recommends libx11-dev libxrandr-dev libxtst-dev libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev
```

### Retrieving Libclipboard files

Libclipboard files can be retrieved from the `neko_dev_server` Docker container. Run the following commands to copy the necessary files:

```shell
mkdir -p /usr/local/lib/pkgconfig/ /usr/local/include/
docker cp neko_dev_server:/usr/local/lib/libclipboard.a /usr/local/lib/
docker cp neko_dev_server:/usr/local/lib/pkgconfig/libclipboard.pc /usr/local/lib/pkgconfig/
docker cp neko_dev_server:/usr/local/include/libclipboard-config.h /usr/local/include/
docker cp neko_dev_server:/usr/local/include/libclipboard.h /usr/local/include/
```
