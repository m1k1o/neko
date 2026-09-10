FROM mcr.microsoft.com/playwright:v1.48.2-jammy

WORKDIR /opt/neko-e2e
COPY package.json ./
RUN npm install --omit=dev --no-audit --no-fund

ENTRYPOINT ["node", "/tests/playwright.mjs"]
