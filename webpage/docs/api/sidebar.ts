import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "api/neko-api",
    },
    {
      type: "category",
      label: "General",
      link: {
        type: "doc",
        id: "api/general",
      },
      collapsed: false,
      items: [
        {
          type: "doc",
          id: "api/healthcheck",
          label: "Health Check",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/metrics",
          label: "Metrics",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/batch",
          label: "Batch Request",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/stats",
          label: "Get Stats",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Current Session",
      link: {
        type: "doc",
        id: "api/current-session",
      },
      collapsed: false,
      items: [
        {
          type: "doc",
          id: "api/login",
          label: "User Login",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/logout",
          label: "User Logout",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/whoami",
          label: "Get Current User",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/profile",
          label: "Update Profile",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Sessions",
      link: {
        type: "doc",
        id: "api/sessions",
      },
      collapsed: false,
      items: [
        {
          type: "doc",
          id: "api/sessions-get",
          label: "List Sessions",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/session-get",
          label: "Get Session",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/session-remove",
          label: "Remove Session",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/session-disconnect",
          label: "Disconnect Session",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Members",
      link: {
        type: "doc",
        id: "api/members",
      },
      collapsed: false,
      items: [
        {
          type: "doc",
          id: "api/schemas/memberprofile",
          label: "MemberProfile",
          className: "schema",
        },
        {
          type: "doc",
          id: "api/members-list",
          label: "List Members",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/members-create",
          label: "Create Member",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/members-get-profile",
          label: "Get Member Profile",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/members-update-profile",
          label: "Update Member Profile",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/members-remove",
          label: "Remove Member",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/members-update-password",
          label: "Update Member Password",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/members-bulk-update",
          label: "Bulk Update Members",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/members-bulk-delete",
          label: "Bulk Delete Members",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Room Settings",
      link: {
        type: "doc",
        id: "api/room-settings",
      },
      collapsed: false,
      items: [
        {
          type: "doc",
          id: "api/settings-get",
          label: "Get Room Settings",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/settings-set",
          label: "Update Room Settings",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Room Broadcast",
      link: {
        type: "doc",
        id: "api/room-broadcast",
      },
      collapsed: false,
      items: [
        {
          type: "doc",
          id: "api/broadcast-status",
          label: "Get Broadcast Status",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/broadcast-start",
          label: "Start Broadcast",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/broadcast-stop",
          label: "Stop Broadcast",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Room Clipboard",
      link: {
        type: "doc",
        id: "api/room-clipboard",
      },
      collapsed: false,
      items: [
        {
          type: "doc",
          id: "api/clipboard-get-text",
          label: "Get Clipboard Content",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/clipboard-set-text",
          label: "Set Clipboard Content",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/clipboard-get-image",
          label: "Get Clipboard Image",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Room Keyboard",
      link: {
        type: "doc",
        id: "api/room-keyboard",
      },
      collapsed: false,
      items: [
        {
          type: "doc",
          id: "api/keyboard-map-get",
          label: "Get Keyboard Map",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/keyboard-map-set",
          label: "Set Keyboard Map",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/keyboard-modifiers-get",
          label: "Get Keyboard Modifiers",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/keyboard-modifiers-set",
          label: "Set Keyboard Modifiers",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Room Control",
      link: {
        type: "doc",
        id: "api/room-control",
      },
      collapsed: false,
      items: [
        {
          type: "doc",
          id: "api/control-status",
          label: "Get Control Status",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/control-request",
          label: "Request Control",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/control-release",
          label: "Release Control",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/control-take",
          label: "Take Control",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/control-give",
          label: "Give Control",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/control-reset",
          label: "Reset Control",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Room Screen",
      link: {
        type: "doc",
        id: "api/room-screen",
      },
      collapsed: false,
      items: [
        {
          type: "doc",
          id: "api/screen-configuration",
          label: "Get Screen Configuration",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/screen-configuration-change",
          label: "Change Screen Configuration",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/screen-configurations-list",
          label: "Get List of Screen Configurations",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/screen-cast-image",
          label: "Get Screencast Image",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/screen-shot-image",
          label: "Get Screenshot Image",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Room Upload",
      link: {
        type: "doc",
        id: "api/room-upload",
      },
      collapsed: false,
      items: [
        {
          type: "doc",
          id: "api/upload-drop",
          label: "Upload and Drop File",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/upload-dialog",
          label: "Upload File to Dialog",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/upload-dialog-close",
          label: "Close File Chooser Dialog",
          className: "api-method delete",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
