# Reverse Proxy Setup

If you want to run Neko behind a reverse proxy, you can use the following examples to configure your server.

:::tip
Do not forget to enable [`server.proxy=true`](/docs/v3/configuration#server.proxy) in your `config.yml` file to allow the server to trust the proxy headers.
:::

Neko pings websocket client every 10 seconds, and client is scheduled to send [heartbeat](/docs/v3/configuration#session.heartbeat_interval) to the server every 120 seconds. Make sure, that your timeout settings in the reverse proxy are set accordingly.

## Traefik v2 {#traefik-v2}

See the example below for a `docker-compose.yml` file.

```yaml title="docker-compose.yml"
labels:
  - "traefik.enable=true"
  - "traefik.http.services.neko-frontend.loadbalancer.server.port=8080"
  - "traefik.http.routers.neko.rule=${TRAEFIK_RULE}"
  - "traefik.http.routers.neko.entrypoints=${TRAEFIK_ENTRYPOINTS}"
  - "traefik.http.routers.neko.tls.certresolver=${TRAEFIK_CERTRESOLVER}"
```

For more information, check out the [official Traefik documentation](https://doc.traefik.io/traefik/v2.0/routing/routers/). For SSL, see the [official Traefik SSL documentation](https://doc.traefik.io/traefik/v2.0/https/acme/).

## Nginx {#nginx}

See the example below for an Nginx configuration file.

```conf title="/etc/nginx/sites-available/neko.conf"
server {
  listen 443 ssl http2;
  server_name example.com;

  location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_cache_bypass $http_upgrade;
  }
}
```

For more information, check out the [official Nginx documentation](https://nginx.org/en/docs/beginners_guide.html). For SSL, see the [official Nginx SSL documentation](https://nginx.org/en/docs/http/configuring_https_servers.html) or use [certbot](https://certbot.eff.org/instructions?ws=nginx&os=ubuntufocal).

## Apache {#apache}

To do this, you need to have a running Apache server. Navigate to the `/etc/apache2/sites-available` folder and create a new configuration file, for example, `neko.conf`.

After creating the new configuration file, you can use the example below and paste it in. Some things may vary on your machine, so read through and modify it if needed.

Bear in mind that your Neko server does not have to run on the same computer as Apache. They just need to be on the same network, and then you replace `localhost` with the correct internal IP.

```xml title="/etc/apache2/sites-available/neko.conf"
<VirtualHost *:80>
  # The ServerName directive sets the request scheme, hostname, and port that
  # the server uses to identify itself. This is used when creating
  # redirection URLs. In the context of virtual hosts, the ServerName
  # specifies what hostname must appear in the request's Host: header to
  # match this virtual host. For the default virtual host (this file), this
  # value is not decisive as it is used as a last resort host regardless.
  # However, you must set it explicitly for any further virtual host.

  # Paths of these modules might vary across different distros.
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

  # Available log levels: trace8, ..., trace1, debug, info, notice, warn,
  # error, crit, alert, emerg.
  # It is also possible to configure the log level for particular
  # modules, e.g.
  #LogLevel info ssl:warn

  ErrorLog ${APACHE_LOG_DIR}/error.log
  CustomLog ${APACHE_LOG_DIR}/access.log combined

  # For most configuration files from conf-available/, which are
  # enabled or disabled at a global level, it is possible to
  # include a line for only one particular virtual host. For example, the
  # following line enables the CGI configuration for this host only
  # after it has been globally disabled with "a2disconf".
  #Include conf-available/serve-cgi-bin.conf
</VirtualHost>
```

After creating your new configuration file, just use `sudo a2ensite neko.conf` and then `sudo systemctl reload apache2`.

See the [official Apache documentation](https://httpd.apache.org/docs/2.4/vhosts/examples.html) for more information. For SSL, see the [official Apache SSL documentation](https://httpd.apache.org/docs/2.4/ssl/ssl_howto.html) or use [certbot](https://certbot.eff.org/instructions?ws=apache&os=snap).

## Caddy {#caddy}

See the example below for a Caddyfile.

```conf title="Caddyfile"
https://example.com {
  reverse_proxy localhost:8080
}
```

For more information, check out the [official Caddy documentation](https://caddyserver.com/docs/caddyfile). For SSL, see the [official Caddy automatic HTTPS documentation](https://caddyserver.com/docs/automatic-https).

## HAProxy {#haproxy}

Using your frontend section *(mine is called http-in)*, add the ACL to redirect correctly to your Neko instance.

```sh title="/etc/haproxy/haproxy.cfg"
frontend http-in
  #/********
  #* NEKO *
  acl neko_rule_http hdr(host) neko.domain.com # Adapt the domain
  use_backend neko_srv if neko_rule_http
  #********/

backend neko_srv
  mode http
  option httpchk
      server neko 172.16.0.0:8080 # Adapt the IP
```

Then, restart the HAProxy service.

```sh
service haproxy restart
```

See the [official HAProxy documentation](https://www.haproxy.com/documentation/haproxy-configuration-manual/1-8r1/) for more information.

<details>

<summary>Troubleshooting HAProxy Issues</summary>

If you're having trouble reaching your HAProxy instance, try the following steps:

1. **Check HAProxy Logs**  
    Verify what HAProxy is reporting by checking its status:
    ```sh
    service haproxy status
    ```

2. **Monitor Logs in Real-Time**  
    If the service is running and the ACL rule and backend configuration seem correct, tail the logs to investigate further:
    ```sh
    tail -f /var/log/haproxy.log
    ```
    Then, access your `neko.instance.com` and observe the logs for any issues.

3. **Adjust Timeout Settings**  
    Ensure the global timeout is set to 60 seconds to prevent premature request failures:
    ```sh
    global
      stats timeout 60s
    ```

4. **Configure Defaults Section**  
    Add the following settings to the `defaults` section to handle timeouts and forward headers properly:
    ```sh
    defaults
      option forwardfor
      timeout connect 30000
      timeout client  65000
      timeout server  65000
    ```

:::note
Don't forget to restart the service each time you modify the `.cfg` file!
:::

</details>
