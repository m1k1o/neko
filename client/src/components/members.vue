<template>
  <div class="members">
    <div class="members-container">
      <ul class="members-list">
        <li v-if="member">
          <div :class="[{ host: member.id === host }, 'self', 'member']">
            <img :src="`https://api.adorable.io/avatars/50/${member.username}.png`" />
          </div>
        </li>
        <template v-for="(member, index) in members">
          <li
            v-if="member.id !== id && member.connected"
            :key="index"
            v-tooltip="{ content: member.username, placement: 'top', offset: 5, boundariesElement: 'body' }"
          >
            <div :class="[{ host: member.id === host, admin: member.admin }, 'member']">
              <img
                :src="`https://api.adorable.io/avatars/50/${member.username}.png`"
                @contextmenu="context($event, { member, index })"
              />
            </div>
          </li>
        </template>
      </ul>
    </div>
    <vue-context class="context" ref="menu">
      <template slot-scope="child" v-if="child.data && admin">
        <li>
          <strong>{{ child.data.member.username }}</strong>
        </li>
        <li class="seperator" />
        <li>
          <span @click="mute(child.data.member)" v-if="!child.data.member.muted">Mute</span>
          <span @click="unmute(child.data.member)" v-else>Unmute</span>
        </li>
        <template v-if="child.data.member.id === host">
          <li>
            <span @click="release(child.data.member)">Release Controls</span>
          </li>
          <li>
            <span @click="control(child.data.member)">Take Controls</span>
          </li>
        </template>
        <li class="seperator" />
        <li>
          <span @click="kick(child.data.member)" style="color: #f04747">Kick</span>
        </li>
        <li>
          <span @click="ban(child.data.member)" style="color: #f04747">Ban</span>
        </li>
      </template>
    </vue-context>
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
                font-family: 'Font Awesome 5 Free';
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
                font-family: 'Font Awesome 5 Free';
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
              font-family: 'Font Awesome 5 Free';
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

            img {
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

    .context {
      background-color: $background-floating;
      background-clip: padding-box;
      border-radius: 0.25rem;
      display: block;
      margin: 0;
      padding: 5px;
      min-width: 150px;
      z-index: 1500;
      position: fixed;
      list-style: none;
      box-sizing: border-box;
      max-height: calc(100% - 50px);
      overflow-y: auto;
      color: $interactive-normal;
      user-select: none;
      box-shadow: $elevation-high;

      > li {
        margin: 0;
        position: relative;

        &.seperator {
          height: 1px;
          background: $background-secondary;
          margin: 3px 0;
        }

        > strong {
          display: block;
          padding: 8px 5px;
          font-weight: 700;
        }

        > span {
          cursor: pointer;
          display: block;
          padding: 5px;
          font-weight: 400;
          text-decoration: none;
          white-space: nowrap;
          background-color: transparent;
          border-radius: 3px;

          &:hover,
          &:focus {
            text-decoration: none;
            background-color: $background-modifier-hover;
            color: $interactive-hover;
          }

          &:focus {
            outline: 0;
          }
        }
      }

      &:focus {
        outline: 0;
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'
  import { Member } from '~/client/types'

  // @ts-ignore
  import { VueContext } from 'vue-context'

  @Component({
    name: 'neko-members',
    components: {
      'vue-context': VueContext,
    },
  })
  export default class extends Vue {
    @Ref('menu') readonly menu!: any

    get id() {
      return this.$accessor.user.id
    }

    get admin() {
      return this.$accessor.user.admin
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

    context(event: MouseEvent, data: any) {
      if (this.admin) {
        event.preventDefault()
        this.menu.open(event, data)
      }
    }

    kick(member: Member) {
      this.$swal({
        title: `Kick ${member.username}?`,
        text: `Are you sure you want to kick ${member.username}?`,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Yes',
      }).then(({ value }) => {
        if (value) {
          this.$accessor.user.kick(member)
        }
      })
    }

    ban(member: Member) {
      this.$swal({
        title: `Ban ${member.username}?`,
        text: `Are you sure you want to ban ${member.username}? You will need to restart the server to undo this.`,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Yes',
      }).then(({ value }) => {
        if (value) {
          this.$accessor.user.ban(member)
        }
      })
    }

    mute(member: Member) {
      this.$swal({
        title: `Mute ${member.username}?`,
        text: `Are you sure you want to mute ${member.username}?`,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Yes',
      }).then(({ value }) => {
        if (value) {
          this.$accessor.user.mute(member)
        }
      })
    }

    unmute(member: Member) {
      this.$swal({
        title: `Unmute ${member.username}?`,
        text: `Are you sure you want to unmute ${member.username}?`,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Yes',
      }).then(({ value }) => {
        if (value) {
          this.$accessor.user.unmute(member)
        }
      })
    }

    release(member: Member) {
      this.$accessor.remote.adminRelease()
    }

    control(member: Member) {
      this.$accessor.remote.adminControl()
    }
  }
</script>
