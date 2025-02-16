---
sidebar_position: 1
description: Configuration related to the Authentication and Sessions in Neko.
---

# Authentication

Authentication is split into two modules:

- **[Member Provider](#member-providers)** - handles authentication and authorization of users, can be used to authenticate users against a database, LDAP, or any other system.
- **[Session Provider](#session-provider)** - handles session management, after the module authenticates the user, it creates a session and handles the session lifecycle.

## Member Profile

A member profile is a structure that describes the user and what the user is allowed to do in the system.

| Field                      | Description | Type |
|----------------------------|-------------|------|
| `name`                     | User's name as shown in the UI, must not be unique within the system (not used as an identifier). | string |
| `is_admin`                 | Whether the user can perform administrative tasks that include managing users, sessions, and settings. | boolean |
| `can_login`                | Whether the user can log in to the system and use the HTTP API. | boolean |
| `can_connect`              | Whether the user can connect to the room using the WebSocket API (needs `can_login` to be enabled). | boolean |
| `can_watch`                | Whether the user can connect to the WebRTC stream and watch the room's audio and video (needs `can_connect` to be enabled). | boolean |
| `can_host`                 | Whether the user can grab control of the room and control the mouse and keyboard. | boolean |
| `can_share_media`          | Whether the user can share their webcam and microphone with the room. | boolean |
| `can_access_clipboard`     | Whether the user can read and write to the room's clipboard. | boolean |
| `sends_inactive_cursor`    | Whether the user sends the cursor position even when the user is not hosting the room, this is used to show the cursor of the user to other users. | boolean |
| `can_see_inactive_cursors` | Whether the user can see the cursor of other users even when they are not hosting the room. | boolean |
| `plugins`                  | A map of plugin names and their configuration, plugins can use this to store user-specific settings, see the [Plugins Configuration](/docs/getting-started/configuration/plugins) for more information. | object |

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
  <TabItem value="yaml" label="YAML" default>
  
    ```yaml title="Example member profile in YAML"
    name: User Name
    is_admin: false
    can_login: true
    can_connect: true
    can_watch: true
    can_host: true
    can_share_media: true
    can_access_clipboard: true
    sends_inactive_cursor: true
    can_see_inactive_cursors: true
    plugins:
      <key>: <value>
    ```

  </TabItem>
  <TabItem value="json" label="JSON">

    ```json title="Example member profile in JSON"
    {
      "name": "User Name",
      "is_admin": false,
      "can_login": true, 
      "can_connect": true, 
      "can_watch": true, 
      "can_host": true, 
      "can_share_media": true, 
      "can_access_clipboard": true, 
      "sends_inactive_cursor": true, 
      "can_see_inactive_cursors": true,
      "plugins": {
        "<key>": "<value>"
      }
    }
    ```

  </TabItem>
</Tabs>

## Member Providers

Member providers are responsible for deciding whether given credentials are valid or not. This validation can either be done against a local database or an external system.

### Multi-User Provider

This is the **default provider** that works exactly like the authentication used to work in v2 of neko.

This provider allows you to define two types of users: **regular** users and **admins**. Which user is an admin is determined by the password they provide when logging in. If the password is correct, the user is an admin; otherwise, they are a regular user. Based on those profiles, the users are generated on demand when they log in and they are removed when they log out. Their username is prefixed with 5 random characters to avoid conflicts when multiple users share the same username.

Profiles for regular users and admins are optional, if not provided, the default profiles are used (see below in the example configuration).

```yaml title="config.yaml"
member:
  provider: multiuser
  multiuser:
    # Password for admins, in plain text.
    admin_password: "adminPassword"
    # Profile fields as described above
    admin_profile:
      ...
    # Password for regular users, in plain text.
    user_password: "userPassword"
    # Profile fields as described above
    user_profile:
      ...
```

<details>
  <summary>See example configuration</summary>
  
  The default profiles for regular users and admins are as follows, highlighting the differences between them.

  ```yaml title="config.yaml"
  member:
    provider: multiuser
    multiuser:
      admin_password: "admin"
      admin_profile:
        name: "" # if empty, the login username is used
        # highlight-start
        is_admin: true
        # highlight-end
        can_login: true
        can_connect: true
        can_watch: true
        can_host: true
        can_share_media: true
        can_access_clipboard: true
        sends_inactive_cursor: true
        # highlight-start
        can_see_inactive_cursors: true
        # highlight-end
      user_password: "neko"
      user_profile:
        name: "" # if empty, the login username is used
        # highlight-start
        is_admin: false
        # highlight-end
        can_login: true
        can_connect: true
        can_watch: true
        can_host: true
        can_share_media: true
        can_access_clipboard: true
        sends_inactive_cursor: true
        # highlight-start
        can_see_inactive_cursors: false
        # highlight-end
  ```

</details>

:::tip
For easier configuration, you can specify only passwords using environment variables:

```yaml title="docker-compose.yaml"
environment:
  NEKO_MEMBER_MULTIUSER_USER_PASSWORD: "neko"
  NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: "admin"
```
:::

### File Provider

This provider reads the user's credentials from a file. It is useful for small deployments where you don't want to set up a database or LDAP server and still want to have persistent users.

```yaml title="config.yaml"
member:
  provider: file
  file:
    # Absolute path to the file containing the users and their passwords.
    path: /opt/neko/members.json
    # Whether the passwords are hashed using sha256 or not.
    hash: true
```

It allows you to store the user's credentials in a JSON file. The JSON structure maps user logins to their passwords and profiles.

```json title="members.json"
{
  "<user_login>": {
    "password": "<user_password>",
    "profile": /* Member Profile, as described above */
  }
}
```

You can leave the file empty and add users later using the HTTP API.

<details>
  <summary>See example `members.json` file</summary>

  We have two users, `admin` and `user` with their passwords and profiles. `admin` is a regular user, while `user` is an admin.

  Please note that the passwords are stored in plain text. To store them securely, set the `hash` field to `true` in the configuration. After that, the passwords are expected to be hashed using the bcrypt algorithm.

  ```json title="members.json"
  {
    "admin": {
      "password": "admin",
      "profile": {
        "name": "Administrator",
        "is_admin": true,
        "can_login": true,
        "can_connect": true,
        "can_watch": true,
        "can_host": true,
        "can_share_media": true,
        "can_access_clipboard": true,
        "sends_inactive_cursor": true,
        "can_see_inactive_cursors": true,
        "plugins": {}
      }
    },
    "user": {
      "password": "neko",
      "profile": {
        "name": "User",
        "is_admin": false,
        "can_login": true,
        "can_connect": true,
        "can_watch": true,
        "can_host": true,
        "can_share_media": true,
        "can_access_clipboard": true,
        "sends_inactive_cursor": true,
        "can_see_inactive_cursors": false,
        "plugins": {}
      }
    }
  }
  ```

  If you want to hash the passwords, you can use the following command to generate a sha256 hash:

  ```bash
  echo -n "password" | sha256sum
  ```  
</details>

### Object Provider

This provider is the same as the file provider, but it saves the users only in memory. That means that the users are lost when the server is restarted. However, the default users can be set in the configuration file. The difference from the multi-user provider is that the users are not generated on demand and we define exactly which users with their passwords and profiles are allowed to log in. They cannot be logged in twice with the same username.

```yaml title="config.yaml"
member:
  provider: object
  object:
  # List of users with their passwords and profiles
  - username: "admin"
    # Password in plain text
    password: "admin"
    # Profile fields as described above
    profile:
      ...
```

<details>
  <summary>See example configuration</summary>
  
  We have two users, `admin` and `user` with their passwords and profiles. `admin` is an admin, while `user` is a regular user.

  ```yaml title="config.yaml"
  member:
    provider: object
    object:
      users:
      - username: "admin"
        password: "admin"
        profile:
          name: "Administrator"
          is_admin: true
          can_login: true
          can_connect: true
          can_watch: true
          can_host: true
          can_share_media: true
          can_access_clipboard: true
          sends_inactive_cursor: true
          can_see_inactive_cursors: true
      - username: "user"
        password: "neko"
        profile:
          name: "User"
          is_admin: false
          can_login: true
          can_connect: true
          can_watch: true
          can_host: true
          can_share_media: true
          can_access_clipboard: true
          sends_inactive_cursor: true
          can_see_inactive_cursors: false
  ```

</details>

### No-Auth Provider

This provider allows any user to log in without any authentication. It is useful for testing and development purposes.

```yaml title="config.yaml"
member:
  provider: noauth
```

:::danger
Do not use this provider in production environments unless you know exactly what you are doing. It allows anyone to log in and control neko as an admin.
:::

## Session Provider

Currently, there are only two providers available for sessions: **memory** and **file**.

Simply by specifying the `session.file` to a file path, the session provider will store the sessions in a file. Otherwise, the sessions are stored in memory and are lost when the server is restarted.

```yaml title="config.yaml"
session:
  file: /opt/neko/sessions.json
```

:::info
In the future, we plan to add more session providers, such as Redis, PostgreSQL, etc. So the Configuration Options may change.
:::

## API User

The API User is a special user that is used to authenticate the HTTP API requests. It cannot connect to the room, but it can perform administrative tasks. The API User does not have a password but only a token that is used to authenticate the requests. If the token is not set, the API User is disabled.

```yaml title="config.yaml"
session:
  api_token: "apiToken"
```

:::tip
This user is useful in some situations when the rooms are generated by the server and the token is guaranteed to be random every time a short-lived room is run. It is not a good idea to define this token for long-lived rooms, as it can be stolen and used to perform administrative tasks.

You can generate a random token using the following command:

```bash
openssl rand -hex 32
```
:::

## Cookies

The authentication between the client and the server can be done using cookies or the `Authorization` header. The cookies are used by default, but you can disable them by setting the `session.cookie.enabled` to `false`.

:::warning
If you disable the cookies, the token will be sent to the client in the login response and saved in local storage. This is less secure than using cookies, as the token **can be stolen using XSS attacks**. Therefore, it is recommended to use cookies.
:::

```yaml title="config.yaml"
session:
  cookie:
    # Whether the cookies are enabled or not.
    enabled: true
    # Name of the cookie used to store the session.
    name: "NEKO_SESSION"
    # Expiration time of the cookie in seconds.
    expiration: 86400
    # Whether the cookie is secure (HTTPS only) or not.
    secure: true
```

:::info
The `session.cookie.secure` is set to `true` by default, which means that the cookie is only sent over HTTPS. If you are using HTTP, you should really consider using HTTPS. Only for testing and development purposes should you consider setting it to `false`.
:::
