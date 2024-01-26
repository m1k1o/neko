FROM node:18-buster-slim

COPY . /app

WORKDIR /app

RUN npm i

EXPOSE 8080

CMD [ "npm", "run", "serve" ]
