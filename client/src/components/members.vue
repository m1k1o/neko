<template>
  <div class="members">
    <div class="members-container">
      <ul class="members-list">
        <li v-if="member">
          <div :class="[{ host: member.id === host }, 'self', 'member']">
            <neko-avatar class="avatar" :seed="member.displayname" :avatar="member.avatar" :size="50" />
            <div v-if="bubble(member.id)" class="member-bubble">{{ bubble(member.id) }}</div>
          </div>
        </li>
        <template v-for="(member, index) in members">
          <li
            v-if="member.id !== id && member.connected"
            :key="index"
            v-tooltip="{ content: member.displayname, placement: 'bottom', offset: 0, boundariesElement: 'body' }"
          >
            <div
              :class="[{ host: member.id === host, admin: member.admin }, 'member']"
              @contextmenu.stop.prevent="onContext($event, { member })"
            >
              <neko-avatar class="avatar" :seed="member.displayname" :avatar="member.avatar" :size="50" />
              <div v-if="bubble(member.id)" class="member-bubble">{{ bubble(member.id) }}</div>
            </div>
          </li>
        </template>
      </ul>
    </div>
    <neko-context ref="context" />
  </div>
</template>

<style lang="scss" scoped>
  .members {
    flex: 0 0 auto;
    overflow: visible;
    padding: 0 0 6px;
    min-height: $party-member-rail-height;
    display: flex;

    .members-container {
      width: 100%;
      min-width: 0;
      max-width: 100%;
      padding: 0 8px;
      margin: 0 auto;

      .members-list {
        display: flex;
        flex-wrap: wrap;
        justify-content: center;
        align-items: center;
        gap: 2px 4px;

        li {
          display: block;
          flex: 0 0 auto;

          .member {
            position: relative;
            display: grid;
            place-items: center;
            width: 56px;
            height: 56px;
            margin: 4px 5px 0;
            border: 2px solid transparent;
            border-radius: 16px;
            transition: transform 0.18s ease, border-color 0.18s ease, background 0.18s ease;

            &:hover {
              transform: translateY(-2px);
              border-color: rgba($style-primary, 0.44);
              background: rgba($style-primary, 0.1);
            }

            &.self {
              &::before {
                font-family: 'Font Awesome 6 Free';
                font-weight: 900;
                content: '\f2bd';
                background: $background-floating;
                color: $style-primary;
                position: absolute;
                top: -3px;
                right: -3px;
                width: 15px;
                height: 15px;
                line-height: 15px;
                font-size: 20px;
                text-align: center;
                border-radius: 50%;
              }
            }

            &.admin {
              &::before {
                display: block;
                font-family: 'Font Awesome 6 Free';
                font-weight: 900;
                content: '\f3ed';
                color: $style-primary;
                background: transparent;
                position: absolute;
                top: -2px;
                right: -2px;
                width: 14px;
                height: 14px;
                font-size: 14px;
                text-align: center;
              }
            }

            &.host::after {
              display: block;
              font-family: 'Font Awesome 6 Free';
              font-weight: 900;
              content: '\f521';
              background: $style-primary;
              color: $background-floating;
              position: absolute;
              left: -3px;
              bottom: -3px;
              width: 20px;
              height: 20px;
              line-height: 20px;
              font-size: 10px;
              text-align: center;
              border-radius: 50%;
            }

            .avatar {
              border-radius: 50%;
              overflow: hidden;
              width: 100%;
              border: 2px solid rgba($background-tertiary, 0.9);
            }

            .member-bubble {
              position: absolute;
              z-index: 5;
              left: 50%;
              bottom: calc(100% + 8px);
              width: max-content;
              max-width: 180px;
              padding: 7px 10px;
              border: 1px solid rgba($text-normal, 0.14);
              border-radius: 10px 10px 10px 3px;
              background: $background-floating;
              color: $text-normal;
              box-shadow: 0 8px 24px rgba(0, 0, 0, 0.28);
              font-size: 12px;
              line-height: 16px;
              white-space: normal;
              overflow: hidden;
              text-overflow: ellipsis;
              animation: avatar-message-in 0.2s ease-out forwards, avatar-message-out 1.1s ease-in 5.4s forwards;
              pointer-events: none;
            }
          }

          &:nth-child(2) {
            margin-left: 20px;

            &::before {
              position: absolute;
              content: ' ';
              height: 45px;
              width: 2px;
              background: $background-secondary;
              margin-top: 13px;
              margin-left: -9px;
            }
          }
        }
      }
    }
  }

  @keyframes avatar-message-in {
    from {
      opacity: 0;
      transform: translate(-50%, 5px) scale(0.96);
    }
    to {
      opacity: 1;
      transform: translate(-50%, 0) scale(1);
    }
  }

  @keyframes avatar-message-out {
    from {
      opacity: 1;
    }
    to {
      opacity: 0;
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Vue } from 'vue-property-decorator'

  import Content from './context.vue'
  import Avatar from './avatar.vue'

  @Component({
    name: 'neko-members',
    components: {
      'neko-context': Content,
      'neko-avatar': Avatar,
    },
  })
  export default class extends Vue {
    @Ref('context') readonly _context!: any

    get id() {
      return this.$accessor.user.id
    }

    get host() {
      return this.$accessor.remote.id
    }

    get member() {
      return this.$accessor.user.member
    }

    get members() {
      return this.$accessor.user.members
    }

    bubble(id: string) {
      return this.$accessor.chat.bubbles[id]
    }

    onContext(event: MouseEvent, data: any) {
      this._context.open(event, data)
    }
  }
</script>
