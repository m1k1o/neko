ARG BASE_IMAGE=node:24-trixie-slim
FROM $BASE_IMAGE AS client

WORKDIR /src

#
# install dependencies (the e2e browser is not needed to build)
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
COPY package*.json ./
RUN npm ci

#
# build client
COPY . .
RUN npm run build

#
# artifacts from this stage
# COPY --from=client /src/dist/ /var/www
