# Behind reverse proxy?

## Traefik2

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.services.neko-frontend.loadbalancer.server.port=8080"
  - "traefik.http.routers.neko.rule=${TRAEFIK_RULE}"
  - "traefik.http.routers.neko.entrypoints=${TRAEFIK_ENTRYPOINTS}"
  - "traefik.http.routers.neko.tls.certresolver=${TRAEFIK_CERTRESOLVER}"
```

(by @m1k1o, [example](https://github.com/m1k1o/neko-vpn/blob/a1b934515dcf597992a515d61d307c2450a11002/docker-compose.yml#L38-L43))

## Nginx

```conf
server {
  listen 443 ssl http2;
  server_name example.com;

  location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 86400;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Port $server_port;
    proxy_set_header X-Forwarded-Protocol $scheme;
  }
}
```

(by @GigaFyde, [source](https://github.com/nurdism/neko/issues/111#issuecomment-742656957))

## Apache

After successfully installing and running neko, you might want to get rid of the port in the url, use DNS instead of IP address and also having SSL.
This will remove the port from the URL and also enables HTTPS.
To do this, you have to get running Apache server. Now you can go into the `/etc/apache2/sites-available` folder and create new config file for example `neko.conf`
After creating new config file, you can use this example config and paste it in. Some things may vary on your machine so read through and modify if needed.
Bear in mind that your neko server doesn't have to run on the same computer as Apache. They just have to be on the same network, and then you replace localhost with correct internal IP.

```xml
<VirtualHost *:80>
        # The ServerName directive sets the request scheme, hostname and port that
        # the server uses to identify itself. This is used when creating
        # redirection URLs. In the context of virtual hosts, the ServerName
        # specifies what hostname must appear in the request's Host: header to
        # match this virtual host. For the default virtual host (this file) this
        # value is not decisive as it is used as a last resort host regardless.
        # However, you must set it for any further virtual host explicitly.

        # Paths of those modules might vary across different distros.
        LoadModule proxy_module /usr/lib/apache2/modules/mod_proxy.so
        LoadModule proxy_http_module /usr/lib/apache2/modules/mod_proxy_http.so
        LoadModule proxy_wstunnel_module /usr/lib/apache2/modules/mod_proxy_wstunnel.so

        ServerName example.com
        ServerAlias www.example.com

        ProxyRequests Off
        ProxyPass / http://localhost:8080/
        ProxyPassReverse / http://localhost:8080/

        RewriteEngine on
        RewriteCond %{HTTP:Upgrade} websocket [NC]
        RewriteCond %{HTTP:Connection} upgrade [NC]
        RewriteRule /ws(.*) "ws://localhost:8080/ws$1" [P,L]

        # Available loglevels: trace8, ..., trace1, debug, info, notice, warn,
        # error, crit, alert, emerg.
        # It is also possible to configure the loglevel for particular
        # modules, e.g.
        #LogLevel info ssl:warn

        ErrorLog ${APACHE_LOG_DIR}/error.log
        CustomLog ${APACHE_LOG_DIR}/access.log combined

        # For most configuration files from conf-available/, which are
        # enabled or disabled at a global level, it is possible to
        # include a line for only one particular virtual host. For example the
        # following line enables the CGI configuration for this host only
        # after it has been globally disabled with "a2disconf".
        #Include conf-available/serve-cgi-bin.conf
</VirtualHost>
```

(by @DarkReaper231, [source](https://github.com/nurdism/neko/blob/cad98a62a5bd7f1daf2c11980631bb14ba81a1f6/docs/apache-proxypass-config.md#example-apache-config))

After creating your new config file, just use `sudo a2ensite neko.conf` and then `sudo systemctl reload apache2`

### Enabling SSL

If you want to use SSL for your apache configuration, you can install certbot and use it with `sudo certbot`
Then you can just select both `example.com` and `www.example.com` and apply. This will copy your `neko.conf` file and creates one for SSL.

## Caddy

```conf
https://example.com {
  reverse_proxy localhost:8080 {
    header_up Host {host}
    header_up X-Real-IP {remote_host}
    header_up X-Forwarded-For {remote_host}
    header_up X-Forwarded-Proto {scheme}
  }
}
```

(by @ccallahan, [source](https://github.com/nurdism/neko/pull/125/commits/eb4ceda75423b0d960c8aea0240acf6d7a10fef4))
