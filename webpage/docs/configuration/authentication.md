---
description: Configuration related to the Authentication and Sessions in Neko.
---

import { Def, Opt } from '@site/src/components/Anchor';
import { ConfigurationTab } from '@site/src/components/Configuration';
import configOptions from './help.json';

# Authentication

Authentication is split into two modules:

- **[Member Provider](#member)** - handles authentication and authorization of users, can be used to authenticate users against a database, LDAP, or any other system.
  :::warning NOTE
  LDAP, OIDC, and other subsystems are _not_ currently implemented. 
  If you are interested in these features, please consider contributing or sponsoring their development.
  :::

- **[Session Provider](#session)** - handles session management, after the module authenticates the user, it creates a session and handles the session lifecycle.

## Member Profile {#profile}

A member profile is a structure that describes the user and what the user is allowed to do in the system.

| Field                      | Description | Type |
|----------------------------|-------------|------|
| <Def id="profile.name" />                     | User's name as shown in the UI, must not be unique within the system (not used as an identifier). | string |
| <Def id="profile.is_admin" />                 | Whether the user can perform administrative tasks that include managing users, sessions, and settings. | boolean |
| <Def id="profile.can_login" />                | Whether the user can log in to the system and use the HTTP API. | boolean |
| <Def id="profile.can_connect" />              | Whether the user can connect to the room using the WebSocket API (needs <Opt id="profile.can_login" /> to be enabled). | boolean |
| <Def id="profile.can_watch" />                | Whether the user can connect to the WebRTC stream and watch the room's audio and video (needs <Opt id="profile.can_connect" /> to be enabled). | boolean |
| <Def id="profile.can_host" />                 | Whether the user can grab control of the room and control the mouse and keyboard. | boolean |
| <Def id="profile.can_share_media" />          | Whether the user can share their webcam and microphone with the room. | boolean |
| <Def id="profile.can_access_clipboard" />     | Whether the user can read and write to the room's clipboard. | boolean |
| <Def id="profile.sends_inactive_cursor" />    | Whether the user sends the cursor position even when the user is not hosting the room, this is used to show the cursor of the user to other users. | boolean |
| <Def id="profile.can_see_inactive_cursors" /> | Whether the user can see the cursor of other users even when they are not hosting the room. | boolean |
| <Def id="profile.plugins" />                  | A map of plugin names and their configuration, plugins can use this to store user-specific settings, see the [Plugins Configuration](/docs/v3/configuration/plugins) for more information. | object |

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

## Member Providers {#member}

Member providers are responsible for deciding whether given credentials are valid or not. This validation can either be done against a local database or an external system.

:::info
Currently, Neko supports configuring only one authentication provider at a time. This means you must choose a single provider that best fits your deployment needs.
:::

### Multi-User Provider {#member.multiuser}

This is the **default provider** that works exactly like the authentication used to work in v2 of neko.

This provider allows you to define two types of users: **regular** users and **admins**. Which user is an admin is determined by the password they provide when logging in. If the password is correct, the user is an admin; otherwise, they are a regular user. Based on those profiles, the users are generated on demand when they log in and they are removed when they log out. Their username is prefixed with 5 random characters to avoid conflicts when multiple users share the same username.

Profiles for regular users and admins are optional, if not provided, the default profiles are used (see below in the example configuration).

<ConfigurationTab options={{
  "member.provider": 'multiuser',
  "member.multiuser.admin_password": {
    defaultValue: "admin",
    description: "Password for admins, in plain text.",
  },
  "member.multiuser.admin_profile": {
    defaultValue: {},
    description: "Profile fields as described above",
  },
  "member.multiuser.user_password": {
    defaultValue: "neko",
    description: "Password for regular users, in plain text.",
  },
  "member.multiuser.user_profile": {
    defaultValue: {},
    description: "Profile fields as described above",
  },
}} />

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
  NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: "admin"
  NEKO_MEMBER_MULTIUSER_USER_PASSWORD: "neko"
```
:::

### File Provider {#member.file}

This provider reads the user's credentials from a file. It is useful for small deployments where you don't want to set up a database or LDAP server and still want to have persistent users.

<ConfigurationTab options={{
  "member.provider": 'file',
  "member.file.path": {
    defaultValue: "/opt/neko/members.json",
    description: "Absolute path to the file containing the users and their passwords.",
  },
  "member.file.hash": {
    defaultValue: false,
    description: "Whether the passwords are hashed using sha256 or not.",
  },
}} />

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

  Please note that the passwords are stored in plain text. To store them securely, set the `hash` field to `true` in the configuration. After that, the passwords are expected to be hashed using sha256 and base64-encoded. The file will look like this:

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

  If you want to hash the passwords, you can use the following command to generate a sha256 base64-encoded hash of the password:

  ```bash
  echo -n "password" | openssl sha256 -binary | base64 -
  ```  
</details>

### Object Provider {#member.object}

This provider is the same as the file provider, but it saves the users only in memory. That means that the users are lost when the server is restarted. However, the default users can be set in the configuration file. The difference from the multi-user provider is that the users are not generated on demand and we define exactly which users with their passwords and profiles are allowed to log in. They cannot be logged in twice with the same username.

<ConfigurationTab options={{
  "member.provider": 'object',
  "member.object.users": {
    defaultValue: [],
    description: "List of users with their passwords and profiles.",
  },
}} />

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

### No-Auth Provider {#member.noauth}

This provider allows any user to log in without any authentication. It is useful for testing and development purposes.

<ConfigurationTab options={configOptions} filter={{
  "member.provider": 'noauth',
}} comments={false} />

:::danger
Do not use this provider in production environments unless you know exactly what you are doing. It allows anyone to log in and control neko as an admin.
:::

## Session Provider {#session}

Currently, there are only two providers available for sessions: **memory** and **file**.

Simply by specifying the `session.file` to a file path, the session provider will store the sessions in a file. Otherwise, the sessions are stored in memory and are lost when the server is restarted.

<ConfigurationTab options={configOptions} filter={{
  "session.file": '/opt/neko/sessions.json',
}} comments={false} />

:::info
In the future, we plan to add more session providers, such as Redis, PostgreSQL, etc. So the Configuration Options may change.
:::

## API User {#api_token}

The API User is a special user that is used to authenticate the HTTP API requests. It cannot connect to the room, but it can perform administrative tasks. The API User does not have a password but only a token that is used to authenticate the requests. If the token is not set, the API User is disabled.

<ConfigurationTab options={configOptions} filter={{
  "session.api_token": '<secret_token>',
}} comments={false} />

:::tip
This user is useful in some situations when the rooms are generated by the server and the token is guaranteed to be random every time a short-lived room is run. It is not a good idea to define this token for long-lived rooms, as it can be stolen and used to perform administrative tasks.

You can generate a random token using the following command:

```bash
openssl rand -hex 32
```
:::

## Cookies {#session.cookie}

The authentication between the client and the server can be done using cookies or the `Authorization` header. The cookies are used by default, but you can disable them by setting the <Opt id="session.cookie.enabled" /> to `false`.

:::warning
If you disable the cookies, the token will be sent to the client in the login response and saved in local storage. This is less secure than using cookies, as the token **can be stolen using XSS attacks**. Therefore, it is recommended to use cookies.
:::

<ConfigurationTab options={configOptions} filter={[
  'session.cookie.enabled',
  'session.cookie.name',
  'session.cookie.expiration',
  'session.cookie.secure',
  'session.cookie.http_only',
  'session.cookie.domain',
  'session.cookie.path'
]} comments={false} />

- <Def id="session.cookie.enabled" /> - Whether the cookies are enabled or not.
- <Def id="session.cookie.name" /> - Name of the cookie used to store the session.
- <Def id="session.cookie.expiration" /> - Expiration time of the cookie, use [go duration format](https://pkg.go.dev/time#ParseDuration) (e.g., `24h`, `1h30m`, `60m`).
- <Def id="session.cookie.secure" /> and <Def id="session.cookie.http_only" /> - Ensures that the cookie is only sent over HTTPS and cannot be accessed by JavaScript, see [MDN Web Docs](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#block_access_to_your_cookies) for more information.
- <Def id="session.cookie.domain" /> and <Def id="session.cookie.path" /> - Define where the cookie is valid, see [MDN Web Docs](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#define_where_cookies_are_sent) for more information.

:::info
The <Opt id="session.cookie.secure" /> and <Opt id="session.cookie.http_only" /> are set to `true` by default, which means that the cookie is only sent over HTTPS. If you are using HTTP, you should really consider using HTTPS. Only for testing and development purposes should you consider setting it to `false`.
:::
