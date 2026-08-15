FROM node:24.19-alpine3.24

WORKDIR /app

COPY package.json .     
# COPY package*.json ./     If package-lock.json exists then

RUN npm install
# npm ci        If there's a package-lock.json

COPY . .

EXPOSE 5000

CMD ["npm", "start"]