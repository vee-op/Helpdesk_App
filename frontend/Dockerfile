FROM nginx:1.31-alpine3.24

COPY index.html /usr/share/nginx/html

COPY style.css /usr/share/nginx/html

COPY app.js /usr/share/nginx/html

COPY default.conf.template /etc/nginx/conf.d/default.conf.template

EXPOSE 80