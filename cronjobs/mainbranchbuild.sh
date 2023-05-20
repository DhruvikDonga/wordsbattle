#!/bin/bash

# Replace the following variables with your own values
GITHUB_USERNAME="your_github_username"
GITHUB_TOKEN="your_github_access_token"
REPO_NAME="your_repo_name"
DOCKER_COMPOSE_FILE="path/to/your/docker-compose-file.yml"

# Set the directory where the script is located
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Change to the directory where the script is located
cd "${DIR}"

# Check if the local repository is behind the remote repository
git remote update
if [[ $(git status -uno | grep 'Your branch is behind') ]]; then
  # Pull the latest changes from the main branch
  git pull origin main

  # Run Docker Compose
  docker-compose -f ${DOCKER_COMPOSE_FILE} up -d
fi
