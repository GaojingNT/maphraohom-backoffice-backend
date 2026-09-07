#!/usr/bin/env bash
# One-time VM bootstrap for maphraohom-backoffice-backend.
# Target: Ubuntu/Debian VPS. Run as a sudo-capable user, not root directly.
#
# Usage:
#   chmod +x setup-vm.sh
#   ./setup-vm.sh
set -euo pipefail

DEPLOY_PATH="/opt/maphraohom-backoffice-backend"
COMPOSE_FILE="docker-compose.prod.yml"

echo "==> Installing Docker Engine + Compose plugin"
if ! command -v docker &>/dev/null; then
  curl -fsSL https://get.docker.com | sudo sh
  sudo usermod -aG docker "$USER"
  echo "Added $USER to the docker group. Log out/in (or run 'newgrp docker') before continuing."
fi

echo "==> Creating deploy directory at ${DEPLOY_PATH}"
sudo mkdir -p "${DEPLOY_PATH}"
sudo chown "$USER":"$USER" "${DEPLOY_PATH}"

echo "==> Opening firewall for the API port (skip if you front this with a reverse proxy on 80/443 instead)"
if command -v ufw &>/dev/null; then
  sudo ufw allow 8000/tcp || true
fi

cat <<'EOF'

==> Manual steps left:

1. Copy docker-compose.prod.yml into ${DEPLOY_PATH} on this VM (scp from your machine, or
   'git clone' the repo here and copy the file out):
     scp docker-compose.prod.yml <user>@<vm-host>:${DEPLOY_PATH}/

2. Create ${DEPLOY_PATH}/.env with REAL production values (copy .env.example as a starting
   point, then change every password/secret). This file never goes through GitHub — it stays
   only on this VM.

3. Since the GitHub repo/image is private, authenticate this VM to pull from GHCR once:
     echo <YOUR_GITHUB_PAT_with_read:packages> | docker login ghcr.io -u <your-github-username> --password-stdin

4. First run:
     cd ${DEPLOY_PATH}
     docker compose -f docker-compose.prod.yml up -d

After that, the GitHub Actions "Deploy" workflow will pull the new image and restart the
app service automatically on every push to main.
EOF
