<template>
  <div class="members">
    <table class="plugins" v-if="plugins">
      <tr>
        <td colspan="2" class="name">Plugins for {{ plugins.profile.name }}</td>
      </tr>
      <tr v-for="([key], i) in plugins.old" :key="key">
        <th>{{ key }}</th>
        <td><input type="text" v-model="plugins.old[i][1]" placeholder="value (JSON)" /></td>
      </tr>
      <tr v-for="([key], i) in plugins.new" :key="key">
        <th><input type="text" v-model="plugins.new[i][0]" placeholder="key (string)" /></th>
        <td><input type="text" v-model="plugins.new[i][1]" placeholder="value (JSON)" /></td>
      </tr>
      <tr>
        <td colspan="2" style="text-align: center">
          <button @click="$set(plugins, 'new', [...plugins.new, ['', '']])">+</button>
        </td>
      </tr>
      <tr>
        <td colspan="2">
          <button @click="savePlugins">save</button>
          <button @click="plugins = null">close</button>
        </td>
      </tr>
    </table>

    <p class="title">
      <span>Sessions</span>
    </p>

    <div
      class="member"
      :class="{
        'is-admin': neko.is_admin,
      }"
      v-for="(session, id) in sessions"
      :key="'session-' + id"
    >
      <div class="topbar">
        <div class="name">
          <i v-if="neko.is_admin" class="fa fa-trash-alt" @click="memberRemove(id)" title="remove" />
          {{ session.profile.name }}
        </div>
        <div class="controls">
          <i
            class="fa fa-shield-alt"
            :class="{
              'state-has': session.profile.is_admin,
            }"
            @click="neko.is_admin && updateProfile(id, { is_admin: !session.profile.is_admin })"
            title="is_admin"
          />
          <i
            class="fa fa-lock-open"
            :class="{
              'state-has': session.profile.can_login,
            }"
            @click="neko.is_admin && updateProfile(id, { can_login: !session.profile.can_login })"
            title="can_login"
          />
          <i
            class="fa fa-sign-in-alt"
            :class="{
              'state-has': session.profile.can_connect,
              'state-is': session.state.is_connected,
              'state-disabled': !session.profile.can_login,
            }"
            @click="neko.is_admin && updateProfile(id, { can_connect: !session.profile.can_connect })"
            title="can_connect"
          />
          <i
            class="fa fa-desktop"
            :class="{
              'state-has': session.profile.can_watch,
              'state-is': session.state.is_watching,
              'state-disabled': !session.profile.can_login || !session.profile.can_connect,
            }"
            @click="neko.is_admin && updateProfile(id, { can_watch: !session.profile.can_watch })"
            title="can_watch"
          />
          <i
            class="fa fa-keyboard"
            :class="{
              'state-has': session.profile.can_host,
              'state-is': neko.state.control.host_id == id,
              'state-disabled': !session.profile.can_login || !session.profile.can_connect,
            }"
            @click="neko.is_admin && updateProfile(id, { can_host: !session.profile.can_host })"
            title="can_host"
          />
          <i
            class="fa fa-microphone"
            :class="{
              'state-has': session.profile.can_share_media,
              'state-disabled': !session.profile.can_login || !session.profile.can_connect,
            }"
            @click="neko.is_admin && updateProfile(id, { can_share_media: !session.profile.can_share_media })"
            title="can_share_media"
          />
          <i
            class="fa fa-clipboard"
            :class="{
              'state-has': session.profile.can_access_clipboard,
              'state-disabled': !session.profile.can_login || !session.profile.can_connect,
            }"
            @click="neko.is_admin && updateProfile(id, { can_access_clipboard: !session.profile.can_access_clipboard })"
            title="can_access_clipboard"
          />
          <i
            class="fa fa-mouse"
            :class="{
              'state-has': session.profile.sends_inactive_cursor,
              'state-is':
                session.profile.sends_inactive_cursor &&
                neko.state.settings.inactive_cursors &&
                neko.state.cursors.some((e) => e.id == id),
              'state-disabled': !session.profile.can_login || !session.profile.can_connect,
            }"
            @click="
              neko.is_admin && updateProfile(id, { sends_inactive_cursor: !session.profile.sends_inactive_cursor })
            "
            title="sends_inactive_cursor"
          />
          <i
            class="fa fa-mouse-pointer"
            :class="{
              'state-has': session.profile.can_see_inactive_cursors,
              'state-disabled': !session.profile.can_login || !session.profile.can_connect,
            }"
            @click="
              neko.is_admin &&
                updateProfile(id, { can_see_inactive_cursors: !session.profile.can_see_inactive_cursors })
            "
            title="can_see_inactive_cursors"
          />
          <i class="fa fa-puzzle-piece state-has" @click="showPlugins(id, session.profile)" title="plugins" />
        </div>
      </div>
    </div>

    <p class="title">
      <span>Members</span>
      <button @click="membersLoad">reload</button>
    </p>

    <div
      class="member"
      :class="{
        'is-admin': neko.is_admin,
      }"
      v-for="member in membersWithoutSessions"
      :key="'member-' + member.id"
    >
      <div class="topbar">
        <div class="name">
          <i v-if="neko.is_admin" class="fa fa-trash-alt" @click="memberRemove(member.id)" title="remove" />
          {{ member.profile.name }}
        </div>
        <div class="controls">
          <i
            class="fa fa-shield-alt"
            :class="{
              'state-has': member.profile.is_admin,
            }"
            @click="neko.is_admin && updateProfile(member.id, { is_admin: !member.profile.is_admin })"
            title="is_admin"
          />
          <i
            class="fa fa-lock-open"
            :class="{
              'state-has': member.profile.can_login,
            }"
            @click="neko.is_admin && updateProfile(member.id, { can_login: !member.profile.can_login })"
            title="can_login"
          />
          <i
            class="fa fa-sign-in-alt"
            :class="{
              'state-has': member.profile.can_connect,
              'state-disabled': !member.profile.can_login,
            }"
            @click="neko.is_admin && updateProfile(member.id, { can_connect: !member.profile.can_connect })"
            title="can_connect"
          />
          <i
            class="fa fa-desktop"
            :class="{
              'state-has': member.profile.can_watch,
              'state-disabled': !member.profile.can_login || !member.profile.can_connect,
            }"
            @click="neko.is_admin && updateProfile(member.id, { can_watch: !member.profile.can_watch })"
            title="can_watch"
          />
          <i
            class="fa fa-keyboard"
            :class="{
              'state-has': member.profile.can_host,
              'state-disabled': !member.profile.can_login || !member.profile.can_connect,
            }"
            @click="neko.is_admin && updateProfile(member.id, { can_host: !member.profile.can_host })"
            title="can_host"
          />
          <i
            class="fa fa-microphone"
            :class="{
              'state-has': member.profile.can_share_media,
              'state-disabled': !member.profile.can_login || !member.profile.can_connect,
            }"
            @click="neko.is_admin && updateProfile(member.id, { can_share_media: !member.profile.can_share_media })"
            title="can_share_media"
          />
          <i
            class="fa fa-clipboard"
            :class="{
              'state-has': member.profile.can_access_clipboard,
              'state-disabled': !member.profile.can_login || !member.profile.can_connect,
            }"
            @click="
              neko.is_admin && updateProfile(member.id, { can_access_clipboard: !member.profile.can_access_clipboard })
            "
            title="can_access_clipboard"
          />
          <i
            class="fa fa-mouse"
            :class="{
              'state-has': member.profile.sends_inactive_cursor,
              'state-disabled': !member.profile.can_login || !member.profile.can_connect,
            }"
            @click="
              neko.is_admin &&
                updateProfile(member.id, { sends_inactive_cursor: !member.profile.sends_inactive_cursor })
            "
            title="sends_inactive_cursor"
          />
          <i
            class="fa fa-mouse-pointer"
            :class="{
              'state-has': member.profile.can_see_inactive_cursors,
              'state-disabled': !member.profile.can_login || !member.profile.can_connect,
            }"
            @click="
              neko.is_admin &&
                updateProfile(member.id, { can_see_inactive_cursors: !member.profile.can_see_inactive_cursors })
            "
            title="can_see_inactive_cursors"
          />
          <i class="fa fa-puzzle-piece state-has" @click="showPlugins(member.id, member.profile)" title="plugins" />
        </div>
      </div>
    </div>

    <table class="new-member" v-if="neko.is_admin">
      <tr>
        <td colspan="2" class="name">New Member</td>
      </tr>
      <tr>
        <th>username</th>
        <td><input type="text" v-model="newUsername" /></td>
      </tr>
      <tr>
        <th>password</th>
        <td><input type="text" v-model="newPassword" /></td>
      </tr>
      <tr>
        <td colspan="2" class="name" style="text-align: center">Profile</td>
      </tr>
      <tr>
        <th>name</th>
        <td><input type="text" v-model="newProfile.name" /></td>
      </tr>
      <tr>
        <th>is_admin</th>
        <td><input type="checkbox" v-model="newProfile.is_admin" /></td>
      </tr>
      <tr>
        <th>can_login</th>
        <td><input type="checkbox" v-model="newProfile.can_login" /></td>
      </tr>
      <tr>
        <th>can_connect</th>
        <td><input type="checkbox" v-model="newProfile.can_connect" /></td>
      </tr>
      <tr>
        <th>can_watch</th>
        <td><input type="checkbox" v-model="newProfile.can_watch" /></td>
      </tr>
      <tr>
        <th>can_host</th>
        <td><input type="checkbox" v-model="newProfile.can_host" /></td>
      </tr>
      <tr>
        <th>can_share_media</th>
        <td><input type="checkbox" v-model="newProfile.can_share_media" /></td>
      </tr>
      <tr>
        <th>can_access_clipboard</th>
        <td><input type="checkbox" v-model="newProfile.can_access_clipboard" /></td>
      </tr>
      <tr>
        <th>sends_inactive_cursor</th>
        <td><input type="checkbox" v-model="newProfile.sends_inactive_cursor" /></td>
      </tr>
      <tr>
        <th>can_see_inactive_cursors</th>
        <td><input type="checkbox" v-model="newProfile.can_see_inactive_cursors" /></td>
      </tr>
      <tr>
        <td colspan="2"><button @click="memberCreate">create</button></td>
      </tr>
    </table>
  </div>
