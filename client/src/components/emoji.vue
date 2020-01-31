<template>
  <div class="neko-emoji">
    <div class="search">
      <input type="text" ref="search" v-model="search" />
    </div>
    <div class="list" ref="scroll" @scroll="onScroll">
      <ul :class="[search === '' ? 'group-list' : 'emoji-list']">
        <template v-if="search === ''">
          <li v-for="(group, index) in groups" :key="index" class="group" ref="groups">
            <span class="label">{{ group.name }}</span>
            <ul class="emoji-list">
              <li
                v-for="emoji in group.list"
                :key="`${group.id}-${emoji}`"
                :class="['emoji', hovered === emoji ? 'active' : '']"
              >
                <span
                  :class="['emoji-20', `e-${emoji}`]"
                  @mouseenter.stop.prevent="onMouseEnter($event, emoji)"
                  @click.stop.prevent="onClick($event, emoji)"
                ></span>
              </li>
            </ul>
          </li>
        </template>
        <template v-else>
          <li v-for="emoji in filtered" :key="emoji" :class="['emoji', hovered === emoji ? 'active' : '']">
            <span
              :class="['emoji-20', `e-${emoji}`]"
              @mouseenter.stop.prevent="onMouseEnter($event, emoji)"
              @click.stop.prevent="onClick($event, emoji)"
            ></span>
          </li>
        </template>
      </ul>
    </div>
    <div class="details">
      <template v-if="hovered !== ''">
        <span :class="['icon', 'emoji-20', `e-${hovered}`]"></span><span class="emoji">:{{ hovered }}:</span>
      </template>
    </div>
    <div class="groups">
      <ul>
        <li
          v-for="(group, index) in groups"
          :key="index"
          :class="[group.id, active === group.id && search === '' ? 'active' : '']"
          @click.stop.prevent="scrollTo($event, index)"
        >
          <i class="fas fa-angry"></i>
        </li>
      </ul>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  $emoji-width: 350px;

  .neko-emoji {
    position: absolute;
    width: $emoji-width;
    height: 200px;
    background: $background-secondary;
    bottom: 30px;
    right: 0;
    display: flex;
    flex-direction: column;
    border-radius: 5px;
    overflow: hidden;

    .search {
      flex-shrink: 0;
      height: 38px;
      border-bottom: 1px solid $background-tertiary;

      input {
      }
    }

    .list {
      position: relative;
      flex-grow: 1;
      overflow-y: scroll;
      overflow-x: hidden;
      scrollbar-width: thin;
      scrollbar-color: $background-tertiary transparent;
      scroll-behavior: smooth;
      padding: 5px;

      &::-webkit-scrollbar {
        width: 4px;
      }

      &::-webkit-scrollbar-track {
        background-color: transparent;
      }

      &::-webkit-scrollbar-thumb {
        background-color: $background-tertiary;
        border-radius: 4px;
      }

      &::-webkit-scrollbar-thumb:hover {
        background-color: $background-floating;
      }

      .group-list {
        width: $emoji-width;
        display: flex;
        flex-direction: column;

        li {
          &.group {
            .label {
              z-index: 2;
              text-transform: uppercase;
              font-weight: 500;
              font-size: 12px;
              position: sticky;
              top: 0;
            }

            .emoji-list {
              margin: 10px 0;
            }
          }
        }
      }

      .emoji-list {
        display: flex;
        flex-direction: row;
        flex-wrap: wrap;
        li {
          &.emoji {
            padding: 2px;
            border-radius: 3px;

            &.active {
              background-color: #fff;
            }
          }
        }
      }
    }

    .details {
      flex-shrink: 0;
      height: 36px;
      background: $background-tertiary;

      span {
        &.icon {
        }

        &.emoji {
        }
      }
    }

    .groups {
      flex-shrink: 0;
      height: 28px;
      background: $background-floating;

      ul {
        display: flex;
        flex-direction: row;
        flex-wrap: wrap;

        li {
          flex-grow: 1;
          height: 23px;

          &.active {
            border-bottom: 2px solid $style-primary;
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'

  import { list } from './emoji/list'
  import { keywords } from './emoji/keywords'
  import { groups } from './emoji/groups'

  @Component({
    name: 'neko-emoji',
  })
  export default class extends Vue {
    @Ref('scroll') readonly _scroll!: HTMLElement
    @Ref('search') readonly _search!: HTMLInputElement
    @Ref('groups') readonly _groups!: HTMLElement[]

    waitingForPaint = false
    search = ''
    active = groups[0].id
    hovered = ''

    get groups() {
      return groups
    }

    get filtered() {
      const filtered = []
      for (const emoji of list) {
        if (
          emoji.includes(this.search) || typeof keywords[emoji] !== 'undefined'
            ? keywords[emoji].some(keyword => keyword.includes(this.search))
            : false
        ) {
          filtered.push(emoji)
        }
      }
      return filtered
    }

    scrollTo(event: MouseEvent, index: number) {
      const ele = this._groups[index]
      if (!ele) {
        return
      }

      let top = ele.offsetTop
      if (index == 0) {
        top = 0
      }
      this._scroll.scrollTop = top
    }

    onScroll() {
      if (!this.waitingForPaint) {
        this.waitingForPaint = true
        window.requestAnimationFrame(this.onScrollPaint.bind(this))
      }
    }

    onScrollPaint() {
      this.waitingForPaint = false
      let scrollTop = this._scroll.scrollTop
      let active = this.groups[0]
      for (let i = 0, l = this.groups.length; i < l; i++) {
        let group = this.groups[i]
        let component = this._groups[i]
        if (component && component.offsetTop > scrollTop) {
          break
        }
        active = group
      }
      this.active = active.id
    }

    onMouseExit(event: MouseEvent, emoji: string) {
      this.hovered = ''
    }

    onMouseEnter(event: MouseEvent, emoji: string) {
      this.hovered = emoji
      this._search.placeholder = `:${emoji}:`
    }

    onClick(event: MouseEvent, emoji: string) {
      this.$emit('done', emoji)
    }
  }
</script>
