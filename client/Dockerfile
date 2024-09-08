FROM node:18-bullseye-slim as client
WORKDIR /src

#
# install dependencies
COPY package*.json ./
RUN npm install

#
# build client
COPY . .
RUN npm run build