</template>

<style lang="scss" scoped>
  @import '@/page/assets/styles/main.scss';

  .title {
    padding: 4px;
    font-weight: bold;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .members {
    display: block;
    width: 100%;
    overflow: hidden;

    .member {
      padding: 5px;
      margin: 5px 0;
      border: 1px solid white;
      box-sizing: border-box;

      &.is-admin .fa {
        cursor: pointer;
      }

      .topbar {
        display: flex;
        align-items: center;

        .name {
          flex: 1 1;
        }
      }

      .fa {
        padding: 5px;
        color: rgb(211, 47, 47);

        &.state-has {
          color: #fff;
        }

        &.state-is {
          color: green;
        }

        &.state-disabled {
          color: #555;
        }
      }
    }

    .new-member,
    .plugins {
      width: 100%;
      margin: 5px 0;

      .name {
        font-weight: bold;
      }

      td,
      th {
        border: 1px solid #ccc;
        padding: 4px;
        width: 50%;
      }

      th {
        text-align: right;
      }

      input[type='text'] {
        width: 100%;
        box-sizing: border-box;
      }
    }

    .plugins {
      position: absolute;
      width: auto;
      box-shadow: 0px 0px 10px 5px black;
      background: $background-tertiary;

      textarea,
      input {
        width: 100%;
        box-sizing: border-box;
      }
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Prop } from 'vue-property-decorator'
  import Neko, { ApiModels, StateModels } from '~/component/main.vue'

  @Component({
    name: 'neko-members',
  })
  export default class extends Vue {
    @Prop() readonly neko!: Neko

    constructor() {
      super()

      // init
      this.newProfile = Object.assign({}, this.defProfile)
    }

    get sessions(): Record<string, StateModels.Session> {
      return this.neko.state.sessions
    }

    get membersWithoutSessions(): ApiModels.MemberData[] {
      return this.members.filter(({ id }) => id && !(id in this.sessions))
    }

    members: ApiModels.MemberData[] = []
    plugins: {
      id: string
      old: Array<Array<string>>
      new: Array<Array<string>>
      profile: ApiModels.MemberProfile
    } | null = null

    newUsername: string = ''
    newPassword: string = ''
    newProfile: ApiModels.MemberProfile = {}
    defProfile: ApiModels.MemberProfile = {
      name: '',
      is_admin: false,
      can_login: true,
      can_connect: true,
      can_watch: true,
      can_host: true,
      can_share_media: true,
      can_access_clipboard: true,
      sends_inactive_cursor: true,
      can_see_inactive_cursors: true,
    }

    async memberCreate() {
      try {
        const res = await this.neko.members.membersCreate({
          username: this.newUsername,
          password: this.newPassword,
          profile: this.newProfile,
        })

        if (res.data) {
          Vue.set(this, 'members', [...this.members, res.data])
        }

        // clear
        Vue.set(this, 'newUsername', '')
        Vue.set(this, 'newPassword', '')
        Vue.set(this, 'newProfile', Object.assign({}, this.defProfile))
      } catch (e: any) {
        alert(e.response ? e.response.data.message : e)
      }
    }

    async membersLoad(limit: number = 0) {
      const offset = 0

      try {
        const res = await this.neko.members.membersList(limit, offset)
        Vue.set(this, 'members', res.data)
      } catch (e: any) {
        alert(e.response ? e.response.data.message : e)
      }
    }

    async memberGetProfile(memberId: string): Promise<ApiModels.MemberProfile | undefined> {
      try {
        const res = await this.neko.members.membersGetProfile(memberId)
        return res.data
      } catch (e: any) {
        alert(e.response ? e.response.data.message : e)
      }
    }

    async updateProfile(memberId: string, memberProfile: ApiModels.MemberProfile) {
      try {
        await this.neko.members.membersUpdateProfile(memberId, memberProfile)
        const members = this.members.map((member) => {
          if (member.id == memberId) {
            return {
              id: memberId,
              profile: { ...member.profile, ...memberProfile },
            }
          } else {
            return member
          }
        })
        Vue.set(this, 'members', members)
      } catch (e: any) {
        alert(e.response ? e.response.data.message : e)
      }
    }

    async updatePassword(memberId: string, password: string) {
      try {
        await this.neko.members.membersUpdatePassword(memberId, { password })
      } catch (e: any) {
        alert(e.response ? e.response.data.message : e)
      }
    }

    async memberRemove(memberId: string) {
      try {
        await this.neko.members.membersRemove(memberId)
        const members = this.members.filter(({ id }) => id != memberId)
        Vue.set(this, 'members', members)
      } catch (e: any) {
        alert(e.response ? e.response.data.message : e)
      }
    }

    showPlugins(id: string, profile: ApiModels.MemberProfile) {
      const old = Object.entries(profile.plugins || {}).map(([key, val]) => [key, JSON.stringify(val, null, 2)])

      this.plugins = {
        id,
        old,
        new: old.length > 0 ? [] : [['', '']],
        profile,
      }
    }

    savePlugins() {
      if (!this.plugins) return

      let errKey = ''
      try {
        let plugins = {} as any
        for (let [key, val] of this.plugins.old) {
          errKey = key
          plugins[key] = JSON.parse(val)
        }
        for (let [key, val] of this.plugins.new) {
          errKey = key
          plugins[key] = JSON.parse(val)
        }

        this.updateProfile(this.plugins.id, { plugins })
        this.plugins = null
      } catch (e: any) {
        alert(errKey + ': ' + e)
      }
    }

    mounted() {
      this.membersLoad(10)
    }
  }
</script>
