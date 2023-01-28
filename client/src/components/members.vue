<template>
  <div class="members">
    <div class="members-container">
      <ul class="members-list">
        <li v-if="member">
          <div :class="[{ host: member.id === host }, 'self', 'member']">
            <neko-avatar class="avatar" :seed="member.displayname" :size="50" />
          </div>
        </li>
        <template v-for="(member, index) in members">
          <li
            v-if="member.id !== id && member.connected"
            :key="index"
            v-tooltip="{ content: member.displayname, placement: 'bottom', offset: -15, boundariesElement: 'body' }"
          >
            <div
              :class="[{ host: member.id === host, admin: member.admin }, 'member']"
              @contextmenu.stop.prevent="onContext($event, { member })"
            >
              <neko-avatar class="avatar" :seed="member.displayname" :size="50" />
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
    flex: 1;
    overflow-x: scroll;
    overflow-y: hidden;
    padding-bottom: 14px;
    scrollbar-width: thin;
    scrollbar-color: $background-secondary $background-tertiary;
    min-height: 60px;
    display: flex;

    &::-webkit-scrollbar {
      height: 4px;
    }

    &::-webkit-scrollbar-track {
      background-color: $background-tertiary;
    }

    &::-webkit-scrollbar-thumb {
      background-color: $background-secondary;
      border-radius: 4px;
    }

    &::-webkit-scrollbar-thumb:hover {
      background-color: $background-primary;
    }

    .members-container {
      display: block;
      clear: both;
      padding: 0 20px;
      margin: 0 auto;

      .members-list {
        white-space: nowrap;
        clear: both;

        li {
          display: inline-block;

          .member {
            position: relative;
            display: block;
            width: 50px;
            height: 50px;
            margin: 10px 5px 0 5px;

            &.self {
              &::before {
                font-family: 'Font Awesome 6 Free';
                font-weight: 900;
                content: '\f2bd';
                background: $background-floating;
                color: $style-primary;
                position: absolute;
                width: 15px;
                height: 15px;
                line-height: 15px;
                font-size: 20px;
                text-align: center;
                margin-top: -2px;
                margin-left: 40px;
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
                width: 14px;
                height: 14px;
                font-size: 14px;
                text-align: center;
                margin-top: -2px;
                margin-left: 44px;
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
              width: 20px;
              height: 20px;
              line-height: 20px;
              font-size: 10px;
              text-align: center;
              margin-top: 42px;
              margin-left: -18px;
              border-radius: 50%;
            }

            .avatar {
              border-radius: 50%;
              overflow: hidden;
              width: 100%;
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

    onContext(event: MouseEvent, data: any) {
      this._context.open(event, data)
    }
  }
</script>
