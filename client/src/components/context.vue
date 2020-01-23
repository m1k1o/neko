<template>
  <vue-context class="context" ref="context">
    <template slot-scope="child" v-if="child.data">
      <li class="header">
        <div class="user">
          <img :src="`https://api.adorable.io/avatars/25/${child.data.member.username}.png`" />
          <strong>{{ child.data.member.username }}</strong>
        </div>
      </li>
      <li class="seperator" />
      <li>
        <span @click="ignore(child.data.member)" v-if="!child.data.member.ignored">Ignore</span>
        <span @click="unignore(child.data.member)" v-else>Unignore</span>
      </li>

      <template v-if="admin">
        <li>
          <span @click="mute(child.data.member)" v-if="!child.data.member.muted">Mute</span>
          <span @click="unmute(child.data.member)" v-else>Unmute</span>
        </li>
        <li v-if="child.data.member.id === host">
          <span @click="adminRelease(child.data.member)">Force Release Controls</span>
        </li>
        <li v-if="child.data.member.id === host">
          <span @click="adminControl(child.data.member)">Force Take Controls</span>
        </li>
        <li>
          <span v-if="child.data.member.id !== host" @click="adminGive(child.data.member)">Give Controls</span>
        </li>
      </template>
      <template v-else>
        <li v-if="hosting">
          <span @click="give(child.data.member)">Give Controls</span>
        </li>
      </template>

      <template v-if="admin">
        <li class="seperator" />
        <li>
          <span @click="kick(child.data.member)" style="color: #f04747">Kick</span>
        </li>
        <li>
          <span @click="ban(child.data.member)" style="color: #f04747">Ban IP</span>
        </li>
      </template>
    </template>
  </vue-context>
</template>

<style lang="scss" scoped>
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
      align-content: center;

      &.header {
        .user {
          display: flex;
          flex-direction: row;
          align-content: center;
          padding: 5px 0;

          img {
            width: 25px;
            height: 25px;
            border-radius: 50%;
            margin-right: 5px;
          }

          strong {
            line-height: 25px;
            font-weight: 700;
            max-width: 200px;
            text-overflow: ellipsis;
          }
        }
      }

      &.seperator {
        height: 1px;
        background: $background-secondary;
        margin: 3px 0;
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
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'
  import { Member } from '~/neko/types'

  // @ts-ignore
  import { VueContext } from 'vue-context'

  @Component({
    name: 'neko-context',
    components: {
      'vue-context': VueContext,
    },
  })
  export default class extends Vue {
    @Ref('context') readonly context!: any

    get admin() {
      return this.$accessor.user.admin
    }

    get hosting() {
      return this.$accessor.remote.hosting
    }

    get host() {
      return this.$accessor.remote.id
    }

    open(event: MouseEvent, data: any) {
      this.context.open(event, data)
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

    adminRelease(member: Member) {
      this.$accessor.remote.adminRelease()
    }

    adminControl(member: Member) {
      this.$accessor.remote.adminControl()
    }

    adminGive(member: Member) {
      this.$accessor.remote.adminGive(member)
    }

    give(member: Member) {
      this.$accessor.remote.give(member)
    }

    ignore(member: Member) {
      this.$accessor.user.setIgnored({ id: member.id, ignored: true })
    }

    unignore(member: Member) {
      this.$accessor.user.setIgnored({ id: member.id, ignored: false })
    }
  }
</script>
